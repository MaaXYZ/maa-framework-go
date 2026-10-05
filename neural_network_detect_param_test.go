package maa

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNeuralNetworkDetectParam_ThresholdJSON(t *testing.T) {
	for _, tc := range []struct {
		name      string
		field     string
		want      []float64
		canonical string
	}{
		{name: "missing"},
		{name: "scalar", field: `,"threshold":0.5`, want: []float64{0.5}, canonical: `[0.5]`},
		{name: "scalar zero", field: `,"threshold":0`, want: []float64{0}, canonical: `[0]`},
		{name: "array", field: `,"threshold":[0.5,0]`, want: []float64{0.5, 0}, canonical: `[0.5,0]`},
		{name: "empty array", field: `,"threshold":[]`, want: []float64{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := `{"type":"NeuralNetworkDetect","param":{"model":"detector.onnx"` + tc.field + `}}`
			var recognition Recognition
			require.NoError(t, json.Unmarshal([]byte(input), &recognition))
			param := recognition.Param.(*NeuralNetworkDetectParam)
			require.Equal(t, tc.want, param.Threshold)

			encoded, err := json.Marshal(recognition)
			require.NoError(t, err)
			var wire struct {
				Param struct {
					Threshold json.RawMessage `json:"threshold"`
				} `json:"param"`
			}
			require.NoError(t, json.Unmarshal(encoded, &wire))
			if tc.canonical == "" {
				require.Nil(t, wire.Param.Threshold, "empty thresholds should inherit")
			} else {
				require.JSONEq(t, tc.canonical, string(wire.Param.Threshold))
			}
		})
	}
}

func TestNeuralNetworkDetectParam_AllFieldsRoundTrip(t *testing.T) {
	const input = `{"roi":[1,2,30,40],"roi_offset":[5,6,7,8],"labels":["cat","dog"],"model":"detector.onnx","expected":[0,"dog"],"threshold":0.5,"order_by":"Score","index":1}`
	var param NeuralNetworkDetectParam
	require.NoError(t, json.Unmarshal([]byte(input), &param))
	require.Equal(t, NeuralNetworkDetectParam{
		ROI:       NewTargetRect(Rect{1, 2, 30, 40}),
		ROIOffset: Rect{5, 6, 7, 8},
		Labels:    []string{"cat", "dog"},
		Model:     "detector.onnx",
		Expected:  ClassSelectors{ClassIndex(0), ClassLabel("dog")},
		Threshold: []float64{0.5},
		OrderBy:   NeuralNetworkDetectOrderByScore,
		Index:     1,
	}, param)

	encoded, err := json.Marshal(param)
	require.NoError(t, err)
	require.JSONEq(t, `{"roi":[1,2,30,40],"roi_offset":[5,6,7,8],"labels":["cat","dog"],"model":"detector.onnx","expected":[0,"dog"],"threshold":[0.5],"order_by":"Score","index":1}`, string(encoded))
}

func TestNeuralNetworkDetectParam_DestinationReuse(t *testing.T) {
	param := NeuralNetworkDetectParam{
		Model:     "detector.onnx",
		Labels:    []string{"cat"},
		Expected:  ClassSelectors{ClassLabel("cat")},
		Threshold: []float64{0.25},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"index":1}`), &param))
	require.Equal(t, "detector.onnx", param.Model)
	require.Equal(t, StringList{"cat"}, param.Labels)
	require.Equal(t, ClassSelectors{ClassLabel("cat")}, param.Expected)
	require.Equal(t, []float64{0.25}, param.Threshold)
	require.Equal(t, 1, param.Index)

	require.NoError(t, json.Unmarshal([]byte(`{"threshold":0.5}`), &param))
	require.Equal(t, "detector.onnx", param.Model)
	require.Equal(t, []float64{0.5}, param.Threshold)
	before := param
	require.NoError(t, json.Unmarshal([]byte(`null`), &param))
	require.Equal(t, before, param)
}

func TestNeuralNetworkDetectParam_InvalidPreservesValue(t *testing.T) {
	for _, threshold := range []string{`null`, `"0.5"`, `true`, `{}`, `[null]`, `[0.5,"bad"]`, `1e400`} {
		t.Run(threshold, func(t *testing.T) {
			labels := []string{"old"}
			expected := ClassSelectors{ClassLabel("old")}
			thresholds := []float64{0.25}
			param := NeuralNetworkDetectParam{
				ROI:       NewTargetRect(Rect{1, 2, 3, 4}),
				ROIOffset: Rect{5, 6, 7, 8},
				Labels:    labels,
				Model:     "old.onnx",
				Expected:  expected,
				Threshold: thresholds,
				OrderBy:   NeuralNetworkDetectOrderByScore,
				Index:     1,
			}
			before := param
			input := `{"labels":["new"],"model":"new.onnx","expected":[0],"threshold":` + threshold + `}`
			require.Error(t, json.Unmarshal([]byte(input), &param))
			require.Equal(t, before, param)
			require.Equal(t, []string{"old"}, labels)
			require.Equal(t, ClassSelectors{ClassLabel("old")}, expected)
			require.Equal(t, []float64{0.25}, thresholds)
		})
	}

	t.Run("invalid labels do not mutate their backing array", func(t *testing.T) {
		labels := []string{"old", "second"}
		param := NeuralNetworkDetectParam{Labels: labels, Model: "old.onnx", Threshold: []float64{0.25}}
		before := param
		require.Error(t, json.Unmarshal([]byte(`{"labels":["new",1],"threshold":0.5}`), &param))
		require.Equal(t, before, param)
		require.Equal(t, []string{"old", "second"}, labels)
	})

	t.Run("invalid expected preserves every field", func(t *testing.T) {
		param := NeuralNetworkDetectParam{
			ROI:       NewTargetRect(Rect{1, 2, 3, 4}),
			ROIOffset: Rect{5, 6, 7, 8},
			Labels:    StringList{"cat"},
			Model:     "old.onnx",
			Expected:  ClassSelectors{ClassLabel("old")},
			Threshold: []float64{0.25},
			OrderBy:   NeuralNetworkDetectOrderByScore,
			Index:     1,
		}
		before := param
		for _, expected := range []string{`true`, `1.5`, `[[0]]`, `[{}]`, `null`, `2147483648`} {
			require.Error(t, json.Unmarshal([]byte(`{"labels":["new"],"model":"new.onnx","expected":`+expected+`,"threshold":0.5}`), &param))
			require.Equal(t, before, param)
		}
	})
}

func TestNeuralNetworkDetectParam_NativeScalarThreshold(t *testing.T) {
	res := createResource(t)
	t.Cleanup(func() { require.NoError(t, res.Destroy()) })
	var recognition Recognition
	require.NoError(t, json.Unmarshal([]byte(`{"type":"NeuralNetworkDetect","param":{"model":"detector.onnx","threshold":0.5}}`), &recognition))
	require.NoError(t, res.OverridePipeline(map[string]*Node{
		"DetectThreshold": NewNode("DetectThreshold").SetRecognition(&recognition),
	}))
	node, err := res.GetNode("DetectThreshold")
	require.NoError(t, err)
	require.Equal(t, []float64{0.5}, node.Recognition.Param.(*NeuralNetworkDetectParam).Threshold)
}
