package maa

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOCRReplaceShorthand_NormalizesValidPair(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
		want  [2]string
	}{
		{"ordinary pair", `["old","new"]`, [2]string{"old", "new"}},
		{"empty strings", `["",""]`, [2]string{"", ""}},
		{"UTF-8 strings", `["错字","正确🙂"]`, [2]string{"错字", "正确🙂"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var param OCRParam
			require.NoError(t, json.Unmarshal([]byte(`{"replace":`+tc.input+`}`), &param))
			require.Equal(t, [][2]string{tc.want}, param.Replace)

			encoded, err := json.Marshal(param)
			require.NoError(t, err)
			require.JSONEq(t, `{"replace":[`+tc.input+`]}`, string(encoded))
		})
	}
}

func TestOCRReplaceShorthand_RejectsInvalidPairWithoutMutation(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
	}{
		{"one element", `["old"]`},
		{"three elements", `["old","new","extra"]`},
		{"null pattern", `[null,"new"]`},
		{"null replacement", `["old",null]`},
		{"numeric pattern", `[7,"new"]`},
		{"numeric replacement", `["old",7]`},
		{"boolean replacement", `["old",false]`},
		{"object replacement", `["old",{}]`},
		{"single string", `"old"`},
		{"single number", `7`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expected := StringList{"keep"}
			replace := [][2]string{{"old", "fixed"}}
			param := OCRParam{
				ROI: NewTargetString("NodeA"), ROIOffset: Rect{1, 2, 3, 4}, Expected: expected,
				Threshold: 0.7, Replace: replace, OrderBy: OCROrderByExpected,
				Index: 2, OnlyRec: true, Model: "custom", ColorFilter: "ColorNode",
			}
			input := fmt.Sprintf(`{"roi":true,"roi_offset":[9,10],"expected":"new","threshold":0.2,"model":"new","replace":%s}`, tc.input)
			require.Error(t, json.Unmarshal([]byte(input), &param))
			require.Equal(t, OCRParam{
				ROI: NewTargetString("NodeA"), ROIOffset: Rect{1, 2, 3, 4}, Expected: StringList{"keep"},
				Threshold: 0.7, Replace: [][2]string{{"old", "fixed"}}, OrderBy: OCROrderByExpected,
				Index: 2, OnlyRec: true, Model: "custom", ColorFilter: "ColorNode",
			}, param)
			require.Equal(t, StringList{"keep"}, expected)
			require.Equal(t, [][2]string{{"old", "fixed"}}, replace)
		})
	}
}

func TestOCRReplaceShorthand_PreservesListDecoding(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
		want  [][2]string
	}{
		{"empty list", `[]`, [][2]string{}},
		{"null list", `null`, nil},
		{"nested pair", `[["old","new"]]`, [][2]string{{"old", "new"}}},
		{"multiple nested pairs", `[["a","b"],["c","d"]]`, [][2]string{{"a", "b"}, {"c", "d"}}},
		{"short nested pair pads", `[["old"]]`, [][2]string{{"old", ""}}},
		{"long nested pair truncates", `[["old","new","extra"]]`, [][2]string{{"old", "new"}}},
		{"nested null strings", `[[null,null]]`, [][2]string{{"", ""}}},
		{"nested null pair", `[["old","new"],null]`, [][2]string{{"old", "new"}, {"", ""}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			param := OCRParam{Replace: [][2]string{{"keep", "fixed"}}, Model: "custom"}
			require.NoError(t, json.Unmarshal([]byte(`{"replace":`+tc.input+`}`), &param))
			require.Equal(t, tc.want, param.Replace)
			require.Equal(t, "custom", param.Model)
		})
	}
}
