package maa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type recognitionDetailRawList []map[string]any

func (l *recognitionDetailRawList) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*l = nil
		return nil
	}
	switch trimmed[0] {
	case '[':
		var items []map[string]any
		if err := json.Unmarshal(trimmed, &items); err != nil {
			return err
		}
		*l = items
		return nil
	case '{':
		var item map[string]any
		if err := json.Unmarshal(trimmed, &item); err != nil {
			return err
		}
		*l = []map[string]any{item}
		return nil
	default:
		return fmt.Errorf("invalid recognition detail list: %s", string(trimmed))
	}
}

type recognitionDetailRaw struct {
	All      recognitionDetailRawList `json:"all"`
	Best     recognitionDetailRawList `json:"best"`
	Filtered recognitionDetailRawList `json:"filtered"`
}

type testRecognitionDetailCustomRec struct{}

func (t *testRecognitionDetailCustomRec) Run(_ *Context, _ *CustomRecognitionArg) (*CustomRecognitionResult, bool) {
	return &CustomRecognitionResult{
		Box:    Rect{10, 20, 30, 40},
		Detail: "custom-recognition-detail",
	}, true
}

type recognitionDetailTestCase struct {
	name  string
	typ   NodeRecognitionType
	param NodeRecognitionParam
}

func recognitionDetailTestCases() []recognitionDetailTestCase {
	return []recognitionDetailTestCase{
		{
			name:  "direct_hit",
			typ:   NodeRecognitionTypeDirectHit,
			param: &NodeDirectHitParam{},
		},
		{
			name: "template_match",
			typ:  NodeRecognitionTypeTemplateMatch,
			param: &NodeTemplateMatchParam{
				Template:  []string{"Wilderness/EnterWilderness.png"},
				Threshold: []float64{0.01},
				OrderBy:   NodeTemplateMatchOrderByScore,
			},
		},
		{
			name: "feature_match",
			typ:  NodeRecognitionTypeFeatureMatch,
			param: &NodeFeatureMatchParam{
				Template: []string{"Wilderness/CollectTrust.png"},
				Count:    1,
				Detector: NodeFeatureMatchMethodORB,
				Ratio:    1.0,
			},
		},
		{
			name: "color_match",
			typ:  NodeRecognitionTypeColorMatch,
			param: &NodeColorMatchParam{
				Method:    NodeColorMatchMethodRGB,
				Lower:     [][]int{{0, 0, 0}},
				Upper:     [][]int{{255, 255, 255}},
				Count:     1,
				Connected: true,
			},
		},
		{
			name: "ocr",
			typ:  NodeRecognitionTypeOCR,
			param: &NodeOCRParam{
				Expected:  []string{".*"},
				Threshold: 0.0,
				OrderBy:   NodeOCROrderByLength,
			},
		},
		{
			name: "and",
			typ:  NodeRecognitionTypeAnd,
			param: &NodeAndRecognitionParam{
				AllOf: []NodeAndRecognitionItem{
					AndItem("template", RecTemplateMatch(
						[]string{"Wilderness/EnterWilderness.png"},
						WithTemplateMatchThreshold([]float64{0.01}),
						WithTemplateMatchOrderBy(NodeTemplateMatchOrderByScore),
					)),
					AndItem("color", RecColorMatch(
						[][]int{{0, 0, 0}},
						[][]int{{255, 255, 255}},
						WithColorMatchCount(1),
						WithColorMatchConnected(true),
					)),
				},
				BoxIndex: 0,
			},
		},
		{
			name: "or",
			typ:  NodeRecognitionTypeOr,
			param: &NodeOrRecognitionParam{
				AnyOf: []*NodeRecognition{
					RecTemplateMatch(
						[]string{"Wilderness/EnterWilderness.png"},
						WithTemplateMatchThreshold([]float64{0.01}),
						WithTemplateMatchOrderBy(NodeTemplateMatchOrderByScore),
					),
					RecColorMatch(
						[][]int{{0, 0, 0}},
						[][]int{{255, 255, 255}},
						WithColorMatchCount(1),
						WithColorMatchConnected(true),
					),
				},
			},
		},
		{
			name: "nn_classify",
			typ:  NodeRecognitionTypeNeuralNetworkClassify,
			param: &NodeNeuralNetworkClassifyParam{
				Labels:   []string{"cat", "dog", "mouse"},
				Model:    "classify/classifier.onnx",
				Expected: []int{0, 2},
				OrderBy:  NodeNeuralNetworkClassifyOrderByScore,
				Index:    0,
			},
		},
		{
			name: "nn_detect",
			typ:  NodeRecognitionTypeNeuralNetworkDetect,
			param: &NodeNeuralNetworkDetectParam{
				Model:    "ocr/det.onnx",
				Expected: []int{0},
				OrderBy:  NodeNeuralNetworkDetectOrderByArea,
				Index:    0,
			},
		},
		{
			name: "custom",
			typ:  NodeRecognitionTypeCustom,
			param: &NodeCustomRecognitionParam{
				CustomRecognition:      "TestRecognitionDetail_Custom",
				CustomRecognitionParam: map[string]any{"key": "value"},
			},
		},
	}
}

