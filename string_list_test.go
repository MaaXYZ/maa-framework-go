package maa

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringList_JSON(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  StringList
	}{
		{`"Cat"`, StringList{"Cat"}},
		{`""`, StringList{""}},
		{` ["Cat","Dog"] `, StringList{"Cat", "Dog"}},
		{`[]`, StringList{}},
	} {
		t.Run(tc.input, func(t *testing.T) {
			var got StringList
			require.NoError(t, json.Unmarshal([]byte(tc.input), &got))
			require.Equal(t, tc.want, got)
		})
	}
}

func TestStringList_RejectInvalidJSON(t *testing.T) {
	for _, input := range []string{
		`null`, `true`, `false`, `1`, `1.5`, `{}`, `[null]`, `["Cat",1]`, `["Cat",`,
	} {
		t.Run(input, func(t *testing.T) {
			got := StringList{"original"}
			require.Error(t, json.Unmarshal([]byte(input), &got))
			require.Equal(t, StringList{"original"}, got)
		})
	}
}

func TestStringList_ParamFields(t *testing.T) {
	t.Run("classify labels accept a scalar", func(t *testing.T) {
		var param NeuralNetworkClassifyParam
		require.NoError(t, json.Unmarshal([]byte(`{"labels":"Cat"}`), &param))
		require.Equal(t, StringList{"Cat"}, param.Labels)
	})
	t.Run("detect labels accept a scalar", func(t *testing.T) {
		var param NeuralNetworkDetectParam
		require.NoError(t, json.Unmarshal([]byte(`{"labels":"Cat"}`), &param))
		require.Equal(t, StringList{"Cat"}, param.Labels)
	})
	t.Run("ocr expected accepts a scalar", func(t *testing.T) {
		var param OCRParam
		require.NoError(t, json.Unmarshal([]byte(`{"expected":"Cat"}`), &param))
		require.Equal(t, StringList{"Cat"}, param.Expected)
	})
	t.Run("labels null is rejected", func(t *testing.T) {
		param := NeuralNetworkClassifyParam{Labels: StringList{"original"}}
		require.Error(t, json.Unmarshal([]byte(`{"labels":null}`), &param))
		require.Equal(t, StringList{"original"}, param.Labels)
	})
	for _, tc := range []struct {
		name  string
		key   string
		param any
	}{
		{"classify labels", "labels", NeuralNetworkClassifyParam{Labels: StringList{}}},
		{"detect labels", "labels", NeuralNetworkDetectParam{Labels: StringList{}}},
		{"ocr expected", "expected", OCRParam{Expected: StringList{}}},
	} {
		t.Run(tc.name+" empty list marshals as []", func(t *testing.T) {
			data, err := json.Marshal(tc.param)
			require.NoError(t, err)
			require.JSONEq(t, `{"`+tc.key+`":[]}`, string(data))
		})
	}
	for _, tc := range []struct {
		name  string
		param any
	}{
		{"classify labels", NeuralNetworkClassifyParam{Labels: nil}},
		{"detect labels", NeuralNetworkDetectParam{Labels: nil}},
		{"ocr expected", OCRParam{Expected: nil}},
	} {
		t.Run(tc.name+" nil list is omitted", func(t *testing.T) {
			data, err := json.Marshal(tc.param)
			require.NoError(t, err)
			require.JSONEq(t, `{}`, string(data))
		})
	}
}
