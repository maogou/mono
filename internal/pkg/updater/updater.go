package updater

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

const (
	defaultCheckInterval   = 300     // 检查周期缺省值(秒)
	defaultDownloadTimeout = 600     // 下载超时缺省值(秒)
	defaultMaxSizeMB       = 128     // 下载包大小上限缺省值(MB)
	manifestMaxBytes       = 1 << 20 // 清单 JSON 体积上限(1MiB)
)

// EnvVersionOverride 覆盖"当前版本号",由升级流程在重启前注入,
// 防止配置版本滞后导致重复升级(见 effectiveVersion)。
const EnvVersionOverride = "UPDATE_CURRENT_VERSION"

// Options 是升级器运行参数,由调用方(如 internal/command/run.go)从业务配置装配。
// updater 为自包含组件,不 import internal/config。
type Options struct {
	CheckIntervalSeconds   int    // 检查周期;<=0 时缺省 300s
	FirstCheckDelaySeconds int    // 启动后首轮延迟;0=立即检查
	CurrentVersion         string // 当前版本号(通常为编译期内置常量,由调用方注入;env 可覆盖)
	CheckURL               string // 更新源接口地址,GET 返回 Manifest JSON
	DownloadTimeoutSeconds int    // 单次 HTTP 请求超时;<=0 时缺省 30s
	MaxSizeMB              int64  // 下载包大小上限;<=0 时缺省 128

	// Restarter 升级提交后的重启执行组件(可选):缺省 nil 时由 New 自动创建。
	// 传入时 New 复用该实例,使 apply 提交(Commit)与调用方收尾后的重启(Exec)
	// 作用于同一对象,替代包级全局变量;测试可注入 spy 或直接传 nil 用默认。
	Restarter *Restarter
}

// Updater 实现单机自升级:检查 → 下载 → 校验 → 原子替换 → 通知重启。
// 仅由 Run 的单协程驱动,单轮内用 inflight CAS 防并发重入。
type Updater struct {
	cfg Options
	log *zap.Logger
	hc  *http.Client

	inflight atomic.Bool

	// restarter 记录 apply 提交的正式路径,供编排方在收尾后 Exec 重启同一实例
	// (见 New:Options.Restarter 缺省自动创建)。
	restarter *Restarter

	// targetPath 仅供同包测试注入:非空时校验/替换该文件,而不是当前进程可执行文件。
	targetPath string
}

// New 构造升级器并套用缺省参数。
func New(o Options, log *zap.Logger) *Updater {
	if log == nil {
		log = zap.NewNop()
	}
	if o.CheckIntervalSeconds <= 0 {
		o.CheckIntervalSeconds = defaultCheckInterval
	}
	if o.DownloadTimeoutSeconds <= 0 {
		o.DownloadTimeoutSeconds = defaultDownloadTimeout
	}
	if o.MaxSizeMB <= 0 {
		o.MaxSizeMB = defaultMaxSizeMB
	}
	if o.Restarter == nil {
		o.Restarter = NewRestarter() // 未注入时自动创建,apply 仍能记录提交路径
	}
	return &Updater{
		cfg:       o,
		log:       log,
		hc:        &http.Client{Timeout: time.Duration(o.DownloadTimeoutSeconds) * time.Second},
		restarter: o.Restarter,
	}
}

// Run 启动自升级循环,收到 ctx 取消即退出,可配合优雅停机使用。
// 就地替换(minio/selfupdate)与重启(Linux/macOS execve 同 PID 换壳、Windows
// 按 minio 语义拉起新实例,见 exec_*.go)三平台均支持,无平台门控。
func (u *Updater) Run(ctx context.Context, upgraded chan<- struct{}) {
	if u.cfg.CheckURL == "" {
		u.log.Warn("updater: check_url 未配置,自升级已关闭")
		return
	}
	u.runLoop(ctx, upgraded)
}

// runLoop 为可测试的循环本体。
// 每轮:等 delay(首轮为 FirstCheckDelay,之后为周期+抖动)→ 执行单轮升级。
func (u *Updater) runLoop(ctx context.Context, upgraded chan<- struct{}) {
	interval := time.Duration(u.cfg.CheckIntervalSeconds) * time.Second
	delay := time.Duration(u.cfg.FirstCheckDelaySeconds) * time.Second
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
		if err := u.CheckAndUpgrade(ctx, upgraded); err != nil {
			u.logUpgradeError(err)
		}
		delay = nextDelay(interval)
	}
}

// nextDelay 在周期基础上加 0~1/3 的随机抖动,错开多实例同时请求更新源。
// 抖动来源 crypto/rand(非安全用途,但避免引入弱随机告警)。
func nextDelay(base time.Duration) time.Duration {
	if base <= 0 {
		base = defaultCheckInterval * time.Second
	}
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return base // 读随机数失败时放弃抖动,退化为固定周期
	}
	maxJitter := uint64(base) / 3
	if maxJitter == 0 {
		return base // 周期过小,无抖动空间
	}
	// #nosec G115 -- 取模结果必小于 maxJitter,而 maxJitter ≤ base/3 ≪ int64 上限,转换无溢出可能
	return base + time.Duration(binary.LittleEndian.Uint64(b[:])%maxJitter)
}

