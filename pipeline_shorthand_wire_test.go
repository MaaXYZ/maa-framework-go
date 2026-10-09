package maa

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// The tests in this file pin the protocol shorthand decoding contracts
// independently of the assertions updated inside the fix commits, so reverting
// a fix turns them red: a recognition object without "param" decodes its
// parameters from the whole object, and next/on_error accept the protocol's
// string and single-value shorthand forms.

func TestRecognition_AbsentParamFallsBackToObject(t *testing.T) {
	t.Run("flat fields decode into the typed param", func(t *testing.T) {
		var recognition Recognition
		require.NoError(t, json.Unmarshal([]byte(`{"type":"ColorMatch","lower":[[1,2,3]],"upper":[[9,9,9]]}`), &recognition))
		param, ok := recognition.Param.(*ColorMatchParam)
		require.True(t, ok, "flat fields should decode into the typed param, got %T", recognition.Param)
		require.Equal(t, [][]int{{1, 2, 3}}, param.Lower)
		require.Equal(t, [][]int{{9, 9, 9}}, param.Upper)
	})

	t.Run("flat fields survive a re-encode in nested form", func(t *testing.T) {
		var recognition Recognition
		require.NoError(t, json.Unmarshal([]byte(`{"type":"ColorMatch","lower":[[1,2,3]],"upper":[[9,9,9]]}`), &recognition))
		encoded, err := json.Marshal(&recognition)
		require.NoError(t, err)
		require.JSONEq(t, `{"type":"ColorMatch","param":{"lower":[[1,2,3]],"upper":[[9,9,9]]}}`, string(encoded))
	})

	t.Run("known type without any fields decodes a zero param", func(t *testing.T) {
		var recognition Recognition
		require.NoError(t, json.Unmarshal([]byte(`{"type":"TemplateMatch"}`), &recognition))
		require.Equal(t, &TemplateMatchParam{}, recognition.Param)
	})

	t.Run("explicit param wins over flat fields", func(t *testing.T) {
		var recognition Recognition
		require.NoError(t, json.Unmarshal([]byte(`{"type":"ColorMatch","lower":[[9,9,9]],"param":{"count":5}}`), &recognition))
		param, ok := recognition.Param.(*ColorMatchParam)
		require.True(t, ok)
		require.Equal(t, 5, param.Count)
		require.Nil(t, param.Lower, "flat fields must be ignored when param is present, matching the native parser")
	})

	t.Run("explicit null param still decodes nil", func(t *testing.T) {
		var recognition Recognition
		require.NoError(t, json.Unmarshal([]byte(`{"type":"ColorMatch","param":null}`), &recognition))
		require.Nil(t, recognition.Param)
	})

	t.Run("unknown type keeps the whole object raw", func(t *testing.T) {
		var recognition Recognition
		require.NoError(t, json.Unmarshal([]byte(`{"type":"FutureRecognition","foo":[1]}`), &recognition))
		raw, ok := recognition.Param.(*RawRecognitionParam)
		require.True(t, ok, "unknown types should keep the whole object, got %T", recognition.Param)
		require.JSONEq(t, `{"type":"FutureRecognition","foo":[1]}`, string(*raw))
	})

	t.Run("failure invariance on a bad flat field", func(t *testing.T) {
		recognition := Recognition{Type: RecognitionTypeColorMatch, Param: &ColorMatchParam{Count: 3}}
		before := recognition
		require.Error(t, json.Unmarshal([]byte(`{"type":"ColorMatch","lower":"x"}`), &recognition))
		require.Equal(t, before, recognition)
	})

	t.Run("inline sub-recognition inherits the fallback", func(t *testing.T) {
		var item SubRecognitionItem
		require.NoError(t, json.Unmarshal([]byte(`{"sub_name":"s","recognition":{"type":"ColorMatch","lower":[[1,2,3]]}}`), &item))
		require.NotNil(t, item.Inline)
		param, ok := item.Inline.Recognition.Param.(*ColorMatchParam)
		require.True(t, ok, "sub-recognition should decode flat fields, got %T", item.Inline.Recognition.Param)
		require.Equal(t, [][]int{{1, 2, 3}}, param.Lower)
	})
}

