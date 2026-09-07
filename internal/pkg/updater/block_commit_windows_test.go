//go:build windows

package updater

import (
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

// blockCommit 让 CommitBinary 的 rename 必然失败:Windows 上目录的只读属性
// 不阻止目录内文件改名(权限由 ACL 决定),故改为用"不含 FILE_SHARE_DELETE"
// 的句柄独占现役文件,使库的 os.Rename(target→.old) 因共享冲突
// (ERROR_SHARING_VIOLATION)失败——同样发生在"旧文件移开"一步,
// 旧文件保留在原路径,与 unix 用例断言语义一致。
func blockCommit(t *testing.T, bin, _ string) {
	t.Helper()
	p, err := syscall.UTF16PtrFromString(bin)
	require.NoError(t, err)
	h, err := syscall.CreateFile(p,
		syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE, // 故意不含 FILE_SHARE_DELETE
		nil, syscall.OPEN_EXISTING, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	require.NoError(t, err, "独占打开现役文件失败")
	t.Cleanup(func() { _ = syscall.CloseHandle(h) })
}
