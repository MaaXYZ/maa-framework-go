package buffer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringBuffer_GetWithSize_PreservesBytes(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{name: "empty", content: ""},
		{name: "plain", content: "buffer content"},
		{name: "embedded NUL", content: "before\x00after"},
		{name: "leading NUL", content: "\x00after"},
		{name: "trailing NUL", content: "before\x00"},
		{name: "multiple NUL", content: "before\x00\x00middle\x00after"},
		{name: "UTF-8 with NUL", content: "中文\x00内容"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stringBuffer := createStringBuffer(t)
			t.Cleanup(stringBuffer.Destroy)

			require.True(t, stringBuffer.SetWithSize(tc.content, uint64(len(tc.content))))
			require.Equal(t, uint64(len(tc.content)), stringBuffer.Size())
			require.Equal(t, tc.content, stringBuffer.GetWithSize())
			require.Equal(t, strings.SplitN(tc.content, "\x00", 2)[0], stringBuffer.Get())
		})
	}
}

func TestStringBuffer_GetWithSize_ReturnsOwnedString(t *testing.T) {
	stringBuffer := createStringBuffer(t)
	destroyed := false
	t.Cleanup(func() {
		if !destroyed {
			stringBuffer.Destroy()
		}
	})

	const content = "before\x00中文\x00after"
	require.True(t, stringBuffer.SetWithSize(content, uint64(len(content))))
	got := stringBuffer.GetWithSize()
	require.Equal(t, content, got)

	replacement := strings.Repeat("x", len(content))
	require.True(t, stringBuffer.SetWithSize(replacement, uint64(len(replacement))))
	require.Equal(t, replacement, stringBuffer.GetWithSize())
	require.Equal(t, content, got, "rewriting the native buffer must not change the returned string")

	require.True(t, stringBuffer.Clear())
	require.Empty(t, stringBuffer.GetWithSize())
	require.Equal(t, content, got, "clearing the native buffer must not change the returned string")

	stringBuffer.Destroy()
	destroyed = true
	require.Equal(t, content, got, "the returned string must remain valid after the native buffer is destroyed")
}
