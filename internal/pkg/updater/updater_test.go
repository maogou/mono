package updater

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// fakeUpdateSource 模拟服务端动态接口:/manifest 返回清单,/bin 返回新二进制。
// binHits 统计下载次数,供断言"防空转/并发只下载一次"。
type fakeUpdateSource struct {
	srv          *httptest.Server
	m            *Manifest
	binHits      atomic.Int64
	manifestHits atomic.Int64
	payload      []byte
	// gate 非 nil 时,/bin 与 /manifest 先在此等待(构造并发场景)
	gate chan struct{}
}

func newFakeSource(t *testing.T, payload []byte, version string) *fakeUpdateSource {
	t.Helper()
	f := &fakeUpdateSource{
		m: &Manifest{
			Version: version,
			MD5:     md5HexOf(payload),
		},
		payload: payload,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/manifest", func(w http.ResponseWriter, _ *http.Request) {
		f.manifestHits.Add(1)
		if f.gate != nil {
			<-f.gate
		}
		_ = json.NewEncoder(w).Encode(f.m)
	})
	mux.HandleFunc("/bin", func(w http.ResponseWriter, _ *http.Request) {
		f.binHits.Add(1)
		if f.gate != nil {
			<-f.gate
		}
		_, _ = w.Write(f.payload)
	})
	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	f.m.DownloadURL = f.srv.URL + "/bin"
	return f
}

// newFullUpdater 造一个 targetPath 指向临时"现役二进制"的升级器。
func newFullUpdater(t *testing.T, opts Options) (*Updater, string) {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "app")
	require.NoError(t, os.WriteFile(bin, []byte("OLD-BINARY"), 0o755))

	t.Setenv(EnvVersionOverride, "")
	// 走生产同款注入路径:实例经 Options 传入,apply 的 Commit 与编排方 Exec 同一对象
	opts.Restarter = NewRestarter()
	u := New(opts, nil)
	u.targetPath = bin
	return u, bin
}

func drainNotify(t *testing.T, ch <-chan struct{}) bool {
	t.Helper()
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

func TestCheckAndUpgradeHappyPath(t *testing.T) {
	payload := []byte("#!/bin/sh\necho NEW-VERSION\n")
	src := newFakeSource(t, payload, "v2.0.0")

	u, bin := newFullUpdater(t, Options{
		CheckURL:       src.srv.URL + "/manifest",
		CurrentVersion: "v1.0.0",
	})
	upgraded := make(chan struct{}, 1)

	err := u.CheckAndUpgrade(context.Background(), upgraded)
	require.NoError(t, err)

	require.True(t, drainNotify(t, upgraded), "升级成功应发出重启通知")
	got, rerr := os.ReadFile(bin)
	require.NoError(t, rerr)
	require.Equal(t, payload, got, "现役二进制应被替换为新内容")
	require.Equal(t, int64(1), src.binHits.Load())
	require.Equal(t, "v2.0.0", os.Getenv(EnvVersionOverride), "应注入版本覆盖防止重复升级")
	require.Equal(t, bin, u.restarter.targetPath,
		"提交成功应把正式路径 Commit 进 Restarter(替换后 /proc/self/exe 已指向被删旧 inode,exec 必须回正式路径)")
	t.Cleanup(func() { _ = os.Unsetenv(EnvVersionOverride) })
}

func TestCheckAndUpgradeSameVersionSkips(t *testing.T) {
	payload := []byte("new-binary")
	src := newFakeSource(t, payload, "v1.0.0")

	u, _ := newFullUpdater(t, Options{
		CheckURL:       src.srv.URL + "/manifest",
		CurrentVersion: "v1.0.0", // 与远端一致
	})
	upgraded := make(chan struct{}, 1)

	err := u.CheckAndUpgrade(context.Background(), upgraded)
	require.NoError(t, err)
	require.False(t, drainNotify(t, upgraded))
	require.Equal(t, int64(0), src.binHits.Load(), "版本一致不应触发下载")
}

// TestAntiLoopContentAlreadyCurrent 覆盖"版本号滞后(未随二进制重编译)"的防空转:
// 现役二进制内容已等于远端新版本,即便版本号不同也不得下载/替换。
func TestAntiLoopContentAlreadyCurrent(t *testing.T) {
	payload := []byte("already-current-binary")
	src := newFakeSource(t, payload, "v2.0.0")

	u, bin := newFullUpdater(t, Options{
		CheckURL:       src.srv.URL + "/manifest",
		CurrentVersion: "v0.9.0", // 内置版本号滞后
	})
	require.NoError(t, os.WriteFile(bin, payload, 0o755)) // 但内容已就位
	upgraded := make(chan struct{}, 1)

	err := u.CheckAndUpgrade(context.Background(), upgraded)
	require.NoError(t, err)
	require.False(t, drainNotify(t, upgraded))
	require.Equal(t, int64(0), src.binHits.Load(), "内容一致不得重复下载")
}

func TestCheckAndUpgradeTamperedPayload(t *testing.T) {
	good := []byte("checksummed-binary-content")
	src := newFakeSource(t, good, "v2.0.0")

	// 传输中被换包:清单摘要针对 good,服务端实际下发的是篡改内容
	tampered := append([]byte(nil), good...)
	tampered[0] = 'X'
	src.payload = tampered

	u, bin := newFullUpdater(t, Options{
		CheckURL:       src.srv.URL + "/manifest",
		CurrentVersion: "v1.0.0",
	})

	err := u.CheckAndUpgrade(context.Background(), nil)
	require.ErrorIs(t, err, ErrVerify)

	got, rerr := os.ReadFile(bin)
	require.NoError(t, rerr)
	require.Equal(t, []byte("OLD-BINARY"), got, "校验失败不得替换二进制")
	require.Empty(t, u.restarter.targetPath, "未提交成功不得 Commit 重启目标路径")
}

func TestCheckAndUpgradeManifestErrors(t *testing.T) {
	t.Run("更新源 404", func(t *testing.T) {
		srv := httptest.NewServer(http.NotFoundHandler())
		t.Cleanup(srv.Close)
		u, _ := newFullUpdater(t, Options{CheckURL: srv.URL + "/manifest", CurrentVersion: "v1"})

		err := u.CheckAndUpgrade(context.Background(), nil)
		require.ErrorIs(t, err, ErrNetwork)
	})

	t.Run("清单非 JSON", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("<html>oops"))
		}))
		t.Cleanup(srv.Close)
		u, _ := newFullUpdater(t, Options{CheckURL: srv.URL, CurrentVersion: "v1"})

		err := u.CheckAndUpgrade(context.Background(), nil)
		require.ErrorIs(t, err, ErrPolicy)
	})
}

