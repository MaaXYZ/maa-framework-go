package maa

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPipelineV2UnknownActionParamRoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		param string
	}{
		{
			name:  "object with large integer",
			param: `{"value":9007199254740993,"nested":{"enabled":true}}`,
		},
		{
			name:  "array",
			param: `[1,"two",{"three":3}]`,
		},
		{
			name:  "string",
			param: `"opaque parameter"`,
		},
		{
			name:  "null",
			param: `null`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := []byte(fmt.Sprintf(`{"type":"FutureAction","param":%s}`, tc.param))
			var action Action
			require.NoError(t, unmarshalJSON(input, &action), "decode unknown action")
			require.Equal(t, ActionType("FutureAction"), action.Type)
			raw, ok := action.Param.(*RawActionParam)
			require.True(t, ok, "unknown action parameter should use RawActionParam, got %T", action.Param)
			require.Equal(t, tc.param, string(*raw), "raw JSON should preserve the parameter data")

			encoded, err := marshalJSON(&action)
			require.NoError(t, err)
			require.JSONEq(t, string(input), string(encoded), "unknown action parameter should not be double encoded")

			standardEncoded, err := json.Marshal(&action)
			require.NoError(t, err)
			require.JSONEq(t, string(input), string(standardEncoded), "standard encoding/json must preserve pipeline v2 shape")
		})
	}
}

func TestPipelineV2UnknownRecognitionParamRoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		param string
	}{
		{
			name:  "object with large integer",
			param: `{"value":9007199254740993,"escaped":"\\u2028"}`,
		},
		{
			name:  "array",
			param: `[false,null,{"value":-7}]`,
		},
		{
			name:  "string",
			param: `"opaque recognition parameter"`,
		},
		{
			name:  "null",
			param: `null`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := []byte(fmt.Sprintf(`{"type":"FutureRecognition","param":%s}`, tc.param))
			var recognition Recognition
			require.NoError(t, unmarshalJSON(input, &recognition), "decode unknown recognition")
			require.Equal(t, RecognitionType("FutureRecognition"), recognition.Type)
			raw, ok := recognition.Param.(*RawRecognitionParam)
			require.True(t, ok, "unknown recognition parameter should use RawRecognitionParam, got %T", recognition.Param)
			require.Equal(t, tc.param, string(*raw), "raw JSON should preserve the parameter data")

			encoded, err := marshalJSON(&recognition)
			require.NoError(t, err)
			require.JSONEq(t, string(input), string(encoded), "unknown recognition parameter should not be double encoded")

			standardEncoded, err := json.Marshal(&recognition)
			require.NoError(t, err)
			require.JSONEq(t, string(input), string(standardEncoded), "standard encoding/json must preserve pipeline v2 shape")
		})
	}
}

func TestPipelineV2UnknownParamMissingAndNullSemantics(t *testing.T) {
	t.Run("missing action parameter remains nil", func(t *testing.T) {
		var action Action
		require.NoError(t, unmarshalJSON([]byte(`{"type":"FutureAction"}`), &action))
		require.Equal(t, ActionType("FutureAction"), action.Type)
		require.Nil(t, action.Param)
	})

	t.Run("explicit null action parameter is retained", func(t *testing.T) {
		var action Action
		require.NoError(t, unmarshalJSON([]byte(`{"type":"FutureAction","param":null}`), &action))
		raw, ok := action.Param.(*RawActionParam)
		require.True(t, ok, "explicit null should use RawActionParam, got %T", action.Param)
		require.Equal(t, []byte("null"), []byte(*raw))
	})

	t.Run("missing recognition parameter remains nil", func(t *testing.T) {
		var recognition Recognition
		require.NoError(t, unmarshalJSON([]byte(`{"type":"FutureRecognition"}`), &recognition))
		require.Equal(t, RecognitionType("FutureRecognition"), recognition.Type)
		require.Nil(t, recognition.Param)
	})

	t.Run("explicit null recognition parameter is retained", func(t *testing.T) {
		var recognition Recognition
		require.NoError(t, unmarshalJSON([]byte(`{"type":"FutureRecognition","param":null}`), &recognition))
		raw, ok := recognition.Param.(*RawRecognitionParam)
		require.True(t, ok, "explicit null should use RawRecognitionParam, got %T", recognition.Param)
		require.Equal(t, []byte("null"), []byte(*raw))
	})
}

