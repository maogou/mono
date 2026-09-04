package updater

import (
	"bytes"
	"crypto/md5"
	"fmt"
)

// verify 对下载的 payload 做 md5 完整性校验,通过才允许替换。
// 注:md5 仅为服务端既有契约指定的完整性比对(防下载损坏/误传),
// 不提供来源认证——信任根是更新源本身,部署上应由 HTTPS/内网隔离保证来源可信,
// 见 doc/self-update.md 信任边界。
func (u *Updater) verify(m *Manifest, payload []byte) error {
	// #nosec G401 -- md5 为服务端既有契约指定,仅完整性比对、非认证场景
	got := md5.Sum(payload)
	if !bytes.Equal(got[:], m.decodedMD5()) {
		return fmt.Errorf("%w: md5 不匹配,期望 %s,实际 %x", ErrVerify, m.MD5, got)
	}
	return nil
}