// logUpgradeError 按错误类别决定告警级别与措辞。
func (u *Updater) logUpgradeError(err error) {
	switch {
	case errors.Is(err, ErrPolicy):
		u.log.Warn("updater: 配置或策略问题,升级本轮跳过", zap.Error(err))
	case errors.Is(err, ErrVerify), errors.Is(err, ErrCommit):
		u.log.Error("updater: 升级失败,保留当前版本继续运行", zap.Error(err))
	default:
		u.log.Warn("updater: 检查升级失败,下轮重试", zap.Error(err))
	}
}

// CheckAndUpgrade 执行单轮升级并返回分类错误;成功提交后向 upgraded 发一次通知。
// 返回的错误仅表示"本轮未完成",不影响进程继续运行。
func (u *Updater) CheckAndUpgrade(ctx context.Context, upgraded chan<- struct{}) error {
	if !u.inflight.CompareAndSwap(false, true) {
		return nil // 上一轮仍在进行
	}
	defer u.inflight.Store(false)

	m, err := u.fetchManifest(ctx)
	if err != nil {
		return err
	}

	// 版本号一致(含重启后 env 覆盖生效)则跳过
	if m.Version == u.effectiveVersion() {
		u.log.Debug("updater: 已是最新版本,跳过", zap.String("version", m.Version))
		return nil
	}

	// 防空转:manifest md5 == 当前可执行文件 md5,说明内容已就位但配置版本号滞后,
	// 直接视为最新,避免"替换→重启→再替换"死循环。
	if same, err := u.exeMatches(m); err != nil {
		u.log.Debug("updater: 读取当前可执行文件失败,跳过防空转检查", zap.Error(err))
	} else if same {
		u.log.Info(
			"updater: 当前可执行文件内容已为该版本(内置版本号滞后?),跳过升级",
			zap.String("version", m.Version),
		)
		return nil
	}

	payload, err := u.download(ctx, m.DownloadURL)
	if err != nil {
		return err
	}
	if err := u.verify(m, payload); err != nil {
		return err
	}
	if err := u.apply(m, payload); err != nil {
		return err
	}

	// 注入版本覆盖,重启后首轮即可命中"版本一致"分支,防止配置漂移导致重复升级
	if err := os.Setenv(EnvVersionOverride, m.Version); err != nil {
		u.log.Warn("updater: 注入版本覆盖失败,重启后可能重复检查", zap.Error(err))
	}
	u.log.Info("updater: 升级包已替换就绪,即将重启", m.logFields()...)

	select {
	case upgraded <- struct{}{}: // 通知编排层发起优雅停服+重启
	default: // 通道无接收方或已满,不阻塞
	}
	return nil
}

// fetchManifest 请求更新源接口并解析清单(体积上限 1MiB)。
func (u *Updater) fetchManifest(ctx context.Context) (*Manifest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.cfg.CheckURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: 构造清单请求失败: %v", ErrPolicy, err)
	}
	resp, err := u.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: 请求更新源失败: %v", ErrNetwork, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: 更新源返回 %s", ErrNetwork, resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, manifestMaxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: 读取清单失败: %v", ErrNetwork, err)
	}
	if len(body) > manifestMaxBytes {
		return nil, fmt.Errorf("%w: 清单体积超出上限 %d 字节", ErrPolicy, manifestMaxBytes)
	}
	return ParseManifest(body)
}

// download 下载新二进制;按 MaxSizeMB 限流,防恶意/异常超大包打爆内存。
func (u *Updater) download(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: 构造下载请求失败: %v", ErrPolicy, err)
	}
	resp, err := u.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: 下载失败: %v", ErrNetwork, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: 下载返回 %s", ErrNetwork, resp.Status)
	}

	limit := u.cfg.MaxSizeMB << 20 // New 已保证 MaxSizeMB > 0
	payload, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("%w: 读取下载内容失败: %v", ErrNetwork, err)
	}
	if int64(len(payload)) > limit {
		return nil, fmt.Errorf("%w: 下载包 %d 字节超出大小上限 %d 字节", ErrPolicy, len(payload), limit)
	}
	return payload, nil
}

// exeMatches 判断当前可执行文件内容是否已等于清单中的新版本(防空转,md5 比较)。
func (u *Updater) exeMatches(m *Manifest) (bool, error) {
	path, err := u.exePath()
	if err != nil {
		return false, err
	}
	got, err := md5File(path)
	if err != nil {
		return false, err
	}
	return bytes.Equal(got, m.decodedMD5()), nil
}

// exePath 返回目标可执行文件路径:测试注入 targetPath 时优先,否则取当前进程路径。
// Linux 上运行中的文件被替换后 /proc/self/exe 会带 " (deleted)" 后缀,需归一化。
func (u *Updater) exePath() (string, error) {
	if u.targetPath != "" {
		return u.targetPath, nil
	}
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(exe, " (deleted)"), nil
}

// md5File 流式计算文件的 md5(防空转比对用,整文件比大小上限大得多,须流式)。
// #nosec G304 -- path 仅来自 os.Executable() 或测试注入的 targetPath,无用户输入参与。
func md5File(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	// #nosec G401 -- md5 为服务端既有契约指定的完整性比对(非认证),防空转仅需同摘要
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// effectiveVersion 返回版本比对基准:env(重启后注入)优先于内置版本号。
func (u *Updater) effectiveVersion() string {
	if v := os.Getenv(EnvVersionOverride); v != "" {
		return v
	}
	return u.cfg.CurrentVersion
}
