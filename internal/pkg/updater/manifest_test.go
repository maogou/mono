package updater

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func validManifest(t *testing.T) *Manifest {
	t.Helper()
	return &Manifest{
		Version:     "v2.0.0",
		DownloadURL: "http://update.example.com/bin/app-v2",
		MD5:         strings.Repeat("ab", 16),
	}
}

func TestParseManifestOK(t *testing.T) {
	m := validManifest(t)
	data, err := jsonMarshal(m)
	require.NoError(t, err)

	got, err := ParseManifest(data)
	require.NoError(t, err)
	require.Equal(t, m.Version, got.Version)
	require.Equal(t, m.DownloadURL, got.DownloadURL)
	require.Equal(t, m.MD5, got.MD5)
}

func TestParseManifestInvalid(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Manifest)
	}{
		{"缺少 version", func(m *Manifest) { m.Version = " " }},
		{"缺少 download_url", func(m *Manifest) { m.DownloadURL = "" }},
		{"download_url scheme 非法", func(m *Manifest) { m.DownloadURL = "ftp://x/bin" }},
		{"download_url 相对路径", func(m *Manifest) { m.DownloadURL = "/bin" }},
		{"md5 非 hex", func(m *Manifest) { m.MD5 = strings.Repeat("zz", 16) }},
		{"md5 长度不足", func(m *Manifest) { m.MD5 = "abc" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := validManifest(t)
			tc.mutate(m)
			data, err := jsonMarshal(m)
			require.NoError(t, err)

			_, err = ParseManifest(data)
			require.Error(t, err)
			require.ErrorIs(t, err, ErrPolicy)
		})
	}
}

func TestParseManifestNotJSON(t *testing.T) {
	_, err := ParseManifest([]byte("<html>not json"))
	require.ErrorIs(t, err, ErrPolicy)
}
