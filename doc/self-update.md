# 进程自升级(self-update)机制与发布说明

> 日期:2026-09-04
> 适用范围:单机 + 外部守护进程(supervise 类)/ 单容器;Linux / macOS / Windows 三平台,
> 重启语义见 §2 与 exec_*.go(2026-09-04 起不再限定 Linux)。
> 2026-09-04 变更:去掉 minisign 签名校验;校验摘要算法按服务端既有契约由 sha256 改为 md5;
> 重启执行收敛为 Restarter 结构体(NewRestarter 构造 → apply 阶段 Commit 正式路径 →
> 编排方 Exec),替代原包级全局 ExecSelf 变量/execTargetPath,无全局可变状态。
> 客户端仅做摘要完整性比对,信任边界与风险见 §6。

## 1. 一句话

进程定期询问**服务端 JSON 更新源**,获得新版本二进制的「下载地址 + md5」;
下载后做 **md5 完整性校验**,通过才用 `minio/selfupdate`
做**原子替换**(同目录 `.new` rename 三连,失败自动回滚);随后**优雅停服**,并在 run() 内显式关闭
DB/Redis/日志(与 cli `After` 同一实现),最后 **`syscall.Exec` 同 PID 换壳重启**——对外部守护进程全程无感。

> 注意:md5 只保证"下载内容与清单一致"(防损坏/误传),**不提供来源认证**。
> 信任根 = 更新源本身;更新源被攻破或 HTTP 链路被改写即可推送任意代码,见 §6。

## 2. 完整时序

```text
updater.Run (全平台)
  │  周期 = check_interval + 0~1/3 随机抖动;启动后首轮延迟 15s
  ▼
CheckAndUpgrade(单轮,CAS 防并发)
  1. GET check_url → ParseManifest(体积上限 1MiB)
  2. version == 当前版本号?→ 跳过
  3. 防空转:当前可执行文件 md5 == 清单 md5?→ 跳过
     (内置版本号滞后时防"替换→重启→再替换"死循环)
  4. GET download_url(HTTP 200,MaxSizeMB 限流,包全量入内存)
  5. verify:md5 比对(仅完整性;来源可信靠更新源,见 §6)
  6. apply: 写 .<name>.new → rename 正名→.<name>.old → rename .new→正名 → 删 .old
     失败自动回滚,回滚也失败记 Error 并报 ErrCommit(需人工恢复)
     成功则 Restarter.Commit(正式路径)——exec 重启目标(替换后 /proc/self/exe 已指向被删的旧 inode)
  7. 注入 env UPDATE_CURRENT_VERSION=新版本号(重启后首轮直接命中跳过分支)
  8. 向 restartCh 发通知(非阻塞)
  ▼
run.go select ← restartCh → restarting=true
  执行既有 srv.Shutdown(ShutdownTimeout 秒内放完在途请求)
  shutdownResources:关 DB/Redis、zap Sync(与 cli After 同一实现;exec 不返回,
  After 不参与该路径,故在此显式收尾)
  restarter.Exec()      # Restarter 实例由 run.go NewRestarter 构造并注入 updater(Options.Restarter),
                        # apply 阶段已 Commit 正式路径,此处同实例重启
  Linux/macOS: syscall.Exec(新二进制路径, os.Args, os.Environ())
    内核:保留 PID、原子关闭全部 CLOEXEC 描述符 → 对 supervise 无感,无端口抢占
  Windows: 以自身 argv/env 启动新实例并阻塞等待其退出,随后本进程 os.Exit(0)
    → 任意时刻仅一个实例在运行(同 minio restartProcess Windows 分支)
  重启失败(极少,如 exec/spawn 报错)→ logger.Error 补记(收尾后再写日志无碍:zap 直写
    无缓冲层,cli After 钩子随后还会兜底再 Sync 一次)并返回错误,main log.Fatal exit(1),
    由守护拉起新二进制(升级目标不变)
```

关键点:**先放端口、再换进程**;SIGTERM 与升级通知竞态时先到先走(若信号先到则进程退出,
外部守护拉起的就是新版本,效果等价)。

## 3. 服务端更新源契约(客户端按此实现校验)

