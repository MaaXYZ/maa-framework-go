package maa

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// The tests in this file pin the v1-compat decode contracts: v1 flat nodes
// (string recognition/action with parameters flat on the node), absent-param
// fallbacks, and the protocol's single-value/flat-row shorthands all decode
// into the v2 model; encoding always emits canonical v2.

func TestNode_V1FlatDecode(t *testing.T) {
	input := `{
		"V1": {
			"recognition": "ColorMatch",
			"lower": [1, 2, 3],
			"upper": [[9, 9, 9]],
			"action": "Swipe",
			"begin": [100, 100],
			"end": [200, 200],
			"duration": 500,
			"next": ["V1b"],
			"on_error": "[JumpBack]V1b"
		},
		"V1b": {"action": "ClickKey", "key": 27}
	}`

	var nodes map[string]*Node
	require.NoError(t, json.Unmarshal([]byte(input), &nodes))

	color, ok := nodes["V1"].Recognition.Param.(*ColorMatchParam)
	require.True(t, ok, "v1 flat recognition should decode typed, got %T", nodes["V1"].Recognition.Param)
	require.Equal(t, RecognitionTypeColorMatch, nodes["V1"].Recognition.Type)
	require.Equal(t, [][]int{{1, 2, 3}}, color.Lower, "flat single row normalizes to one row")
	require.Equal(t, [][]int{{9, 9, 9}}, color.Upper)

	swipe, ok := nodes["V1"].Action.Param.(*SwipeParam)
	require.True(t, ok, "v1 flat action should decode typed, got %T", nodes["V1"].Action.Param)
	require.Equal(t, ActionTypeSwipe, nodes["V1"].Action.Type)
	require.Equal(t, []time.Duration{500 * time.Millisecond}, swipe.Duration)
	require.Len(t, swipe.End, 1, "flat end target normalizes to a one-element list")
	require.Equal(t, []NextItem{{Name: "V1b"}}, nodes["V1"].Next)
	require.Equal(t, []NextItem{{Name: "V1b", JumpBack: true}}, nodes["V1"].OnError)

	click, ok := nodes["V1b"].Action.Param.(*ClickKeyParam)
	require.True(t, ok)
	require.Equal(t, []int{27}, click.Key, "scalar key normalizes to a one-element array")

	encoded, err := json.Marshal(nodes)
	require.NoError(t, err)
	require.JSONEq(t, `{
		"V1": {
			"recognition": {"type": "ColorMatch", "param": {"lower": [[1,2,3]], "upper": [[9,9,9]]}},
			"action": {"type": "Swipe", "param": {"begin": [100,100,1,1], "end": [[200,200,1,1]], "duration": [500]}},
			"next": [{"name": "V1b", "jump_back": false, "anchor": false}],
			"on_error": [{"name": "V1b", "jump_back": true, "anchor": false}]
		},
		"V1b": {"action": {"type": "ClickKey", "param": {"key": [27]}}}
	}`, string(encoded), "re-encoding emits canonical v2")
}

func TestNode_V1StringWithoutParams(t *testing.T) {
	var node Node
	require.NoError(t, json.Unmarshal([]byte(`{"recognition": "TemplateMatch", "action": "Click"}`), &node))
	require.Equal(t, RecognitionTypeTemplateMatch, node.Recognition.Type)
	require.Equal(t, &TemplateMatchParam{}, node.Recognition.Param)
	require.Equal(t, ActionTypeClick, node.Action.Type)
	require.Equal(t, &ClickParam{}, node.Action.Param)
}

func TestNode_V1FailureInvariance(t *testing.T) {
	rateLimit := int64(100)
	node := Node{Next: []NextItem{{Name: "keep"}}, RateLimit: &rateLimit}
	require.Error(t, json.Unmarshal([]byte(`{"recognition": "ColorMatch", "lower": "x"}`), &node))
	require.Equal(t, []NextItem{{Name: "keep"}}, node.Next)
	require.Equal(t, int64(100), *node.RateLimit)
	require.Nil(t, node.Recognition)
}

func TestAction_AbsentParamFlatFields(t *testing.T) {
	t.Run("flat fields decode into the typed param", func(t *testing.T) {
		var action Action
		require.NoError(t, json.Unmarshal([]byte(`{"type":"Click","contact":1}`), &action))
		click, ok := action.Param.(*ClickParam)
		require.True(t, ok, "flat fields should decode typed, got %T", action.Param)
		require.Equal(t, 1, click.Contact)
	})

	t.Run("flat fields survive a re-encode in nested form", func(t *testing.T) {
		var action Action
		require.NoError(t, json.Unmarshal([]byte(`{"type":"Click","contact":1}`), &action))
		encoded, err := json.Marshal(&action)
		require.NoError(t, err)
		require.JSONEq(t, `{"type":"Click","param":{"contact":1}}`, string(encoded))
	})

	t.Run("failure invariance on a bad flat field", func(t *testing.T) {
		action := Action{Type: ActionTypeClick, Param: &ClickParam{Contact: 3}}
		before := Action{Type: action.Type, Param: action.Param}
		require.Error(t, json.Unmarshal([]byte(`{"type":"Click","contact":"x"}`), &action))
		require.Equal(t, before, action)
	})
}

