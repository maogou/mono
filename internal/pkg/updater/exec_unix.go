//go:build !windows

package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// exec 同 PID 执行新二进制(Linux/macOS 均支持 execve):
// exec 成功即完成进程替换,本函数不返回。优先执行 Commit 记录的正式路径。
func (r *Restarter) exec() error {
	exe := r.targetPath
	if exe == "" {
		// 未经过 apply(如外部替换场景),从 /proc/self/exe 推导正式路径
		var err error
		if exe, err = canonicalExecPath(); err != nil {
			return err
		}
	}
	// #nosec G204,G702 -- 用自身 argv/env 同 PID 换壳(execve 语义):无 shell 参与、
	// 路径为已提交的正式路径、argv/env 均来自进程自身,不存在命令注入面。
	if err := syscall.Exec(exe, os.Args, os.Environ()); err != nil {
		// 附带路径与文件状态,便于定位 exec 失败(如特殊文件系统/挂载缓存问题)
		_, statErr := os.Stat(exe)
		return fmt.Errorf("exec %s: %w (stat: %v)", exe, err, statErr)
	}
	return nil
}

// canonicalExecPath 推导"替换后应执行"的正式可执行文件路径。
// Linux 上被替换后 /proc/self/exe 指向已删除的旧 inode,readlink 返回
// "<路径> (deleted)",且最后链接名可能是 commit rename 走位后的
// ".name.old"/".name.new" 花名——需一并归一化回正式文件名;
// macOS 无 "(deleted)" 后缀也无花名(rename 期间目标名始终为正式路径),
// 经此归一化结果不变,逻辑通用。
func canonicalExecPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	exe = strings.TrimSuffix(exe, " (deleted)")

	dir, name := filepath.Split(exe)
	if name != "" && name[0] == '.' {
		stem := name[1:]
		for _, suffix := range []string{".old", ".new"} {
			if before, ok := strings.CutSuffix(stem, suffix); ok {
				return filepath.Join(dir, before), nil
			}
		}
	}
	return exe, nil
}