- 接口:**GET check_url**(不带参数,幂等;建议服务端按请求方 IP 做简单限频),返回单个 JSON 对象:

```json
{
  "version": "v1.2.3",
  "download_url": "http://update.example.com/bin/go_template-v1.2.3",
  "md5": "21b236f5...32 位十六进制"
}
```

| 字段 | 必填 | 说明 |
|---|---|---|
| version | ✅ | 与编译期内置版本号 `constant.Version`(env 注入值可覆盖)不相等才会升级 |
| download_url | ✅ | 绝对 http/https 地址 |
| md5 | ✅ | 32 位十六进制,新二进制整体摘要(客户端比对,不一致拒绝替换) |

- 下载地址 GET 建议幂等(同一 URL 内容不变),支持断点重试更佳(客户端目前全量重下)。
- `download_url` 指向的二进制不得超过配置 `max_size_mb`(缺省 128MB),否则拒绝。
- 服务端返回非 200 / 非 JSON / 字段缺失 → 客户端按类别记日志,下个周期重试(不退出进程)。
- **信任边界(重要)**:客户端不做密码学签名认证,md5 无法防止"更新源/传输被伪造"——
  明文 HTTP + 源被攻破即可推送任意二进制并以服务用户身份执行。部署须满足:
  更新源走 **HTTPS 或隔离内网**、托管处做访问控制与防篡改、建议服务进程降权运行(见 §6)。

## 4. 客户端配置

`config/go_template.yaml`(字段全量列表):

```yaml
update:
  enable: false           # 总开关;true 时启动自升级协程
  check_interval: 300     # 检查周期秒,最小 30
  check_url: http://127.0.0.1:8080/update/manifest
  # download_timeout: 30  # 下载/请求超时秒(缺省 30)
  # max_size_mb: 128      # 下载包上限 MB(缺省 128)
```

**版本号不在此配置**:比较基准为编译期内置常量 `internal/constant.Version`
(发版 = bump 常量 → 重新编译,见 §7),另可被 env `UPDATE_CURRENT_VERSION`
覆盖(升级流程重启前自动注入,防一次重启内重复升级);其余无运行时密钥类配置。

## 5. 发布步骤(每次发版)

```bash
# 1) bump 版本号并编译:改 internal/constant.Version → 重新编译
#    (如 CGO_ENABLED=0 go build -o go_template ...)
#    二进制与清单必须出自同一份源码,版本号才一致。

# 2) 生成清单 JSON(md5 + 版本号,--version 缺省读取常量);download_url 直接传入真实地址
go run scripts/manifest/main.go --binary ./go_template \
    --download-url http://update.example.com/bin/go_template-v1.2.3
```

把工具打印的清单 JSON **原样**作为 `check_url` 接口的响应;把二进制部署到 `download_url`。
灰度验证后即完成:版本号随二进制内建,无需逐台 bump 配置(§7)。

Makefile 封装:`make manifest ARGS="--binary ./go_template --download-url http://..."`。

## 6. 信任边界与安全(2026-09-04 起:无签名校验,摘要算法按服务端既有契约改用 md5)

- **现状**:曾采用 minisign(ed25519)签名自认证(明文 HTTP 也可防伪造);为简化密钥分发与发布链路,
  已移除签名与公钥机制;摘要算法同日由 sha256 改为 **md5**(服务端既有契约指定,客户端只做完整性比对)。
- **风险**:更新源服务器被攻破、或 http 明文链路被中间人改写时,攻击者可下发任意二进制;
  进程以服务用户执行(当前部署为 root)→ 等价远程代码执行。md5 的碰撞也已可工程化构造,
  若未来出现"清单走可信通道(HTTPS)、二进制走另一通道"的信任分离部署,md5 挡不住伪造——
  那种场景应回退 sha256 或恢复签名。
- **缓解**(按优先级):
  1. 更新源走 HTTPS 或隔离内网,限制可写面;
  2. 服务进程降权运行(非 root),二进制目录写权限仅授予发布账户;
  3. 更新清单/二进制托管处启用访问审计,对异常版本变化告警;
  4. md5 仍能拦截:下载损坏、错传、内容与清单不一致——防空转比对也依赖它。