func TestPipelineV2UnknownParamDecodeErrorsAndDestinationReuse(t *testing.T) {
	t.Run("missing known action parameter clears an existing parameter", func(t *testing.T) {
		action := Action{Type: ActionTypeClick, Param: &ClickParam{Contact: 7}}
		require.NoError(t, unmarshalJSON([]byte(`{"type":"Click"}`), &action))
		require.Equal(t, ActionTypeClick, action.Type)
		require.Nil(t, action.Param)
	})

	t.Run("null known action parameter clears an existing parameter", func(t *testing.T) {
		action := Action{Type: ActionTypeClick, Param: &ClickParam{Contact: 7}}
		require.NoError(t, unmarshalJSON([]byte(`{"type":"Click","param":null}`), &action))
		require.Equal(t, ActionTypeClick, action.Type)
		require.Nil(t, action.Param)
	})

	t.Run("missing known recognition parameter clears an existing parameter", func(t *testing.T) {
		recognition := Recognition{Type: RecognitionTypeTemplateMatch, Param: &TemplateMatchParam{Template: []string{"old.png"}}}
		require.NoError(t, unmarshalJSON([]byte(`{"type":"TemplateMatch"}`), &recognition))
		require.Equal(t, RecognitionTypeTemplateMatch, recognition.Type)
		require.Nil(t, recognition.Param)
	})

	t.Run("null known recognition parameter clears an existing parameter", func(t *testing.T) {
		recognition := Recognition{Type: RecognitionTypeTemplateMatch, Param: &TemplateMatchParam{Template: []string{"old.png"}}}
		require.NoError(t, unmarshalJSON([]byte(`{"type":"TemplateMatch","param":null}`), &recognition))
		require.Equal(t, RecognitionTypeTemplateMatch, recognition.Type)
		require.Nil(t, recognition.Param)
	})

	t.Run("missing parameter clears an existing action parameter", func(t *testing.T) {
		action := Action{
			Type:  ActionTypeClick,
			Param: &ClickParam{Contact: 7},
		}
		require.NoError(t, unmarshalJSON([]byte(`{"type":"FutureAction"}`), &action))
		require.Equal(t, ActionType("FutureAction"), action.Type)
		require.Nil(t, action.Param)
	})

	t.Run("null parameter replaces an existing recognition parameter", func(t *testing.T) {
		recognition := Recognition{
			Type:  RecognitionTypeDirectHit,
			Param: &DirectHitParam{},
		}
		require.NoError(t, unmarshalJSON([]byte(`{"type":"FutureRecognition","param":null}`), &recognition))
		raw, ok := recognition.Param.(*RawRecognitionParam)
		require.True(t, ok, "explicit null should replace the old parameter, got %T", recognition.Param)
		require.Equal(t, []byte("null"), []byte(*raw))
	})

	t.Run("invalid unknown parameter leaves action unchanged", func(t *testing.T) {
		action := Action{Type: ActionTypeClick, Param: &ClickParam{Contact: 7}}
		beforeParam := *action.Param.(*ClickParam)
		err := unmarshalJSON([]byte(`{"type":"FutureAction","param":{"broken":}}`), &action)
		require.Error(t, err)
		require.Equal(t, ActionTypeClick, action.Type)
		gotParam, ok := action.Param.(*ClickParam)
		require.True(t, ok)
		require.Equal(t, beforeParam, *gotParam)
	})

	t.Run("invalid unknown parameter leaves recognition unchanged", func(t *testing.T) {
		recognition := Recognition{Type: RecognitionTypeDirectHit, Param: &DirectHitParam{}}
		err := unmarshalJSON([]byte(`{"type":"FutureRecognition","param":{"broken":}}`), &recognition)
		require.Error(t, err)
		require.Equal(t, RecognitionTypeDirectHit, recognition.Type)
		_, ok := recognition.Param.(*DirectHitParam)
		require.True(t, ok)
	})

	t.Run("invalid JSON is rejected", func(t *testing.T) {
		var action Action
		require.Error(t, unmarshalJSON([]byte(`{"type":"FutureAction","param":`), &action))
		var recognition Recognition
		require.Error(t, unmarshalJSON([]byte(`{"type":"FutureRecognition","param":`), &recognition))
	})
}