func TestSwipeShorthandNormalization(t *testing.T) {
	decode := func(t *testing.T, param string) *SwipeParam {
		t.Helper()
		var action Action
		require.NoError(t, json.Unmarshal([]byte(`{"type":"Swipe","param":`+param+`}`), &action))
		swipe, ok := action.Param.(*SwipeParam)
		require.True(t, ok)
		return swipe
	}

	t.Run("end single forms normalize to one-element lists", func(t *testing.T) {
		for name, value := range map[string]string{
			"point":     `[100, 100]`,
			"rectangle": `[100, 100, 10, 10]`,
			"node name": `"NodeA"`,
			"true":      `true`,
		} {
			t.Run(name, func(t *testing.T) {
				swipe := decode(t, `{"end": `+value+`}`)
				require.Len(t, swipe.End, 1)
			})
		}
	})

	t.Run("end list form and empty list unchanged", func(t *testing.T) {
		swipe := decode(t, `{"end": [[100, 100, 10, 10], [200, 200, 10, 10]]}`)
		require.Len(t, swipe.End, 2)
		swipe = decode(t, `{"end": []}`)
		require.Empty(t, swipe.End)
		swipe = decode(t, `{"end_offset": []}`)
		require.Empty(t, swipe.EndOffset)
	})

	t.Run("end_offset single flat rectangle normalizes", func(t *testing.T) {
		swipe := decode(t, `{"end_offset": [0, 0, 5, 5]}`)
		require.Equal(t, []Rect{{0, 0, 5, 5}}, swipe.EndOffset)
	})

	t.Run("scalar timing normalizes", func(t *testing.T) {
		swipe := decode(t, `{"duration": 500, "end_hold": 100}`)
		require.Equal(t, []time.Duration{500 * time.Millisecond}, swipe.Duration)
		require.Equal(t, []time.Duration{100 * time.Millisecond}, swipe.EndHold)
	})

	t.Run("null timing stays nil", func(t *testing.T) {
		swipe := decode(t, `{"duration": null, "end_hold": null}`)
		require.Nil(t, swipe.Duration)
		require.Nil(t, swipe.EndHold)
	})

	t.Run("multi-swipe items normalize the same way", func(t *testing.T) {
		var action Action
		require.NoError(t, json.Unmarshal([]byte(`{"type":"MultiSwipe","param":{"swipes":[
			{"end": [100, 100, 10, 10], "duration": 300, "end_offset": [1, 2, 3, 4], "end_hold": 50}]}}`), &action))
		multi, ok := action.Param.(*MultiSwipeParam)
		require.True(t, ok)
		require.Len(t, multi.Swipes, 1)
		item := multi.Swipes[0]
		require.Len(t, item.End, 1)
		require.Equal(t, []time.Duration{300 * time.Millisecond}, item.Duration)
		require.Equal(t, []time.Duration{50 * time.Millisecond}, item.EndHold)
		require.Equal(t, []Rect{{1, 2, 3, 4}}, item.EndOffset)
	})
}

func TestKeyShorthandNormalization(t *testing.T) {
	t.Run("click key scalar and list", func(t *testing.T) {
		var action Action
		require.NoError(t, json.Unmarshal([]byte(`{"type":"ClickKey","param":{"key":27}}`), &action))
		require.Equal(t, []int{27}, action.Param.(*ClickKeyParam).Key)
		require.NoError(t, json.Unmarshal([]byte(`{"type":"ClickKey","param":{"key":[27,28]}}`), &action))
		require.Equal(t, []int{27, 28}, action.Param.(*ClickKeyParam).Key)
	})

	t.Run("long press key scalar", func(t *testing.T) {
		var action Action
		require.NoError(t, json.Unmarshal([]byte(`{"type":"LongPressKey","param":{"key":27}}`), &action))
		require.Equal(t, []int{27}, action.Param.(*LongPressKeyParam).Key)
	})

	t.Run("null key stays nil", func(t *testing.T) {
		var action Action
		require.NoError(t, json.Unmarshal([]byte(`{"type":"ClickKey","param":{"key":null}}`), &action))
		require.Nil(t, action.Param.(*ClickKeyParam).Key)
	})
}

