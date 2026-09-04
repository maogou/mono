package updater

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

// Manifest 是更新源(服务端 JSON 接口)返回的升级信息契约,字段全部必填:
//   - Version:发布版本号,与当前版本号(编译期内置常量,env 可覆盖)不相等才升级
//   - DownloadURL:新二进制下载地址(绝对 http/https)
//   - MD5:新二进制 md5(32 位十六进制),客户端仅校验此完整性
//     注:md5 按服务端既有契约选定,只防下载损坏/误传,不提供来源认证
//     (碰撞攻击需先攻破更新源本身,信任边界见 doc/self-update.md)。
type Manifest struct {
	Version     string `json:"version"`
	DownloadURL string `json:"download_url"`
	MD5         string `json:"md5"`
}

// ParseManifest 解析并校验服务端返回的清单。契约违规视为策略错误。
func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%w: 解析升级信息失败: %v", ErrPolicy, err)
	}
	if err := m.validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

func (m *Manifest) validate() error {
	if len(strings.TrimSpace(m.Version)) == 0 {
		return fmt.Errorf("%w: 升级信息缺少 version", ErrPolicy)
	}
	u, err := url.Parse(m.DownloadURL)
	if err != nil || !u.IsAbs() || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("%w: 升级信息 download_url 非法: %q", ErrPolicy, m.DownloadURL)
	}
	if _, err := hex.DecodeString(m.MD5); err != nil || len(m.MD5) != md5HexLen {
		return fmt.Errorf("%w: 升级信息 md5 非法(需 %d 位十六进制)", ErrPolicy, md5HexLen)
	}
	return nil
}

const md5HexLen = 32

// decodedMD5 返回清单中的 md5 二进制值(validate 已保证可解码)。
func (m *Manifest) decodedMD5() []byte {
	sum, _ := hex.DecodeString(m.MD5)
	return sum
}

// logFields 提取清单的日志字段,便于统一打印。
func (m *Manifest) logFields() []zap.Field {
	return []zap.Field{
		zap.String("version", m.Version),
		zap.String("download_url", m.DownloadURL),
		zap.String("md5", m.MD5),
	}
}