// TestCheckAndUpgradeConcurrent 校验 inflight CAS:并发第二轮应立即返回,不重复下载。
func TestCheckAndUpgradeConcurrent(t *testing.T) {
	payload := []byte("checksummed-binary-content")
	src := newFakeSource(t, payload, "v2.0.0")
	src.gate = make(chan struct{}) // 第一轮进入请求后即阻塞

	u, _ := newFullUpdater(t, Options{
		CheckURL:       src.srv.URL + "/manifest",
		CurrentVersion: "v1.0.0",
	})
	upgraded := make(chan struct{}, 1)

	firstDone := make(chan error, 1)
	go func() {
		firstDone <- u.CheckAndUpgrade(context.Background(), upgraded)
	}()

	// 等第一轮真正进入清单请求(被 gate 卡住)
	deadline := time.Now().Add(3 * time.Second)
	for src.manifestHits.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	require.Equal(t, int64(1), src.manifestHits.Load(), "第一轮应进入请求")

	// 第二轮:inflight 命中,立即返回且不发通知
	err2 := u.CheckAndUpgrade(context.Background(), upgraded)
	require.NoError(t, err2)
	require.False(t, drainNotify(t, upgraded))

	// 放行第一轮
	close(src.gate)
	require.NoError(t, <-firstDone)
	require.True(t, drainNotify(t, upgraded))
	require.Equal(t, int64(1), src.binHits.Load(), "并发下二进制只能下载一次")
	t.Cleanup(func() { _ = os.Unsetenv(EnvVersionOverride) })
}

func TestRunLoopUpgradeAndCancel(t *testing.T) {
	payload := []byte("loop-version-binary")
	src := newFakeSource(t, payload, "v3.0.0")

	u, bin := newFullUpdater(t, Options{
		CheckURL:               src.srv.URL + "/manifest",
		CurrentVersion:         "v1.0.0",
		CheckIntervalSeconds:   60, // 仅首轮延迟 0,之后由取消打断
		FirstCheckDelaySeconds: 0,
	})
	upgraded := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		u.runLoop(ctx, upgraded)
		close(done)
	}()

	select {
	case <-upgraded:
	case <-time.After(3 * time.Second):
		t.Fatal("首轮检查应在延迟 0 时立即触发并升级")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("ctx 取消后 runLoop 应退出")
	}

	got, rerr := os.ReadFile(bin)
	require.NoError(t, rerr)
	require.Equal(t, payload, got)
	require.Equal(t, int64(1), src.binHits.Load())
	t.Cleanup(func() { _ = os.Unsetenv(EnvVersionOverride) })
}

// TestRunNoCheckURLGates 三平台通用:check_url 为空时 Run 立即返回(不再轮询)。
func TestRunNoCheckURLGates(t *testing.T) {
	u := New(Options{}, nil)
	done := make(chan struct{})
	go func() {
		u.Run(context.Background(), make(chan struct{}, 1))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("check_url 为空时 Run 应立即返回")
	}
}
