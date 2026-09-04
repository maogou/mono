package updater

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// jsonMarshal 序列化 Manifest 为 JSON(便于模拟服务端接口响应)。
func jsonMarshal(m *Manifest) ([]byte, error) {
	return json.Marshal(m)
}

// mustDecodeMD5 解码 32 位十六进制 md5。
func mustDecodeMD5(t *testing.T, hexMD5 string) []byte {
	t.Helper()
	sum, err := hex.DecodeString(hexMD5)
	require.NoError(t, err)
	return sum
}

// md5HexOf 计算 payload 的 md5 十六进制串。
func md5HexOf(payload []byte) string {
	sum := md5.Sum(payload)
	return hex.EncodeToString(sum[:])
}
