package maa

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClassSelectors_JSON(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  ClassSelectors
		json  string
	}{
		{`0`, ClassSelectors{ClassIndex(0)}, `[0]`},
		{`-1`, ClassSelectors{ClassIndex(-1)}, `[-1]`},
		{`"Cat"`, ClassSelectors{ClassLabel("Cat")}, `["Cat"]`},
		{`""`, ClassSelectors{ClassLabel("")}, `[""]`},
		{`"0"`, ClassSelectors{ClassLabel("0")}, `["0"]`},
		{`[0,2]`, ClassSelectors{ClassIndex(0), ClassIndex(2)}, `[0,2]`},
		{`["猫","Mouse"]`, ClassSelectors{ClassLabel("猫"), ClassLabel("Mouse")}, `["猫","Mouse"]`},
		{` [0,"Mouse",0] `, ClassSelectors{ClassIndex(0), ClassLabel("Mouse"), ClassIndex(0)}, `[0,"Mouse",0]`},
		{`[]`, ClassSelectors{}, `[]`},
	} {
		t.Run(tc.input, func(t *testing.T) {
			var got ClassSelectors
			require.NoError(t, json.Unmarshal([]byte(tc.input), &got))
			require.Equal(t, tc.want, got)
			data, err := json.Marshal(got)
			require.NoError(t, err)
			require.JSONEq(t, tc.json, string(data))
		})
	}
}

func TestClassSelectors_RejectInvalidJSON(t *testing.T) {
	for _, input := range []string{
		`null`, `true`, `false`, `1.5`, `1.0`, `{}`, `{"index":0}`,
		`[null]`, `[0,true]`, `["Cat",1.5]`, `[[0]]`, `[{}]`, `[0,`,
		`999999999999999999999999999999999`,
	} {
		t.Run(input, func(t *testing.T) {
			got := ClassSelectors{ClassLabel("original")}
			require.Error(t, json.Unmarshal([]byte(input), &got))
			require.Equal(t, ClassSelectors{ClassLabel("original")}, got)
		})
	}
}

func TestClassSelector_Accessors(t *testing.T) {
	var zero ClassSelector
	for _, selector := range []ClassSelector{zero, ClassIndex(2)} {
		require.True(t, selector.IsIndex())
		require.False(t, selector.IsLabel())
		index, err := selector.AsIndex()
		require.NoError(t, err)
		require.Equal(t, selector, ClassIndex(index))
		_, err = selector.AsLabel()
		require.Error(t, err)
	}
	label := ClassLabel("Cat")
	require.True(t, label.IsLabel())
	require.False(t, label.IsIndex())
	got, err := label.AsLabel()
	require.NoError(t, err)
	require.Equal(t, "Cat", got)
	_, err = label.AsIndex()
	require.Error(t, err)
	// A failed decode must not replace a previously valid value.
	require.Error(t, json.Unmarshal([]byte(`null`), &label))
	require.Equal(t, ClassLabel("Cat"), label)
	require.NoError(t, json.Unmarshal([]byte(`2`), &label))
	require.Equal(t, ClassIndex(2), label)
	require.NoError(t, json.Unmarshal([]byte(`"Dog"`), &label))
	require.Equal(t, ClassLabel("Dog"), label)
}

func neuralExpected(t *testing.T, rec *Recognition) ClassSelectors {
	t.Helper()
	switch p := rec.Param.(type) {
	case *NeuralNetworkClassifyParam:
		return p.Expected
	case *NeuralNetworkDetectParam:
		return p.Expected
	default:
		t.Fatalf("unexpected recognition parameter type %T", rec.Param)
		return nil
	}
}

func neuralRecognition(tp RecognitionType, expected ClassSelectors) *Recognition {
	if tp == RecognitionTypeNeuralNetworkClassify {
		return RecNeuralNetworkClassify(NeuralNetworkClassifyParam{Model: "model.onnx", Expected: expected})
	}
	return RecNeuralNetworkDetect(NeuralNetworkDetectParam{Model: "model.onnx", Expected: expected})
}

