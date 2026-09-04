package updater

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerifyOK(t *testing.T) {
	payload := []byte("fake-binary-payload")
	m := &Manifest{MD5: md5HexOf(payload)}

	err := New(Options{}, nil).verify(m, payload)
	require.NoError(t, err)
}

func TestVerifyChecksumMismatch(t *testing.T) {
	payload := []byte("fake-binary-payload")
	m := &Manifest{MD5: md5HexOf([]byte("other-content"))} // 清单摘要与实际内容不符

	err := New(Options{}, nil).verify(m, payload)
	require.ErrorIs(t, err, ErrVerify)
}

func TestVerifyPayloadTampered(t *testing.T) {
	good := []byte("fake-binary-payload")
	m := &Manifest{MD5: md5HexOf(good)}

	// 传输中被篡改 1 字节(摘要针对原始内容)
	tampered := append([]byte(nil), good...)
	tampered[3] ^= 0xff

	err := New(Options{}, nil).verify(m, tampered)
	require.ErrorIs(t, err, ErrVerify)
}
