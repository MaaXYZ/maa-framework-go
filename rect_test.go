package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRect_UnmarshalJSON_Shapes(t *testing.T) {
	t.Run("two-element point expands to a 1x1 rect", func(t *testing.T) {
		var r Rect
		require.NoError(t, unmarshalJSON([]byte(`[5,10]`), &r))
		require.Equal(t, Rect{5, 10, 1, 1}, r)
	})

	t.Run("four-element rect decodes as-is", func(t *testing.T) {
		var r Rect
		require.NoError(t, unmarshalJSON([]byte(`[1,2,3,4]`), &r))
		require.Equal(t, Rect{1, 2, 3, 4}, r)
	})

	t.Run("invalid shapes are rejected", func(t *testing.T) {
		for _, input := range []string{
			`[1,2,3]`, `[1,2,3,4,5]`, `[]`, `null`, `["a","b"]`, `[1.5,2]`,
			`[null,2]`, `5`, `"rect"`, `{"x":1}`,
		} {
			var r Rect
			require.Error(t, unmarshalJSON([]byte(input), &r), "input %s", input)
		}
	})

	t.Run("destination is preserved on error", func(t *testing.T) {
		r := Rect{7, 8, 9, 10}
		require.Error(t, unmarshalJSON([]byte(`[1,2,3]`), &r))
		require.Equal(t, Rect{7, 8, 9, 10}, r)
	})
}

func TestRect_IsZero(t *testing.T) {
	require.True(t, Rect{}.IsZero())
	require.False(t, Rect{1, 0, 0, 0}.IsZero())
	require.False(t, Rect{0, 0, 1, 0}.IsZero())
	require.False(t, Rect{0, 0, 0, 1}.IsZero())
}

func TestRect_ZeroOmittedFromEncodedJSON(t *testing.T) {
	encoded, err := marshalJSON(ClickParam{Target: NewTargetRect(Rect{1, 2, 3, 4})})
	require.NoError(t, err)
	require.JSONEq(t, `{"target":[1,2,3,4]}`, string(encoded))
}