func TestPipelineV2KnownParamsRemainTypedAndValidate(t *testing.T) {
	t.Run("known action", func(t *testing.T) {
		var action Action
		require.NoError(t, unmarshalJSON([]byte(`{"type":"Click","param":{"contact":7}}`), &action))
		param, ok := action.Param.(*ClickParam)
		require.True(t, ok, "known action should remain typed, got %T", action.Param)
		require.Equal(t, 7, param.Contact)

		beforeParam := *param
		err := unmarshalJSON([]byte(`{"type":"LongPress","param":{"duration":"bad"}}`), &action)
		require.Error(t, err, "malformed known fields must fail")
		require.Equal(t, ActionTypeClick, action.Type)
		gotParam, ok := action.Param.(*ClickParam)
		require.True(t, ok)
		require.Equal(t, beforeParam, *gotParam)
	})

	t.Run("known recognition", func(t *testing.T) {
		var recognition Recognition
		require.NoError(t, unmarshalJSON([]byte(`{"type":"TemplateMatch","param":{"template":["icon.png"]}}`), &recognition))
		param, ok := recognition.Param.(*TemplateMatchParam)
		require.True(t, ok, "known recognition should remain typed, got %T", recognition.Param)
		require.Equal(t, []string{"icon.png"}, param.Template)

		beforeParam := *param
		err := unmarshalJSON([]byte(`{"type":"ColorMatch","param":{"lower":["bad"]}}`), &recognition)
		require.Error(t, err, "malformed known fields must fail")
		require.Equal(t, RecognitionTypeTemplateMatch, recognition.Type)
		gotParam, ok := recognition.Param.(*TemplateMatchParam)
		require.True(t, ok)
		require.Equal(t, beforeParam, *gotParam)
	})
}

func TestPipelineV2UnknownNestedRecognitionRoundTrip(t *testing.T) {
	nestedUnknown := []byte(`{"type":"And","param":{"all_of":[{"sub_name":"future","recognition":{"type":"FutureRecognition","param":{"value":9007199254740993}}}]}}`)
	var decoded Recognition
	require.NoError(t, unmarshalJSON(nestedUnknown, &decoded))
	nestedAnd := decoded.Param.(*AndRecognitionParam)
	nestedRaw, ok := nestedAnd.AllOf[0].Inline.Param.(*RawRecognitionParam)
	require.True(t, ok, "unknown nested recognition should use RawRecognitionParam, got %T", nestedAnd.AllOf[0].Inline.Param)
	require.Equal(t, `{"value":9007199254740993}`, string(*nestedRaw))
	nestedRoundTrip, err := marshalJSON(&decoded)
	require.NoError(t, err)
	require.JSONEq(t, string(nestedUnknown), string(nestedRoundTrip))
}

func TestRawPipelineParamsMarshalAndUnmarshal(t *testing.T) {
	t.Run("action", func(t *testing.T) {
		valid := RawActionParam(`{"value":9007199254740993}`)
		encoded, err := json.Marshal(valid)
		require.NoError(t, err)
		require.Equal(t, string(valid), string(encoded))

		invalid := RawActionParam(`{"broken":}`)
		_, err = json.Marshal(invalid)
		require.Error(t, err, "invalid raw JSON must be rejected by MarshalJSON")

		input := []byte(`{"value":9007199254740993}`)
		var decoded RawActionParam
		require.NoError(t, json.Unmarshal(input, &decoded))
		require.Equal(t, string(input), string(decoded), "large integer must remain uninterpreted")
		input[0] = '['
		require.Equal(t, `{"value":9007199254740993}`, string(decoded), "UnmarshalJSON must copy its input")

		decoded = RawActionParam(`{"old":1}`)
		require.Error(t, json.Unmarshal([]byte(`{"broken":}`), &decoded))
		require.Equal(t, `{"old":1}`, string(decoded), "failed UnmarshalJSON must preserve its destination")
	})

	t.Run("recognition", func(t *testing.T) {
		valid := RawRecognitionParam(`{"value":9007199254740993}`)
		encoded, err := json.Marshal(valid)
		require.NoError(t, err)
		require.Equal(t, string(valid), string(encoded))

		invalid := RawRecognitionParam(`{"broken":}`)
		_, err = json.Marshal(invalid)
		require.Error(t, err, "invalid raw JSON must be rejected by MarshalJSON")

		input := []byte(`{"value":9007199254740993}`)
		var decoded RawRecognitionParam
		require.NoError(t, json.Unmarshal(input, &decoded))
		require.Equal(t, string(input), string(decoded), "large integer must remain uninterpreted")
		input[0] = '['
		require.Equal(t, `{"value":9007199254740993}`, string(decoded), "UnmarshalJSON must copy its input")

		decoded = RawRecognitionParam(`{"old":1}`)
		require.Error(t, json.Unmarshal([]byte(`{"broken":}`), &decoded))
		require.Equal(t, `{"old":1}`, string(decoded), "failed UnmarshalJSON must preserve its destination")
	})
}