func TestNode_NextOnErrorShorthand(t *testing.T) {
	t.Run("single string becomes a one-item list", func(t *testing.T) {
		var node Node
		require.NoError(t, json.Unmarshal([]byte(`{"next":"A"}`), &node))
		require.Equal(t, []NextItem{{Name: "A"}}, node.Next)
	})

	t.Run("single object becomes a one-item list", func(t *testing.T) {
		var node Node
		require.NoError(t, json.Unmarshal([]byte(`{"on_error":{"name":"A","anchor":true}}`), &node))
		require.Equal(t, []NextItem{{Name: "A", Anchor: true}}, node.OnError)
	})

	t.Run("mixed array of strings and objects", func(t *testing.T) {
		var node Node
		require.NoError(t, json.Unmarshal([]byte(`{"next":["A",{"name":"B","jump_back":true}]}`), &node))
		require.Equal(t, []NextItem{{Name: "A"}, {Name: "B", JumpBack: true}}, node.Next)
	})

	t.Run("attribute prefixes set the corresponding flags", func(t *testing.T) {
		var node Node
		require.NoError(t, json.Unmarshal([]byte(`{"next":"[JumpBack][Anchor]X"}`), &node))
		require.Equal(t, []NextItem{{Name: "X", JumpBack: true, Anchor: true}}, node.Next)
	})

	t.Run("unrecognized prefix is ignored", func(t *testing.T) {
		var node Node
		require.NoError(t, json.Unmarshal([]byte(`{"next":"[Typo]Y"}`), &node))
		require.Equal(t, []NextItem{{Name: "Y"}}, node.Next)
	})

	t.Run("shorthand re-encodes in object form", func(t *testing.T) {
		var node Node
		require.NoError(t, json.Unmarshal([]byte(`{"next":"[JumpBack]A"}`), &node))
		encoded, err := json.Marshal(&node)
		require.NoError(t, err)
		require.JSONEq(t, `{"next":[{"name":"A","jump_back":true,"anchor":false}]}`, string(encoded))
	})

	t.Run("null and empty values stay nil-length", func(t *testing.T) {
		var node Node
		require.NoError(t, json.Unmarshal([]byte(`{"next":null,"on_error":[]}`), &node))
		require.Empty(t, node.Next)
		require.Empty(t, node.OnError)
	})

	t.Run("other node fields survive normalization", func(t *testing.T) {
		var node Node
		input := `{"anchor":{"a":"N"},"recognition":{"type":"ColorMatch","lower":[[1,2,3]],"upper":[[9,9,9]]},"next":"self","rate_limit":500,"enabled":false,"attach":{"x":1},"unknown_key":{"deep":[1,2]}}`
		require.NoError(t, json.Unmarshal([]byte(input), &node))
		require.Equal(t, map[string]string{"a": "N"}, node.Anchor)
		require.NotNil(t, node.Recognition)
		require.Equal(t, int64(500), *node.RateLimit)
		require.NotNil(t, node.Enabled)
		require.False(t, *node.Enabled)
		require.Equal(t, map[string]any{"x": float64(1)}, node.Attach)
		require.Equal(t, []NextItem{{Name: "self"}}, node.Next)
	})

	t.Run("failure invariance on a bad next entry", func(t *testing.T) {
		rateLimit := int64(100)
		node := Node{Next: []NextItem{{Name: "keep"}}, RateLimit: &rateLimit}
		require.Error(t, json.Unmarshal([]byte(`{"next":["A",null]}`), &node))
		require.Equal(t, []NextItem{{Name: "keep"}}, node.Next)
		require.Equal(t, int64(100), *node.RateLimit)
	})

	t.Run("rejects invalid entries", func(t *testing.T) {
		for name, input := range map[string]string{
			"unclosed prefix":     `{"next":"[JumpBack"}`,
			"empty name":          `{"next":"[JumpBack]"}`,
			"null array entry":    `{"next":["A",null]}`,
			"number array entry":  `{"next":[1]}`,
			"single number":       `{"next":5}`,
			"unclosed on_error":   `{"on_error":"[Anchor"}`,
			"boolean array entry": `{"on_error":[true]}`,
		} {
			t.Run(name, func(t *testing.T) {
				var node Node
				require.Error(t, json.Unmarshal([]byte(input), &node), "%s should be rejected", input)
			})
		}
	})
}
