//go:build !windows

package updater

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// blockCommit 让 CommitBinary 的 rename 必然失败:把目录改为只读后,
// rename 在"旧文件移开"(target→.old)一步即因目录不可写而失败,
// 旧文件保留在原路径,与后续断言语义一致。
func blockCommit(t *testing.T, _, dir string) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("root 不受目录权限约束,跳过只读目录用例")
	}
	require.NoError(t, os.Chmod(dir, 0o555))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
}