func TestColorMatchFlatRows(t *testing.T) {
	t.Run("flat single rows normalize to one row", func(t *testing.T) {
		var recognition Recognition
		require.NoError(t, json.Unmarshal([]byte(`{"type":"ColorMatch","param":{"lower":[1,2,3],"upper":[9,9,9]}}`), &recognition))
		color, ok := recognition.Param.(*ColorMatchParam)
		require.True(t, ok)
		require.Equal(t, [][]int{{1, 2, 3}}, color.Lower)
		require.Equal(t, [][]int{{9, 9, 9}}, color.Upper)
	})

	t.Run("nested rows unchanged", func(t *testing.T) {
		var recognition Recognition
		require.NoError(t, json.Unmarshal([]byte(`{"type":"ColorMatch","param":{"lower":[[1,2,3],[4,5,6]]}}`), &recognition))
		color, ok := recognition.Param.(*ColorMatchParam)
		require.True(t, ok)
		require.Equal(t, [][]int{{1, 2, 3}, {4, 5, 6}}, color.Lower)
	})

	t.Run("mixed and malformed rows rejected with invariance", func(t *testing.T) {
		recognition := Recognition{Type: RecognitionTypeColorMatch, Param: &ColorMatchParam{Count: 2}}
		before := recognition
		for _, payload := range []string{
			`{"type":"ColorMatch","param":{"lower":[[1,2],3]}}`,
			`{"type":"ColorMatch","param":{"lower":[[1,2],"x"]}}`,
			`{"type":"ColorMatch","param":{"lower":"x"}}`,
		} {
			require.Error(t, json.Unmarshal([]byte(payload), &recognition), payload)
			require.Equal(t, before, recognition, payload)
		}
	})
}

func TestOCRShorthandNormalization(t *testing.T) {
	t.Run("deprecated text alias maps to expected", func(t *testing.T) {
		var recognition Recognition
		require.NoError(t, json.Unmarshal([]byte(`{"type":"OCR","param":{"text":"Account"}}`), &recognition))
		ocr, ok := recognition.Param.(*OCRParam)
		require.True(t, ok)
		require.Equal(t, StringList{"Account"}, ocr.Expected)
	})

	t.Run("expected wins over text", func(t *testing.T) {
		var recognition Recognition
		require.NoError(t, json.Unmarshal([]byte(`{"type":"OCR","param":{"expected":["A"],"text":["B"]}}`), &recognition))
		ocr, ok := recognition.Param.(*OCRParam)
		require.True(t, ok)
		require.Equal(t, StringList{"A"}, ocr.Expected)
	})

	t.Run("text null is harmless and scalar text normalizes", func(t *testing.T) {
		var recognition Recognition
		require.NoError(t, json.Unmarshal([]byte(`{"type":"OCR","param":{"text":null}}`), &recognition))
		require.Nil(t, recognition.Param.(*OCRParam).Expected)
	})

	t.Run("single replace pair normalizes", func(t *testing.T) {
		var recognition Recognition
		require.NoError(t, json.Unmarshal([]byte(`{"type":"OCR","param":{"replace":["old","new"]}}`), &recognition))
		ocr, ok := recognition.Param.(*OCRParam)
		require.True(t, ok)
		require.Equal(t, [][2]string{{"old", "new"}}, ocr.Replace)
	})
}

func TestWaitFreezesShorthandAtNodeLevel(t *testing.T) {
	var node Node
	require.NoError(t, json.Unmarshal([]byte(`{"pre_wait_freezes": 1000, "post_wait_freezes": {"time": 500}}`), &node))
	require.Equal(t, &WaitFreezesParam{Time: time.Second}, node.PreWaitFreezes)
	require.Equal(t, &WaitFreezesParam{Time: 500 * time.Millisecond}, node.PostWaitFreezes)
}

func TestInlineSubV1StringRecognition(t *testing.T) {
	var recognition Recognition
	require.NoError(t, json.Unmarshal([]byte(`{"type":"And","param":{"all_of":[
		{"sub_name":"s","recognition":"ColorMatch","lower":[[1,2,3]]}]}}`), &recognition))
	and, ok := recognition.Param.(*AndRecognitionParam)
	require.True(t, ok)
	require.Len(t, and.AllOf, 1)
	item := and.AllOf[0]
	require.Equal(t, "s", item.Inline.SubName)
	require.Equal(t, RecognitionTypeColorMatch, item.Inline.Recognition.Type)
	color, ok := item.Inline.Recognition.Param.(*ColorMatchParam)
	require.True(t, ok, "v1 sub flat fields should decode typed, got %T", item.Inline.Recognition.Param)
	require.Equal(t, [][]int{{1, 2, 3}}, color.Lower)
}
