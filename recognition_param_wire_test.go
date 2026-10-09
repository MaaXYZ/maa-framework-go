package maa

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// The tests in this file pin the wire contracts of the recognition param
// fixes independently of the assertions updated inside the fix commits, so
// reverting a fix turns them red.

func TestRecognitionParams_ROIOffsetOmission(t *testing.T) {
	for name, param := range map[string]any{
		"TemplateMatch": TemplateMatchParam{},
		"FeatureMatch":  FeatureMatchParam{},
		"ColorMatch":    ColorMatchParam{},
		"OCR":           OCRParam{},
		"NNClassify":    NeuralNetworkClassifyParam{},
		"NNDetect":      NeuralNetworkDetectParam{},
		"Custom":        CustomRecognitionParam{},
		"DirectHit":     DirectHitParam{},
	} {
		t.Run(name+" zero offset omitted", func(t *testing.T) {
			data, err := json.Marshal(param)
			require.NoError(t, err)
			require.NotContains(t, string(data), "roi_offset",
				"zero roi_offset must encode as absent so parent-node inheritance applies, got %s", data)
		})
	}
	t.Run("nonzero offset encoded", func(t *testing.T) {
		data, err := json.Marshal(TemplateMatchParam{ROIOffset: Rect{1, 2, 3, 4}})
		require.NoError(t, err)
		require.Contains(t, string(data), `"roi_offset":[1,2,3,4]`)
	})
}

func TestTemplateMatchParam_TemplateThresholdJSON(t *testing.T) {
	t.Run("scalar template normalizes", func(t *testing.T) {
		var param TemplateMatchParam
		require.NoError(t, json.Unmarshal([]byte(`{"template":"a.png"}`), &param))
		require.Equal(t, StringList{"a.png"}, param.Template)
	})
	t.Run("scalar threshold normalizes", func(t *testing.T) {
		var param TemplateMatchParam
		require.NoError(t, json.Unmarshal([]byte(`{"threshold":0.7}`), &param))
		require.Equal(t, []float64{0.7}, param.Threshold)
	})
	t.Run("null template rejected", func(t *testing.T) {
		param := TemplateMatchParam{Template: StringList{"keep.png"}}
		require.Error(t, json.Unmarshal([]byte(`{"template":null}`), &param))
		require.Equal(t, StringList{"keep.png"}, param.Template)
	})
	t.Run("null threshold rejected", func(t *testing.T) {
		param := TemplateMatchParam{Threshold: []float64{0.5}}
		require.Error(t, json.Unmarshal([]byte(`{"threshold":null}`), &param))
		require.Equal(t, []float64{0.5}, param.Threshold)
	})
	t.Run("nil lists omitted to inherit", func(t *testing.T) {
		data, err := json.Marshal(TemplateMatchParam{})
		require.NoError(t, err)
		require.JSONEq(t, `{}`, string(data))
	})
	t.Run("empty non-nil lists encode as clear", func(t *testing.T) {
		data, err := json.Marshal(TemplateMatchParam{Template: StringList{}, Threshold: []float64{}})
		require.NoError(t, err)
		require.JSONEq(t, `{"template":[],"threshold":[]}`, string(data))
	})
	t.Run("feature match scalar template", func(t *testing.T) {
		var param FeatureMatchParam
		require.NoError(t, json.Unmarshal([]byte(`{"template":"a.png"}`), &param))
		require.Equal(t, StringList{"a.png"}, param.Template)
	})
	t.Run("failure invariance keeps all fields", func(t *testing.T) {
		seeded := TemplateMatchParam{
			ROI:       NewTargetString("NodeA"),
			ROIOffset: Rect{1, 2, 3, 4},
			Template:  StringList{"old.png"},
			Threshold: []float64{0.7},
			OrderBy:   TemplateMatchOrderByScore,
			Index:     2,
			Method:    TemplateMatchMethodCCORR_NORMED,
			GreenMask: true,
		}
		before := seeded
		require.Error(t, json.Unmarshal([]byte(`{"template":null}`), &seeded))
		require.Equal(t, before, seeded)
		require.Error(t, json.Unmarshal([]byte(`{"threshold":null}`), &seeded))
		require.Equal(t, before, seeded)
	})
}