type testRecognitionDetailFromRecognitionAct struct {
	t *testing.T
}

func (a *testRecognitionDetailFromRecognitionAct) Run(ctx *Context, _ *CustomActionArg) bool {
	img := ctx.GetTasker().GetController().CacheImage()
	require.NotNil(a.t, img)

	assertions := map[string]func(t *testing.T, detail *RecognitionDetail){
		"direct_hit": func(t *testing.T, detail *RecognitionDetail) {
			require.Nil(t, detail.Results)
		},
		"template_match": func(t *testing.T, detail *RecognitionDetail) {
			if len(detail.Results.All) == 0 {
				return
			}
			val, ok := detail.Results.All[0].AsTemplateMatch()
			require.True(t, ok)
			require.NotNil(t, val)
		},
		"feature_match": func(t *testing.T, detail *RecognitionDetail) {
			if len(detail.Results.All) == 0 {
				return
			}
			val, ok := detail.Results.All[0].AsFeatureMatch()
			require.True(t, ok)
			require.NotNil(t, val)
		},
		"color_match": func(t *testing.T, detail *RecognitionDetail) {
			if len(detail.Results.All) == 0 {
				return
			}
			val, ok := detail.Results.All[0].AsColorMatch()
			require.True(t, ok)
			require.NotNil(t, val)
		},
		"ocr": func(t *testing.T, detail *RecognitionDetail) {
			if len(detail.Results.All) == 0 {
				return
			}
			val, ok := detail.Results.All[0].AsOCR()
			require.True(t, ok)
			require.NotNil(t, val)
		},
		"nn_classify": func(t *testing.T, detail *RecognitionDetail) {
			if len(detail.Results.All) == 0 {
				return
			}
			for _, item := range detail.Results.All {
				val, ok := item.AsNeuralNetworkClassify()
				require.True(t, ok)
				require.NotNil(t, val)
			}
		},
		"nn_detect": func(t *testing.T, detail *RecognitionDetail) {
			if len(detail.Results.All) == 0 {
				return
			}
			val, ok := detail.Results.All[0].AsNeuralNetworkDetect()
			require.True(t, ok)
			require.NotNil(t, val)
		},
		"custom": func(t *testing.T, detail *RecognitionDetail) {
			if len(detail.Results.All) == 0 {
				return
			}
			val, ok := detail.Results.All[0].AsCustom()
			require.True(t, ok)
			require.NotNil(t, val)
			require.Equal(t, "custom-recognition-detail", val.Detail)
		},
		"and": func(t *testing.T, detail *RecognitionDetail) {
			require.Nil(t, detail.Results)
			require.NotNil(t, detail.CombinedResult)
		},
		"or": func(t *testing.T, detail *RecognitionDetail) {
			require.Nil(t, detail.Results)
			require.NotNil(t, detail.CombinedResult)
		},
	}

	runRecognition := func(t *testing.T, name string, recoType NodeRecognitionType, param NodeRecognitionParam) {
		detail, err := ctx.RunRecognitionDirect(recoType, param, img)
		require.NoError(t, err)
		require.NotNil(t, detail)

		switch recoType {
		case NodeRecognitionTypeAnd, NodeRecognitionTypeOr:
			requireRecognitionDetailMatchesCombinedRaw(t, detail)
		case NodeRecognitionTypeDirectHit:
			require.Nil(t, detail.Results)
			requireRecognitionDetailMatchesRaw(t, detail)
		default:
			require.NotNil(t, detail.Results)
			requireRecognitionDetailMatchesRaw(t, detail)
		}

		assert, ok := assertions[name]
		require.True(t, ok, "missing assertion for %s", name)
		assert(t, detail)
	}

	for _, tc := range recognitionDetailTestCases() {
		tc := tc
		a.t.Run(tc.name, func(t *testing.T) {
			runRecognition(t, tc.name, tc.typ, tc.param)
		})
	}

	return true
}