func TestDecodeActionParamUnknownAndKnown(t *testing.T) {
	unknown, err := decodeActionParam(ActionType("FutureAction"), []byte(`{"value":9007199254740993}`))
	require.NoError(t, err)
	unknownRaw, ok := unknown.(*RawActionParam)
	require.True(t, ok, "unknown helper result should be RawActionParam, got %T", unknown)
	require.Equal(t, `{"value":9007199254740993}`, string(*unknownRaw))

	known, err := decodeActionParam(ActionTypeClick, []byte(`{"contact":7}`))
	require.NoError(t, err)
	knownClick, ok := known.(*ClickParam)
	require.True(t, ok, "known helper result should remain typed, got %T", known)
	require.Equal(t, 7, knownClick.Contact)

	_, err = decodeActionParam(ActionTypeClick, []byte(`{"contact":"bad"}`))
	require.Error(t, err, "known helper errors must not fall back to RawActionParam")

	missing, err := decodeActionParam(ActionType("FutureAction"), nil)
	require.NoError(t, err)
	require.Nil(t, missing)

	knownNull, err := decodeActionParam(ActionTypeClick, []byte("null"))
	require.NoError(t, err)
	require.Nil(t, knownNull)

	unknownNull, err := decodeActionParam(ActionType("FutureAction"), []byte("null"))
	require.NoError(t, err)
	unknownNullRaw, ok := unknownNull.(*RawActionParam)
	require.True(t, ok)
	require.Equal(t, []byte("null"), []byte(*unknownNullRaw))
}

func TestDecodeRecognitionParamUnknownAndKnown(t *testing.T) {
	unknown, err := decodeRecognitionParam(RecognitionType("FutureRecognition"), []byte(`{"value":9007199254740993}`))
	require.NoError(t, err)
	unknownRaw, ok := unknown.(*RawRecognitionParam)
	require.True(t, ok, "unknown helper result should be RawRecognitionParam, got %T", unknown)
	require.Equal(t, `{"value":9007199254740993}`, string(*unknownRaw))

	known, err := decodeRecognitionParam(RecognitionTypeTemplateMatch, []byte(`{"template":["icon.png"]}`))
	require.NoError(t, err)
	knownTemplate, ok := known.(*TemplateMatchParam)
	require.True(t, ok, "known helper result should remain typed, got %T", known)
	require.Equal(t, []string{"icon.png"}, knownTemplate.Template)

	_, err = decodeRecognitionParam(RecognitionTypeTemplateMatch, []byte(`{"threshold":["bad"]}`))
	require.Error(t, err, "known helper errors must not fall back to RawRecognitionParam")

	missing, err := decodeRecognitionParam(RecognitionType("FutureRecognition"), nil)
	require.NoError(t, err)
	require.Nil(t, missing)

	knownNull, err := decodeRecognitionParam(RecognitionTypeTemplateMatch, []byte("null"))
	require.NoError(t, err)
	require.Nil(t, knownNull)

	unknownNull, err := decodeRecognitionParam(RecognitionType("FutureRecognition"), []byte("null"))
	require.NoError(t, err)
	unknownNullRaw, ok := unknownNull.(*RawRecognitionParam)
	require.True(t, ok)
	require.Equal(t, []byte("null"), []byte(*unknownNullRaw))
}
