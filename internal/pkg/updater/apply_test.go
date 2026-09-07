package updater

import (
	"bytes"
	"crypto"
	"os"
	"path/filepath"
	"runtime"
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

	// 权限沿用现役文件:Windows 无执行位概念恒为 0666,故以"与替换前一致"
	// 为断言,unix 下即等于 newApplyTarget 写入的 0755
	info, err := os.Stat(bin)
	require.NoError(t, err)
	modeBefore := info.Mode().Perm()

	require.NoError(t, u.apply(m, payload))

	got, err := os.ReadFile(bin)
	require.NoError(t, err)
	require.Equal(t, payload, got)

	info, err = os.Stat(bin)
	require.NoError(t, err)
	require.Equal(t, modeBefore, info.Mode().Perm(), "权限位应沿用现役文件")

	if runtime.GOOS != "windows" {
		// 无 .new/.old 残留;windows 上库对移除失败的 .old 采用隐藏(库文档行为),不适用
		entries, err := os.ReadDir(filepath.Dir(bin))
		require.NoError(t, err)
		for _, e := range entries {
			require.False(t,
				strings.HasSuffix(e.Name(), ".new") || strings.HasSuffix(e.Name(), ".old"),
				"残留临时文件: %s", e.Name())
		}
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

// TestCommitFailureKeepsOldFile 模拟 commit 阶段的 rename 失败(失败注入方式见
// blockCommit,unix=只读目录,windows=独占句柄):
// 验证旧文件仍在原处、错误类别正确、且未触发"回滚失败"误报(RollbackError == nil)。
func TestCommitFailureKeepsOldFile(t *testing.T) {
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

	// 第 2 步起让 commit 的 rename 在"旧文件移开"一步即失败
	blockCommit(t, bin, dir)

	err := selfupdate.CommitBinary(opts)
	require.Error(t, err)
	require.Nil(t, selfupdate.RollbackError(err), "此处回滚未执行,不应误报回滚失败")

	got, rerr := os.ReadFile(bin)
	require.NoError(t, rerr)
	require.Equal(t, []byte("OLD-BINARY"), got, "commit 失败后旧文件必须仍在原处")
}