func requireRecognitionResultMatchesRaw(t *testing.T, result *RecognitionResult, raw map[string]any) {
	t.Helper()
	require.NotNil(t, result)

	resultJSON, err := json.Marshal(result.Value())
	require.NoError(t, err)

	resultMap := map[string]any{}
	require.NoError(t, json.Unmarshal(resultJSON, &resultMap))

	for key, rawVal := range raw {
		resultVal, ok := resultMap[key]
		require.True(t, ok, "result missing key: %s", key)
		require.Equal(t, rawVal, resultVal)
	}
}

func requireRecognitionResultsMatchRaw(t *testing.T, results *RecognitionResults, raw recognitionDetailRaw) {
	t.Helper()
	require.NotNil(t, results)
	require.Len(t, results.All, len(raw.All))
	require.Len(t, results.Best, len(raw.Best))
	require.Len(t, results.Filtered, len(raw.Filtered))

	for i, item := range raw.All {
		requireRecognitionResultMatchesRaw(t, results.All[i], item)
	}
	for i, item := range raw.Best {
		requireRecognitionResultMatchesRaw(t, results.Best[i], item)
	}
	for i, item := range raw.Filtered {
		requireRecognitionResultMatchesRaw(t, results.Filtered[i], item)
	}
}

func requireRecognitionDetailMatchesRaw(t *testing.T, detail *RecognitionDetail) {
	t.Helper()
	require.NotNil(t, detail)
	if detail.Results == nil {
		require.True(t, detail.DetailJson == "" || detail.DetailJson == "{}" || detail.DetailJson == "null")
		return
	}

	raw := recognitionDetailRaw{}
	if detail.DetailJson != "" && detail.DetailJson != "{}" && detail.DetailJson != "null" {
		require.NoError(t, json.Unmarshal([]byte(detail.DetailJson), &raw))
	}
	requireRecognitionResultsMatchRaw(t, detail.Results, raw)
}

func requireRecognitionDetailMatchesCombinedRaw(t *testing.T, detail *RecognitionDetail) {
	t.Helper()
	require.NotNil(t, detail)
	require.NotNil(t, detail.CombinedResult)

	combined, err := parseCombinedResult(detail.DetailJson)
	require.NoError(t, err)
	require.Len(t, detail.CombinedResult, len(combined))

	for i, rawItem := range combined {
		gotItem := detail.CombinedResult[i]
		require.NotNil(t, gotItem)
		require.Equal(t, rawItem.ID, gotItem.ID)
		require.Equal(t, rawItem.Name, gotItem.Name)
		require.Equal(t, rawItem.Algorithm, gotItem.Algorithm)
		require.Equal(t, rawItem.Box, gotItem.Box)

		switch NodeRecognitionType(rawItem.Algorithm) {
		case NodeRecognitionTypeAnd, NodeRecognitionTypeOr:
			requireRecognitionDetailMatchesCombinedRaw(t, gotItem)
		default:
			requireRecognitionDetailMatchesRaw(t, gotItem)
		}
	}
}

func TestRecognitionDetail_ResultMatchesRaw(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomRecognition("TestRecognitionDetail_Custom", &testRecognitionDetailCustomRec{})
	require.True(t, ok)

	act := &testRecognitionDetailFromRecognitionAct{t: t}
	ok = res.RegisterCustomAction("TestRecognitionDetail_ResultMatchesRawAct", act)
	require.True(t, ok)

	pipeline := NewPipeline()
	testNode := NewNode("TestRecognitionDetail_ResultMatchesRaw",
		WithAction(ActCustom("TestRecognitionDetail_ResultMatchesRawAct")),
	)
	pipeline.AddNode(testNode)

	got := tasker.PostTask(testNode.Name, pipeline).
		Wait().Success()
	require.True(t, got)

	require.NotNil(t, act)
}
