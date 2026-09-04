package updater

import (
	"bytes"
	"crypto"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/minio/selfupdate"
	"github.com/stretchr/testify/require"
)

// newApplyTarget 在临时目录造一个"现役二进制",返回其路径与 updater(targetPath 已注入)。
func newApplyTarget(t *testing.T) (*Updater, string) {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "app")
	require.NoError(t, os.WriteFile(bin, []byte("OLD-BINARY"), 0o755))
	u := New(Options{}, nil)
	u.targetPath = bin
	return u, bin
}

func TestApplySuccessReplacesBinary(t *testing.T) {
	u, bin := newApplyTarget(t)
	payload := []byte("NEW-BINARY-CONTENT-v2")
	m := &Manifest{MD5: md5HexOf(payload)}

	require.NoError(t, u.apply(m, payload))

	got, err := os.ReadFile(bin)
	require.NoError(t, err)
	require.Equal(t, payload, got)

	// 权限沿用现役文件的 0755,且无 .new/.old 残留
	info, err := os.Stat(bin)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o755), info.Mode().Perm())

	entries, err := os.ReadDir(filepath.Dir(bin))
	require.NoError(t, err)
	for _, e := range entries {
		require.False(t,
			strings.HasSuffix(e.Name(), ".new") || strings.HasSuffix(e.Name(), ".old"),
			"残留临时文件: %s", e.Name())
	}
}

func TestApplyChecksumMismatchKeepsOldFile(t *testing.T) {
	u, bin := newApplyTarget(t)
	payload := []byte("NEW-BINARY-CONTENT-v2")
	// 清单 md5 与真实内容不符(verify 漏网时由库内二次校验兜底)
	m := &Manifest{MD5: md5HexOf([]byte("someone-else"))}

	err := u.apply(m, payload)
	require.ErrorIs(t, err, ErrCommit)

	got, rerr := os.ReadFile(bin)
	require.NoError(t, rerr)
	require.Equal(t, []byte("OLD-BINARY"), got, "失败时不得改动现役文件")
}

// TestCommitFailureKeepsOldFile 用只读目录模拟 commit 阶段的 rename 失败:
// 验证旧文件仍在原处、错误类别正确、且未触发"回滚失败"误报(RollbackError == nil)。
func TestCommitFailureKeepsOldFile(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root 不受目录权限约束,跳过只读目录用例")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "app")
	require.NoError(t, os.WriteFile(bin, []byte("OLD-BINARY"), 0o755))

	payload := []byte("NEW-BINARY-CONTENT-v2")
	// 与生产 apply 的 Options 保持一致(md5 + 显式 Hash),直测库的校验路径
	opts := selfupdate.Options{
		TargetPath: bin,
		TargetMode: 0o755,
		Hash:       crypto.MD5,
		Checksum:   mustDecodeMD5(t, md5HexOf(payload)),
	}

	// 第 1 步 prepare 正常(此时目录可写),产出 .app.new
	require.NoError(t, selfupdate.PrepareAndCheckBinary(bytes.NewReader(payload), opts))

	// 第 2 步把目录改只读,commit 的 rename 必然失败
	require.NoError(t, os.Chmod(dir, 0o555))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	err := selfupdate.CommitBinary(opts)
	require.Error(t, err)
	require.Nil(t, selfupdate.RollbackError(err), "此处回滚未执行,不应误报回滚失败")

	got, rerr := os.ReadFile(bin)
	require.NoError(t, rerr)
	require.Equal(t, []byte("OLD-BINARY"), got, "commit 失败后旧文件必须仍在原处")
}
