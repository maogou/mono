package updater

import "errors"

// 分类错误:上层(调用方)可借 errors.Is 决定重试或告警策略。
var (
	// ErrNetwork 表示与更新源的网络交互失败(可下轮重试)。
	ErrNetwork = errors.New("updater: network error")
	// ErrVerify 表示 md5 完整性校验失败(下载损坏或与清单不一致,不应继续)。
	ErrVerify = errors.New("updater: verify failed")
	// ErrCommit 表示二进制替换失败。
	ErrCommit = errors.New("updater: commit failed")
	// ErrPolicy 表示配置/策略问题(URL 非法、体积超限等,不会自行恢复)。
	ErrPolicy = errors.New("updater: policy error")
)