## 7. 运维注意事项

- **发版流程**:1) bump `internal/constant.Version` → 2) 重新编译 → 3) `make manifest`
  生成清单(工具缺省读同一常量,版本号天然一致) → 4) 部署二进制与清单。版本号随
  二进制内建,不存在"配置文件版本滞后"类的长期漂移;升级进程重启前仍自动注入
  `UPDATE_CURRENT_VERSION`(防一次重启内的重复升级)。
- 防空转 md5 比对是兜底:内容已就位而版本号滞后(如手工覆盖了二进制却未重编译)时跳过,
  不会"替换→重启→再替换"死循环——但会每轮白跑检查,属需修正的部署错误。
- **ShutdownTimeout 语义**:`srv.Shutdown` 若超时,run.go 直接返回错误、**不会 exec**——
  进程退出,外部守护拉起的新实例即为新版本;无守护时需人工拉起(见已知限制)。
- 适用:裸机 + supervise/runit/systemd、单容器(卷可写、有 restart policy 更稳)。
- **不适用**:k8s 多副本(滚动更新应由控制器做)、只读文件系统(替换写不进)。
- 检查/下载全部在进程内完成,期间服务照常对外;只有「替换成功→重启」的瞬间停服
  (优雅 Shutdown 等完在途请求后 exec,窗口 < 1s)。
- 日志关键字:启动 `自升级已开启`;失败 `updater:` 前缀 + 类别(网络/校验/提交);
  升级就绪 `升级包已替换就绪,即将重启`;exec 失败 `自升级重启失败,交由外部守护拉起新二进制`
  (zap.Error 附 `自升级重启失败(exec): ...`,同时 main log.Fatal 落 stderr)。

## 8. 已知限制

- 二进制**全量驻留内存**(下载 → 校验 → 替换),体积受 `max_size_mb` 限制;流式 hash 留作后续。
- 替换两步 rename 之间存在毫秒级"正名缺失"窗口,非严格原子(与 minio 相同);进程崩溃在最坏
  情况下留下 `.old`,启动路径不做恢复(本模板未实现),依赖外部守护 + 发布灰度兜底。
- 回滚失败(rename 三连第二步失败且回滚失败)只记 Error 日志,需人工恢复;该路径极难触发
  (目录权限剧变/挂载点变化),单测覆盖到 commit 失败保旧文件,回滚失败分支依赖库自身保障。
- 平台支持(参考 minio/minio `restartProcess`,build tag 隔离):
  - **Linux/macOS**:`syscall.Exec` 同 PID 换壳,守护进程全程无感(§2 时序);
  - **Windows**:无 execve,以自身 argv/env 启动新实例并**阻塞等待其退出**(成功路径随
    新实例退出),保证任意时刻只有一个实例在运行——以 Windows 服务/NSSM 托管时,由服务
    管理器直接管理首进程即可(勿再配"退出即拉起第二份"的自动重启,否则与新实例冲突)。
- 无管理 API / 手动触发入口(模板未建),按需再加。

## 9. 相关文件

| 文件 | 职责 |
|---|---|
| `internal/pkg/updater/updater.go` | 主流程(检查/下载/通知),Options 装配(含 Restarter 注入) |
| `internal/pkg/updater/manifest.go` | 服务端契约解析与校验 |
| `internal/pkg/updater/verify.go` | md5 完整性校验 |
| `internal/pkg/updater/apply.go` | minio/selfupdate 原子替换封装;成功后 Restarter.Commit 重启目标 |
| `internal/pkg/updater/err.go` | 分类错误 |
| `internal/pkg/updater/exec.go` | Restarter 结构体(NewRestarter/Commit/Exec),承载重启目标与执行 |
| `internal/pkg/updater/exec_unix.go` `exec_windows.go` | 按平台重启实现:unix syscall.Exec 同 PID 换壳 / windows 拉起新实例(build tag 隔离) |
| `scripts/manifest/main.go` | 发布工具:算 md5 + 输出清单 JSON |
| `doc/upgrade-comparison.md` | 三方升级方案对比(历史决策依据,签名链一节已被 §6 取代) |