func TestNeuralNetworkExpected_JSON(t *testing.T) {
	for _, tp := range []RecognitionType{RecognitionTypeNeuralNetworkClassify, RecognitionTypeNeuralNetworkDetect} {
		for _, raw := range []string{"", `[]`, `0`, `"Cat"`, `[0,"Mouse"]`} {
			t.Run(string(tp)+"/"+raw, func(t *testing.T) {
				param := `"model":"model.onnx"`
				if raw != "" {
					param += `,"expected":` + raw
				}
				var rec Recognition
				require.NoError(t, json.Unmarshal([]byte(fmt.Sprintf(`{"type":%q,"param":{%s}}`, tp, param)), &rec))
				expected := neuralExpected(t, &rec)
				if raw == "" {
					require.Nil(t, expected)
				} else {
					require.NotNil(t, expected)
				}
				data, err := json.Marshal(rec)
				require.NoError(t, err)
				var output struct {
					Param map[string]json.RawMessage `json:"param"`
				}
				require.NoError(t, json.Unmarshal(data, &output))
				if raw == "" {
					require.NotContains(t, output.Param, "expected")
				} else {
					want, err := json.Marshal(expected)
					require.NoError(t, err)
					require.JSONEq(t, string(want), string(output.Param["expected"]))
				}
				var roundTrip Recognition
				require.NoError(t, json.Unmarshal(data, &roundTrip))
				require.Equal(t, expected, neuralExpected(t, &roundTrip))
			})
		}
	}
}

func TestNeuralNetworkExpected_ConstructorCopiesList(t *testing.T) {
	for _, tp := range []RecognitionType{RecognitionTypeNeuralNetworkClassify, RecognitionTypeNeuralNetworkDetect} {
		t.Run(string(tp), func(t *testing.T) {
			input := ClassSelectors{ClassIndex(0), ClassLabel("Mouse")}
			rec := neuralRecognition(tp, input)
			input[0] = ClassIndex(2)
			require.Equal(t, ClassSelectors{ClassIndex(0), ClassLabel("Mouse")}, neuralExpected(t, rec))
			require.Nil(t, neuralExpected(t, neuralRecognition(tp, nil)))
			require.Equal(t, ClassSelectors{}, neuralExpected(t, neuralRecognition(tp, ClassSelectors{})))
		})
	}
}

func TestResource_OverridePipeline_NeuralNetworkExpected(t *testing.T) {
	for _, tp := range []RecognitionType{RecognitionTypeNeuralNetworkClassify, RecognitionTypeNeuralNetworkDetect} {
		t.Run(string(tp), func(t *testing.T) {
			res := createResource(t)
			t.Cleanup(func() { require.NoError(t, res.Destroy()) })
			mixed := ClassSelectors{ClassIndex(0), ClassLabel("Mouse")}
			for _, step := range []struct {
				name        string
				input, want ClassSelectors
			}{
				{"set mixed", mixed, mixed},
				{"inherit", nil, mixed},
				{"clear", ClassSelectors{}, ClassSelectors{}},
			} {
				require.NoError(t, res.OverridePipeline(map[string]*Node{
					"Expected": {Recognition: neuralRecognition(tp, step.input)},
				}), step.name)
				node, err := res.GetNode("Expected")
				require.NoError(t, err, step.name)
				require.NotNil(t, node.Recognition)
				require.Equal(t, step.want, neuralExpected(t, node.Recognition), step.name)
			}
		})
	}
}

func TestClassSelectors_UsesConfiguredJSONCodec(t *testing.T) {
	resetJSONCodecForTest(t)
	encoded, decoded := false, false
	SetJSONEncoder(func(v any) ([]byte, error) {
		encoded = true
		return json.Marshal(v)
	})
	SetJSONDecoder(func(data []byte, v any) error {
		decoded = true
		return json.Unmarshal(data, v)
	})
	var selectors ClassSelectors
	require.NoError(t, json.Unmarshal([]byte(`[0,"Mouse"]`), &selectors))
	require.True(t, decoded)
	data, err := json.Marshal(selectors)
	require.NoError(t, err)
	require.True(t, encoded)
	require.JSONEq(t, `[0,"Mouse"]`, string(data))
}
