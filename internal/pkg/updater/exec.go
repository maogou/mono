package updater

// Restarter 承载"升级提交后重启进程"的职责与状态,替代原先的包级全局
// (var ExecSelf 函数变量 + var execTargetPath 字符串),消除全局可变数据:
// 提交路径随实例走,不同升级器实例互不干扰,天然无数据竞争。
//
// 生命周期:NewRestarter 构造 → apply 提交成功后 Commit(记录正式路径)→
// 主流程收尾完成后 Exec(按平台重启)。Commit 与 Exec 通常分属两协程
// (升级协程 vs 编排主流程),二者经 restartCh 通道同步,顺序为先 Commit
// 后 Exec,读取发生在写之后(happens-before),无并发修改。
type Restarter struct {
	// targetPath 记录最近一次 apply 提交到的正式路径(apply.go 写)。
	// exec 时优先使用:提交成功后 /proc/self/exe 指向已删除的旧 inode,其最后
	// 链接名可能是 ".name.old",不能直接作为 exec 目标,必须回到正式路径
	// 执行新二进制;空值表示未经过 apply(如外部替换场景),exec 时回退推导。
	targetPath string
}

// NewRestarter 构造重启器。实例为普通值类型,无全局状态,零值亦可使用
// (仅缺省未提交路径的回退逻辑)。
func NewRestarter() *Restarter {
	return &Restarter{}
}

// Commit 记录 apply 提交到的正式路径(替换成功后调用)。幂等,后写覆盖先写。
// 须在 Exec 前调用;两协程经 restartCh 同步,勿与 Exec 并发执行。
func (r *Restarter) Commit(path string) {
	r.targetPath = path
}

// Exec 重启进程,使刚替换的可执行文件生效,供主流程在升级提交、
// 资源收尾完成后调用。
func (r *Restarter) Exec() error {
	return r.exec()
}
