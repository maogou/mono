//go:build windows

package updater

import (
	"fmt"
	"os"
	"os/exec"
)

// exec Windows 重启实现,语义对齐 minio/minio cmd/service.go 的
// restartProcess Windows 分支:Windows 无 execve 同 PID 换壳,改为以自身
// argv/env 从正式路径启动一个新实例(apply 已完成,磁盘上即新版本),并
// 同步等待其退出——阻塞期间本进程作为外层存活,保证任意时刻只有一个实例
// 运行,外部守护/服务管理器不会额外拉起第二份造成抢端口。
// 优先执行 Commit 记录的正式路径;空值回退当前进程路径。
//
// 新实例正常结束(如收到退出指令完成优雅停服)后,本进程随 os.Exit(0) 退出,
// 不再返回调用方;启动失败或新实例异常退出则返回错误,由调用方打印并
// 以非 0 退出码交守护处理(磁盘已是新版本,守护拉起即新版本)。
func (r *Restarter) exec() error {
	exe := r.targetPath
	if exe == "" {
		// 未经过 apply(如外部替换场景),退回当前进程路径
		var err error
		if exe, err = os.Executable(); err != nil {
			return err
		}
	}

	// #nosec G204 -- 目标为已提交的正式路径,argv/env/标准流均来自进程自身,
	// 无 shell 参与、无注入面
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restart %s: %w", exe, err)
	}
	os.Exit(0)
	return nil // 不可达:成功路径已 os.Exit;保持返回类型完整
}
