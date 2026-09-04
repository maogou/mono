package updater

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/minio/selfupdate"
	"go.uber.org/zap"
)

// apply 把已通过校验的 payload 原子替换到目标可执行文件。
func (u *Updater) apply(m *Manifest, payload []byte) error {
	path, err := u.exePath()
	if err != nil {
		return fmt.Errorf("%w: 定位目标可执行文件失败: %v", ErrCommit, err)
	}

	sum, err := hex.DecodeString(m.MD5)
	if err != nil { // validate 已保证,此处仅防御
		return fmt.Errorf("%w: 清单 md5 非法: %v", ErrPolicy, err)
	}

	opts := selfupdate.Options{
		TargetPath: path,
		TargetMode: targetMode(path),
		Hash:       crypto.MD5, // 库内按清单同摘要(md5)二次校验,纵深防御
		Checksum:   sum,
	}
	if err = selfupdate.Apply(bytes.NewReader(payload), opts); err != nil {
		if err = selfupdate.RollbackError(err); err != nil {
			u.log.Error("自升级替换失败且回滚失败,需人工恢复", zap.String("path", path), zap.Error(err))
			return fmt.Errorf("%w: 替换失败且回滚失败: %v", ErrCommit, err)
		}
		return fmt.Errorf("%w: 替换失败: %v", ErrCommit, err)
	}
	// 把正式路径记入 Restarter(替换成功后 /proc/self/exe 已指向被删的旧 inode,
	// exec 必须回正式路径执行新二进制);编排方经同一实例 Exec,见 exec.go。
	u.restarter.Commit(path)
	return nil
}

// targetMode 沿用现役文件的权限位;stat 失败时回退 0755。
func targetMode(path string) os.FileMode {
	if info, err := os.Stat(path); err == nil {
		return info.Mode().Perm()
	}
	return 0o755
}
