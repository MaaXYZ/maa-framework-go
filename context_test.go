package maa

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testContextRunTaskAct struct {
	t *testing.T
}

func (t *testContextRunTaskAct) Run(ctx *Context, _ *CustomActionArg) bool {
	pipeline := NewPipeline()
	testNode := NewNode("Test").
		SetAction(ActClick(ClickParam{
			Target: NewTargetRect(Rect{100, 100, 10, 10}),
		}))
	pipeline.AddNode(testNode)

	detail, err := ctx.RunTask(testNode.Name, pipeline)
	require.NoError(t.t, err)
	require.NotNil(t.t, detail)
	return true
}

func TestContext_RunTask(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	err := res.RegisterCustomAction("TestContext_RunPipelineAct", &testContextRunTaskAct{t})
	require.NoError(t, err)

	pipeline := NewPipeline()
	testContext_RunPipelineNode := NewNode("TestContext_RunPipeline").
		SetAction(ActCustom(CustomActionParam{CustomAction: "TestContext_RunPipelineAct"}))
	pipeline.AddNode(testContext_RunPipelineNode)

	got := tasker.PostTask(testContext_RunPipelineNode.Name, pipeline).
		Wait().Success()
	require.True(t, got)
}

type testContextRunRecognitionAct struct {
	t *testing.T
}

func (t *testContextRunRecognitionAct) Run(ctx *Context, _ *CustomActionArg) bool {
	img, err := ctx.GetTasker().GetController().CacheImage()
	require.NoError(t.t, err)
	require.NotNil(t.t, img)

	pipeline := NewPipeline()
	testNode := NewNode("Test").
		SetRecognition(RecOCR(OCRParam{
			Expected: []string{"Hello"},
		}))
	pipeline.AddNode(testNode)

	detail, err := ctx.RunRecognition("Test", img, pipeline)
	require.NoError(t.t, err)
	require.NotNil(t.t, detail)
	return true
}

func TestContext_RunRecognition(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	err := res.RegisterCustomAction("TestContext_RunRecognitionAct", &testContextRunRecognitionAct{t})
	require.NoError(t, err)

	pipeline := NewPipeline()
	testContext_RunRecognitionNode := NewNode("TestContext_RunRecognition").
		AddNext("RunRecognition").
		AddNext("Stop")
	pipeline.AddNode(testContext_RunRecognitionNode)
	runRecognitionNode := NewNode("RunRecognition").
		SetRecognition(RecCustom(CustomRecognitionParam{CustomRecognition: "TestContext_RunRecognitionAct"}))
	pipeline.AddNode(runRecognitionNode)
	stopNode := NewNode("Stop")
	pipeline.AddNode(stopNode)

	got := tasker.PostTask(testContext_RunRecognitionNode.Name, pipeline).
		Wait().Success()
	require.True(t, got)
}

type testContextRunActionAct struct {
	t *testing.T
}

func (a testContextRunActionAct) Run(ctx *Context, arg *CustomActionArg) bool {
	pipeline := NewPipeline()
	testNode := NewNode("Test").
		SetAction(ActClick(ClickParam{
			Target: NewTargetRect(Rect{100, 100, 10, 10}),
		}))
	pipeline.AddNode(testNode)

	detail, err := ctx.RunAction(testNode.Name, arg.Box, arg.RecognitionDetail.DetailJson, pipeline)
	require.NoError(a.t, err)
	require.NotNil(a.t, detail)
	return true
}

func TestContext_RunAction(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	err := res.RegisterCustomAction("TestContext_RunActionAct", &testContextRunActionAct{t})
	require.NoError(t, err)

	pipeline := NewPipeline()
	testContext_RunActionNode := NewNode("TestContext_RunAction").
		SetAction(ActCustom(CustomActionParam{CustomAction: "TestContext_RunActionAct"}))
	pipeline.AddNode(testContext_RunActionNode)

	got := tasker.PostTask(testContext_RunActionNode.Name, pipeline).
		Wait().Success()
	require.True(t, got)
}

type testContextOverriderPipelineAct struct {
	t *testing.T
}

func (t *testContextOverriderPipelineAct) Run(ctx *Context, _ *CustomActionArg) bool {
	pipeline1 := NewPipeline()
	testNode1 := NewNode("Test").
		SetAction(ActClick(ClickParam{
			Target: NewTargetRect(Rect{100, 100, 10, 10}),
		}))
	pipeline1.AddNode(testNode1)

	detail1, err := ctx.RunTask(testNode1.Name, pipeline1)
	require.NoError(t.t, err)
	require.NotNil(t.t, detail1)

	pipeline2 := NewPipeline()
	testNode2 := NewNode("Test").
		SetAction(ActClick(ClickParam{
			Target: NewTargetRect(Rect{200, 200, 10, 10}),
		}))
	pipeline2.AddNode(testNode2)

	err = ctx.OverridePipeline(pipeline2)
	require.NoError(t.t, err)

	detail2, err2 := ctx.RunTask("Test")
	require.NoError(t.t, err2)
	require.NotNil(t.t, detail2)
	return true
}

func TestContext_OverridePipeline(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	err := res.RegisterCustomAction("TestContext_OverridePipelineAct", &testContextOverriderPipelineAct{t})
	require.NoError(t, err)

	pipeline := NewPipeline()
	testContext_OverridePipelineNode := NewNode("TestContext_OverridePipeline").
		SetAction(ActCustom(CustomActionParam{CustomAction: "TestContext_OverridePipelineAct"}))
	pipeline.AddNode(testContext_OverridePipelineNode)

	got := tasker.PostTask(testContext_OverridePipelineNode.Name, pipeline).
		Wait().Success()
	require.True(t, got)
}

type testContextOverrideNextAct struct {
	t *testing.T
}

func (t *testContextOverrideNextAct) Run(ctx *Context, _ *CustomActionArg) bool {
	pipeline := NewPipeline()
	testNode := NewNode("Test").
		SetNext([]NextItem{
			{Name: "TaskA"},
		})
	pipeline.AddNode(testNode)
	taskANode := NewNode("TaskA")
	pipeline.AddNode(taskANode)
	taskBNode := NewNode("TaskB")
	pipeline.AddNode(taskBNode)

	err := ctx.OverridePipeline(pipeline)
	require.NoError(t.t, err)

	err = ctx.OverrideNext(testNode.Name, []NextItem{{Name: "TaskB"}})
	require.NoError(t.t, err)

	detail, err2 := ctx.RunTask(testNode.Name, pipeline)
	require.NoError(t.t, err2)
	require.NotNil(t.t, detail)
	return true
}

func TestContext_OverrideNext(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	err := res.RegisterCustomAction("TestContext_OverrideNextAct", &testContextOverrideNextAct{t})
	require.NoError(t, err)

	pipeline := NewPipeline()
	testContext_OverrideNextNode := NewNode("TestContext_OverrideNext").
		SetAction(ActCustom(CustomActionParam{CustomAction: "TestContext_OverrideNextAct"}))
	pipeline.AddNode(testContext_OverrideNextNode)

	got := tasker.PostTask(testContext_OverrideNextNode.Name, pipeline).
		Wait().Success()
	require.True(t, got)
}

func TestContext_handleOverride(t *testing.T) {
	ctx := &Context{}

	type typedNil struct {
		V string
	}

	cases := []struct {
		name     string
		override []any
		want     string
	}{
		{
			name:     "no override",
			override: nil,
			want:     "{}",
		},
		{
			name:     "untyped nil",
			override: []any{nil},
			want:     "{}",
		},
		{
			name:     "typed nil pointer",
			override: []any{(*typedNil)(nil)},
			want:     "{}",
		},
		{
			name:     "nil byte slice",
			override: []any{[]byte(nil)},
			want:     "{}",
		},
		{
			name:     "string passthrough",
			override: []any{`{"A":1}`},
			want:     `{"A":1}`,
		},
		{
			name:     "byte passthrough",
			override: []any{[]byte(`{"A":1}`)},
			want:     `{"A":1}`,
		},
		{
			name:     "marshal object",
			override: []any{map[string]any{"A": 1}},
			want:     `{"A":1}`,
		},
		{
			name: "marshal error fallback",
			override: []any{map[string]any{
				"f": func() {},
			}},
			want: "{}",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ctx.handleOverride(tc.override...)
			assert.Equal(t, tc.want, got)
		})
	}
}

type testContextGetNodeDataAct struct {
	t *testing.T
}

func (a *testContextGetNodeDataAct) Run(ctx *Context, _ *CustomActionArg) bool {
	type Case struct {
		Name     string
		testFunc func(ctx *Context)
	}

	cases := []Case{
		// recognition
		{
			Name:     "DirectHit",
			testFunc: a.testDirectHitRecognition,
		},
		{
			Name:     "TemplateMatch",
			testFunc: a.testTemplateMatchRecognition,
		},
		{
			Name:     "FeatureMatch",
			testFunc: a.testFeatureMatchRecognition,
		},
		{
			Name:     "ColorMatch",
			testFunc: a.testColorMatchRecognition,
		},
		{
			Name:     "OCR",
			testFunc: a.testOCRRecognition,
		},
		{
			Name:     "NeuralNetworkClassify",
			testFunc: a.testNeuralNetworkClassifyRecognition,
		},
		{
			Name:     "NeuralNetworkDetect",
			testFunc: a.testNeuralNetworkDetectRecognition,
		},
		{
			Name:     "CustomRecognition",
			testFunc: a.testCustomRecognition,
		},
		// action
		{
			Name:     "DoNothing",
			testFunc: a.testDoNothingAction,
		},
		{
			Name:     "Click",
			testFunc: a.testClickAction,
		},
		{
			Name:     "LongPress",
			testFunc: a.testLongPressAction,
		},
		{
			Name:     "Swipe",
			testFunc: a.testSwipeAction,
		},
		{
			Name:     "MultiSwipe",
			testFunc: a.testMultiSwipeAction,
		},
		{
			Name:     "TouchDown",
			testFunc: a.testTouchDownAction,
		},
		{
			Name:     "TouchMove",
			testFunc: a.testTouchMoveAction,
		},
		{
			Name:     "TouchUp",
			testFunc: a.testTouchUpAction,
		},
		{
			Name:     "ClickKey",
			testFunc: a.testClickKeyAction,
		},
		{
			Name:     "LongPressKey",
			testFunc: a.testLongPressKeyAction,
		},
		{
			Name:     "KeyDown",
			testFunc: a.testKeyDownAction,
		},
		{
			Name:     "KeyUp",
			testFunc: a.testKeyUpAction,
		},
		{
			Name:     "InputText",
			testFunc: a.testInputTextAction,
		},
		{
			Name:     "StartApp",
			testFunc: a.testStartAppAction,
		},
		{
			Name:     "StopApp",
			testFunc: a.testStopAppAction,
		},
		{
			Name:     "StopTask",
			testFunc: a.testStopTaskAction,
		},
		{
			Name:     "Command",
			testFunc: a.testCommandAction,
		},
		{
			Name:     "Scroll",
			testFunc: a.testScrollAction,
		},
		{
			Name:     "Screencap",
			testFunc: a.testScreencapAction,
		},
		{
			Name:     "CustomAction",
			testFunc: a.testCustomAction,
		},
		// node attributes
		{
			Name:     "NodeAttributes",
			testFunc: a.testNodeAttributes,
		},
	}

	for _, c := range cases {
		a.t.Run(c.Name, func(t *testing.T) {
			c.testFunc(ctx)
		})
	}

	return true
}

func (a *testContextGetNodeDataAct) testClickAction(ctx *Context) {
	raw := map[string]any{
		"test_click": map[string]any{
			"action": map[string]any{
				"type": "Click",
				"param": map[string]any{
					"target":        []int{100, 200, 50, 50},
					"target_offset": []int{10, 10, 0, 0},
					"contact":       1,
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_click")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeClick, nodeData.Action.Type)
	assert.IsType(a.t, (*ClickParam)(nil), nodeData.Action.Param)

	clickParam := nodeData.Action.Param.(*ClickParam)
	assert.Equal(a.t, 1, clickParam.Contact)
}

func (a *testContextGetNodeDataAct) testDirectHitRecognition(ctx *Context) {
	raw := map[string]any{
		"test_direct_hit": map[string]any{
			"recognition": map[string]any{
				"type": "DirectHit",
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_direct_hit")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, RecognitionTypeDirectHit, nodeData.Recognition.Type)
}

func (a *testContextGetNodeDataAct) testTemplateMatchRecognition(ctx *Context) {
	raw := map[string]any{
		"test_template": map[string]any{
			"recognition": map[string]any{
				"type": "TemplateMatch",
				"param": map[string]any{
					"template":   []string{"test.png", "test2.png"},
					"roi":        []int{0, 0, 100, 100},
					"roi_offset": []int{10, 10, 0, 0},
					"threshold":  0.8,
					"order_by":   "Score",
					"index":      1,
					"method":     5,
					"green_mask": true,
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_template")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, RecognitionTypeTemplateMatch, nodeData.Recognition.Type)
	assert.IsType(a.t, (*TemplateMatchParam)(nil), nodeData.Recognition.Param)

	param := nodeData.Recognition.Param.(*TemplateMatchParam)
	assert.Equal(a.t, []string{"test.png", "test2.png"}, param.Template)
	assert.Equal(a.t, []float64{0.8}, param.Threshold)
	assert.Equal(a.t, TemplateMatchOrderByScore, param.OrderBy)
	assert.Equal(a.t, 1, param.Index)
	assert.Equal(a.t, TemplateMatchMethodCCOEFF_NORMED, param.Method)
	assert.True(a.t, param.GreenMask)
}

func (a *testContextGetNodeDataAct) testFeatureMatchRecognition(ctx *Context) {
	raw := map[string]any{
		"test_feature": map[string]any{
			"recognition": map[string]any{
				"type": "FeatureMatch",
				"param": map[string]any{
					"template":   []string{"feature.png"},
					"roi":        []int{0, 0, 200, 200},
					"count":      10,
					"order_by":   "Area",
					"index":      0,
					"green_mask": false,
					"detector":   "SIFT",
					"ratio":      0.7,
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_feature")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, RecognitionTypeFeatureMatch, nodeData.Recognition.Type)
	assert.IsType(a.t, (*FeatureMatchParam)(nil), nodeData.Recognition.Param)

	param := nodeData.Recognition.Param.(*FeatureMatchParam)
	assert.Equal(a.t, []string{"feature.png"}, param.Template)
	assert.Equal(a.t, 10, param.Count)
	assert.Equal(a.t, FeatureMatchOrderByArea, param.OrderBy)
	assert.Equal(a.t, FeatureMatchMethodSIFT, param.Detector)
	assert.Equal(a.t, 0.7, param.Ratio)
}

func (a *testContextGetNodeDataAct) testColorMatchRecognition(ctx *Context) {
	raw := map[string]any{
		"test_color": map[string]any{
			"recognition": map[string]any{
				"type": "ColorMatch",
				"param": map[string]any{
					"roi":       []int{0, 0, 100, 100},
					"method":    4,
					"lower":     [][]int{{0, 0, 0}},
					"upper":     [][]int{{255, 255, 255}},
					"count":     100,
					"order_by":  "Score",
					"connected": true,
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_color")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, RecognitionTypeColorMatch, nodeData.Recognition.Type)
	assert.IsType(a.t, (*ColorMatchParam)(nil), nodeData.Recognition.Param)

	param := nodeData.Recognition.Param.(*ColorMatchParam)
	assert.Equal(a.t, ColorMatchMethodRGB, param.Method)
	assert.Equal(a.t, [][]int{{0, 0, 0}}, param.Lower)
	assert.Equal(a.t, [][]int{{255, 255, 255}}, param.Upper)
	assert.Equal(a.t, 100, param.Count)
	assert.True(a.t, param.Connected)
}

func (a *testContextGetNodeDataAct) testOCRRecognition(ctx *Context) {
	raw := map[string]any{
		"test_ocr": map[string]any{
			"recognition": map[string]any{
				"type": "OCR",
				"param": map[string]any{
					"roi":          []int{0, 0, 300, 100},
					"expected":     []string{"Hello", "World"},
					"threshold":    0.5,
					"replace":      [][]string{{"0", "O"}, {"1", "l"}},
					"order_by":     "Length",
					"index":        0,
					"only_rec":     true,
					"model":        "ppocr_v4",
					"color_filter": "RecoColorMatch",
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_ocr")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, RecognitionTypeOCR, nodeData.Recognition.Type)
	assert.IsType(a.t, (*OCRParam)(nil), nodeData.Recognition.Param)

	param := nodeData.Recognition.Param.(*OCRParam)
	assert.Equal(a.t, []string{"Hello", "World"}, param.Expected)
	assert.Equal(a.t, 0.5, param.Threshold)
	assert.Equal(a.t, OCROrderByLength, param.OrderBy)
	assert.True(a.t, param.OnlyRec)
	assert.Equal(a.t, "ppocr_v4", param.Model)
	assert.Equal(a.t, "RecoColorMatch", param.ColorFilter)
}

func (a *testContextGetNodeDataAct) testNeuralNetworkClassifyRecognition(ctx *Context) {
	raw := map[string]any{
		"test_nn_classify": map[string]any{
			"recognition": map[string]any{
				"type": "NeuralNetworkClassify",
				"param": map[string]any{
					"roi":      []int{0, 0, 224, 224},
					"labels":   []string{"Cat", "Dog", "Mouse"},
					"model":    "classifier.onnx",
					"expected": []int{0, 2},
					"order_by": "Score",
					"index":    0,
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_nn_classify")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, RecognitionTypeNeuralNetworkClassify, nodeData.Recognition.Type)
	assert.IsType(a.t, (*NeuralNetworkClassifyParam)(nil), nodeData.Recognition.Param)

	param := nodeData.Recognition.Param.(*NeuralNetworkClassifyParam)
	assert.Equal(a.t, []string{"Cat", "Dog", "Mouse"}, param.Labels)
	assert.Equal(a.t, "classifier.onnx", param.Model)
	assert.Equal(a.t, []int{0, 2}, param.Expected)
}

func (a *testContextGetNodeDataAct) testNeuralNetworkDetectRecognition(ctx *Context) {
	raw := map[string]any{
		"test_nn_detect": map[string]any{
			"recognition": map[string]any{
				"type": "NeuralNetworkDetect",
				"param": map[string]any{
					"roi":      []int{0, 0, 640, 640},
					"labels":   []string{"person", "car", "bicycle"},
					"model":    "yolov8.onnx",
					"expected": []int{0, 1},
					"order_by": "Area",
					"index":    -1,
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_nn_detect")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, RecognitionTypeNeuralNetworkDetect, nodeData.Recognition.Type)
	assert.IsType(a.t, (*NeuralNetworkDetectParam)(nil), nodeData.Recognition.Param)

	param := nodeData.Recognition.Param.(*NeuralNetworkDetectParam)
	assert.Equal(a.t, []string{"person", "car", "bicycle"}, param.Labels)
	assert.Equal(a.t, "yolov8.onnx", param.Model)
	assert.Equal(a.t, []int{0, 1}, param.Expected)
	assert.Equal(a.t, NeuralNetworkDetectOrderByArea, param.OrderBy)
	assert.Equal(a.t, -1, param.Index)
}

func (a *testContextGetNodeDataAct) testCustomRecognition(ctx *Context) {
	raw := map[string]any{
		"test_custom_rec": map[string]any{
			"recognition": map[string]any{
				"type": "Custom",
				"param": map[string]any{
					"roi":                      []int{0, 0, 100, 100},
					"custom_recognition":       "MyCustomRecognizer",
					"custom_recognition_param": map[string]any{"key": "value"},
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_custom_rec")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, RecognitionTypeCustom, nodeData.Recognition.Type)
	assert.IsType(a.t, (*CustomRecognitionParam)(nil), nodeData.Recognition.Param)

	param := nodeData.Recognition.Param.(*CustomRecognitionParam)
	assert.Equal(a.t, "MyCustomRecognizer", param.CustomRecognition)
	assert.NotNil(a.t, param.CustomRecognitionParam)
}

func (a *testContextGetNodeDataAct) testDoNothingAction(ctx *Context) {
	raw := map[string]any{
		"test_do_nothing": map[string]any{
			"action": map[string]any{
				"type": "DoNothing",
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_do_nothing")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, ActionTypeDoNothing, nodeData.Action.Type)
}

func (a *testContextGetNodeDataAct) testLongPressAction(ctx *Context) {
	raw := map[string]any{
		"test_long_press": map[string]any{
			"action": map[string]any{
				"type": "LongPress",
				"param": map[string]any{
					"target":   []int{100, 200, 50, 50},
					"duration": 2000,
					"contact":  0,
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_long_press")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeLongPress, nodeData.Action.Type)
	assert.IsType(a.t, (*LongPressParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*LongPressParam)
	assert.Equal(a.t, int64(2000), param.Duration.Milliseconds())
}

func (a *testContextGetNodeDataAct) testSwipeAction(ctx *Context) {
	raw := map[string]any{
		"test_swipe": map[string]any{
			"action": map[string]any{
				"type": "Swipe",
				"param": map[string]any{
					"begin":        []int{100, 500, 10, 10},
					"begin_offset": []int{0, 0, 0, 0},
					"end":          []int{100, 100, 10, 10},
					"end_offset":   []int{0, 0, 0, 0},
					"duration":     500,
					"end_hold":     100,
					"only_hover":   false,
					"contact":      0,
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_swipe")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeSwipe, nodeData.Action.Type)
	assert.IsType(a.t, (*SwipeParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*SwipeParam)
	assert.Len(a.t, param.Duration, 1)
	assert.Equal(a.t, int64(500), param.Duration[0].Milliseconds())
	assert.Len(a.t, param.EndHold, 1)
	assert.Equal(a.t, int64(100), param.EndHold[0].Milliseconds())
}

func (a *testContextGetNodeDataAct) testMultiSwipeAction(ctx *Context) {
	raw := map[string]any{
		"test_multi_swipe": map[string]any{
			"action": map[string]any{
				"type": "MultiSwipe",
				"param": map[string]any{
					"swipes": []map[string]any{
						{
							"starting": 0,
							"begin":    []int{100, 500, 10, 10},
							"end":      []int{100, 100, 10, 10},
							"duration": 300,
							"contact":  0,
						},
						{
							"starting": 100,
							"begin":    []int{200, 500, 10, 10},
							"end":      []int{200, 100, 10, 10},
							"duration": 300,
							"contact":  1,
						},
					},
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_multi_swipe")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeMultiSwipe, nodeData.Action.Type)
	assert.IsType(a.t, (*MultiSwipeParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*MultiSwipeParam)
	assert.Len(a.t, param.Swipes, 2)
	assert.Equal(a.t, int64(0), param.Swipes[0].Starting.Milliseconds())
	assert.Equal(a.t, int64(100), param.Swipes[1].Starting.Milliseconds())
}

func (a *testContextGetNodeDataAct) testTouchDownAction(ctx *Context) {
	raw := map[string]any{
		"test_touch_down": map[string]any{
			"action": map[string]any{
				"type": "TouchDown",
				"param": map[string]any{
					"target":   []int{100, 200, 10, 10},
					"pressure": 50,
					"contact":  0,
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_touch_down")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeTouchDown, nodeData.Action.Type)
	assert.IsType(a.t, (*TouchDownParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*TouchDownParam)
	assert.Equal(a.t, 50, param.Pressure)
}

func (a *testContextGetNodeDataAct) testTouchMoveAction(ctx *Context) {
	raw := map[string]any{
		"test_touch_move": map[string]any{
			"action": map[string]any{
				"type": "TouchMove",
				"param": map[string]any{
					"target":   []int{150, 250, 10, 10},
					"pressure": 30,
					"contact":  0,
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_touch_move")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeTouchMove, nodeData.Action.Type)
	assert.IsType(a.t, (*TouchMoveParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*TouchMoveParam)
	assert.Equal(a.t, 30, param.Pressure)
}

func (a *testContextGetNodeDataAct) testTouchUpAction(ctx *Context) {
	raw := map[string]any{
		"test_touch_up": map[string]any{
			"action": map[string]any{
				"type": "TouchUp",
				"param": map[string]any{
					"contact": 1,
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_touch_up")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeTouchUp, nodeData.Action.Type)
	assert.IsType(a.t, (*TouchUpParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*TouchUpParam)
	assert.Equal(a.t, 1, param.Contact)
}

func (a *testContextGetNodeDataAct) testClickKeyAction(ctx *Context) {
	raw := map[string]any{
		"test_click_key": map[string]any{
			"action": map[string]any{
				"type": "ClickKey",
				"param": map[string]any{
					"key": []int{4, 66},
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_click_key")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeClickKey, nodeData.Action.Type)
	assert.IsType(a.t, (*ClickKeyParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*ClickKeyParam)
	assert.Equal(a.t, []int{4, 66}, param.Key)
}

func (a *testContextGetNodeDataAct) testLongPressKeyAction(ctx *Context) {
	raw := map[string]any{
		"test_long_press_key": map[string]any{
			"action": map[string]any{
				"type": "LongPressKey",
				"param": map[string]any{
					"key":      4,
					"duration": 1500,
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_long_press_key")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeLongPressKey, nodeData.Action.Type)
	assert.IsType(a.t, (*LongPressKeyParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*LongPressKeyParam)
	assert.Equal(a.t, []int{4}, param.Key)
	assert.Equal(a.t, int64(1500), param.Duration.Milliseconds())
}

func (a *testContextGetNodeDataAct) testKeyDownAction(ctx *Context) {
	raw := map[string]any{
		"test_key_down": map[string]any{
			"action": map[string]any{
				"type": "KeyDown",
				"param": map[string]any{
					"key": 29,
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_key_down")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeKeyDown, nodeData.Action.Type)
	assert.IsType(a.t, (*KeyDownParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*KeyDownParam)
	assert.Equal(a.t, 29, param.Key)
}

func (a *testContextGetNodeDataAct) testKeyUpAction(ctx *Context) {
	raw := map[string]any{
		"test_key_up": map[string]any{
			"action": map[string]any{
				"type": "KeyUp",
				"param": map[string]any{
					"key": 29,
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_key_up")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeKeyUp, nodeData.Action.Type)
	assert.IsType(a.t, (*KeyUpParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*KeyUpParam)
	assert.Equal(a.t, 29, param.Key)
}

func (a *testContextGetNodeDataAct) testInputTextAction(ctx *Context) {
	raw := map[string]any{
		"test_input_text": map[string]any{
			"action": map[string]any{
				"type": "InputText",
				"param": map[string]any{
					"input_text": "Hello World",
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_input_text")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeInputText, nodeData.Action.Type)
	assert.IsType(a.t, (*InputTextParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*InputTextParam)
	assert.Equal(a.t, "Hello World", param.InputText)
}

func (a *testContextGetNodeDataAct) testStartAppAction(ctx *Context) {
	raw := map[string]any{
		"test_start_app": map[string]any{
			"action": map[string]any{
				"type": "StartApp",
				"param": map[string]any{
					"package": "com.example.app/com.example.MainActivity",
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_start_app")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeStartApp, nodeData.Action.Type)
	assert.IsType(a.t, (*StartAppParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*StartAppParam)
	assert.Equal(a.t, "com.example.app/com.example.MainActivity", param.Package)
}

func (a *testContextGetNodeDataAct) testStopAppAction(ctx *Context) {
	raw := map[string]any{
		"test_stop_app": map[string]any{
			"action": map[string]any{
				"type": "StopApp",
				"param": map[string]any{
					"package": "com.example.app",
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_stop_app")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeStopApp, nodeData.Action.Type)
	assert.IsType(a.t, (*StopAppParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*StopAppParam)
	assert.Equal(a.t, "com.example.app", param.Package)
}

func (a *testContextGetNodeDataAct) testStopTaskAction(ctx *Context) {
	raw := map[string]any{
		"test_stop_task": map[string]any{
			"action": map[string]any{
				"type": "StopTask",
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_stop_task")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeStopTask, nodeData.Action.Type)
}

func (a *testContextGetNodeDataAct) testCommandAction(ctx *Context) {
	raw := map[string]any{
		"test_command": map[string]any{
			"action": map[string]any{
				"type": "Command",
				"param": map[string]any{
					"exec":   "python",
					"args":   []string{"{RESOURCE_DIR}/script.py", "{NODE}", "{IMAGE}"},
					"detach": true,
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_command")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeCommand, nodeData.Action.Type)
	assert.IsType(a.t, (*CommandParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*CommandParam)
	assert.Equal(a.t, "python", param.Exec)
	assert.Equal(a.t, []string{"{RESOURCE_DIR}/script.py", "{NODE}", "{IMAGE}"}, param.Args)
	assert.True(a.t, param.Detach)
}

func (a *testContextGetNodeDataAct) testScrollAction(ctx *Context) {
	raw := map[string]any{
		"test_scroll": map[string]any{
			"action": map[string]any{
				"type": "Scroll",
				"param": map[string]any{
					"dx": 100,
					"dy": 200,
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_scroll")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeScroll, nodeData.Action.Type)
	assert.IsType(a.t, (*ScrollParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*ScrollParam)
	assert.Equal(a.t, 100, param.Dx)
	assert.Equal(a.t, 200, param.Dy)
}

func (a *testContextGetNodeDataAct) testScreencapAction(ctx *Context) {
	raw := map[string]any{
		"test_screencap": map[string]any{
			"action": map[string]any{
				"type": "Screencap",
				"param": map[string]any{
					"filename": "capture_test",
					"format":   "jpg",
					"quality":  85,
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_screencap")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeScreencap, nodeData.Action.Type)
	assert.IsType(a.t, (*ScreencapParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*ScreencapParam)
	assert.Equal(a.t, "capture_test", param.Filename)
	assert.Equal(a.t, "jpg", param.Format)
	assert.Equal(a.t, 85, param.Quality)
}

func (a *testContextGetNodeDataAct) testCustomAction(ctx *Context) {
	raw := map[string]any{
		"test_custom_act": map[string]any{
			"action": map[string]any{
				"type": "Custom",
				"param": map[string]any{
					"target":              true,
					"custom_action":       "MyCustomAction",
					"custom_action_param": map[string]any{"option": "value"},
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_custom_act")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)
	assert.Equal(a.t, ActionTypeCustom, nodeData.Action.Type)
	assert.IsType(a.t, (*CustomActionParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*CustomActionParam)
	assert.Equal(a.t, "MyCustomAction", param.CustomAction)
	assert.NotNil(a.t, param.CustomActionParam)
}

func (a *testContextGetNodeDataAct) testNodeAttributes(ctx *Context) {
	raw := map[string]any{
		"test_attributes": map[string]any{
			"recognition": map[string]any{
				"type": "DirectHit",
			},
			"action": map[string]any{
				"type": "Click",
			},
			"next": []any{
				"NodeA",
				map[string]any{
					"name":      "NodeB",
					"jump_back": true,
				},
				map[string]any{
					"name":   "AnchorX",
					"anchor": true,
				},
			},
			"rate_limit": 500,
			"timeout":    30000,
			"on_error": []any{
				"ErrorHandler",
			},
			"inverse":    true,
			"enabled":    true,
			"max_hit":    5,
			"pre_delay":  100,
			"post_delay": 150,
			"pre_wait_freezes": map[string]any{
				"time":       1000,
				"target":     true,
				"threshold":  0.95,
				"method":     5,
				"rate_limit": 500,
				"timeout":    10000,
			},
			"post_wait_freezes": map[string]any{
				"time": 500,
			},
			"focus": map[string]any{
				"Node.Action.Succeeded": "{name} completed",
			},
			"attach": map[string]any{
				"custom_key": "custom_value",
			},
			"anchor": map[string]any{
				"MyAnchor":      "test_attributes",
				"AnotherAnchor": "test_attributes",
				"ClearedAnchor": "",
			},
		},
		"NodeA": map[string]any{},
		"NodeB": map[string]any{},
		"NodeC": map[string]any{
			"anchor": map[string]any{
				"AnchorX":       "NodeC",
				"ClearedAnchor": "",
			},
		},
		"ErrorHandler": map[string]any{},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNode("test_attributes")
	assert.NoError(a.t, err)
	assert.NotNil(a.t, nodeData)

	// Check recognition and action
	assert.Equal(a.t, RecognitionTypeDirectHit, nodeData.Recognition.Type)
	assert.Equal(a.t, ActionTypeClick, nodeData.Action.Type)

	// Check next list
	assert.Len(a.t, nodeData.Next, 3)
	assert.Equal(a.t, "NodeA", nodeData.Next[0].Name)
	assert.False(a.t, nodeData.Next[0].JumpBack)
	assert.Equal(a.t, "NodeB", nodeData.Next[1].Name)
	assert.True(a.t, nodeData.Next[1].JumpBack)
	assert.Equal(a.t, "AnchorX", nodeData.Next[2].Name)
	assert.True(a.t, nodeData.Next[2].Anchor)

	// Check timing properties
	assert.NotNil(a.t, nodeData.RateLimit)
	assert.Equal(a.t, int64(500), *nodeData.RateLimit)
	assert.NotNil(a.t, nodeData.Timeout)
	assert.Equal(a.t, int64(30000), *nodeData.Timeout)

	// Check on_error
	assert.Len(a.t, nodeData.OnError, 1)
	assert.Equal(a.t, "ErrorHandler", nodeData.OnError[0].Name)

	// Check boolean and numeric properties
	assert.True(a.t, nodeData.Inverse)
	assert.NotNil(a.t, nodeData.Enabled)
	assert.True(a.t, *nodeData.Enabled)
	assert.NotNil(a.t, nodeData.MaxHit)
	assert.Equal(a.t, uint64(5), *nodeData.MaxHit)

	// Check delays
	assert.NotNil(a.t, nodeData.PreDelay)
	assert.Equal(a.t, int64(100), *nodeData.PreDelay)
	assert.NotNil(a.t, nodeData.PostDelay)
	assert.Equal(a.t, int64(150), *nodeData.PostDelay)

	// Check wait freezes
	assert.NotNil(a.t, nodeData.PreWaitFreezes)
	assert.Equal(a.t, int64(1000), nodeData.PreWaitFreezes.Time.Milliseconds())
	assert.Equal(a.t, 0.95, nodeData.PreWaitFreezes.Threshold)
	assert.NotNil(a.t, nodeData.PostWaitFreezes)
	assert.Equal(a.t, int64(500), nodeData.PostWaitFreezes.Time.Milliseconds())

	// Check focus
	assert.NotNil(a.t, nodeData.Focus)

	// Check attach
	assert.NotNil(a.t, nodeData.Attach)
	assert.Equal(a.t, "custom_value", nodeData.Attach["custom_key"])

	// Check anchor (GetNodeData outputs anchor as object map)
	assert.Equal(a.t, map[string]string{
		"MyAnchor":      "test_attributes",
		"AnotherAnchor": "test_attributes",
		"ClearedAnchor": "",
	}, nodeData.Anchor)
}

func TestContext_GetNode(t *testing.T) {
	ctrl, err := NewBlankController()
	require.NoError(t, err)
	require.NotNil(t, ctrl)
	defer ctrl.Destroy()
	connected := ctrl.PostConnect().Wait().Success()
	require.True(t, connected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()

	taskerBind(t, tasker, ctrl, res)

	err = res.RegisterCustomAction("TestContext_GetNodeAct", &testContextGetNodeDataAct{t})
	require.NoError(t, err)

	pipeline := NewPipeline()
	launchNode := NewNode("launch").
		SetAction(ActCustom(CustomActionParam{CustomAction: "TestContext_GetNodeAct"}))
	pipeline.AddNode(launchNode)

	got := tasker.PostTask(launchNode.Name, pipeline).
		Wait().
		Success()
	require.True(t, got)
}

type testContextGetTaskJobAct struct {
	t *testing.T
}

func (t *testContextGetTaskJobAct) Run(ctx *Context, _ *CustomActionArg) bool {
	job := ctx.GetTaskJob()
	require.NotNil(t.t, job)
	return true
}

func TestContext_GetTaskJob(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	err := res.RegisterCustomAction("TestContext_GetTaskJobAct", &testContextGetTaskJobAct{t})
	require.NoError(t, err)

	pipeline := NewPipeline()
	testContext_GetTaskJobNode := NewNode("TestContext_GetTaskJob").
		SetAction(ActCustom(CustomActionParam{CustomAction: "TestContext_GetTaskJobAct"}))
	pipeline.AddNode(testContext_GetTaskJobNode)

	got := tasker.PostTask(testContext_GetTaskJobNode.Name, pipeline).
		Wait().Success()
	require.True(t, got)
}

type testContextGetTaskerAct struct {
	t *testing.T
}

func (t testContextGetTaskerAct) Run(ctx *Context, _ *CustomActionArg) bool {
	tasker := ctx.GetTasker()
	require.NotNil(t.t, tasker)
	return true
}

func TestContext_GetTasker(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	err := res.RegisterCustomAction("TestContext_GetTaskerAct", &testContextGetTaskerAct{t})
	require.NoError(t, err)

	pipeline := NewPipeline()
	testContext_GetTaskerNode := NewNode("TestContext_GetTasker").
		SetAction(ActCustom(CustomActionParam{CustomAction: "TestContext_GetTaskerAct"}))
	pipeline.AddNode(testContext_GetTaskerNode)

	got := tasker.PostTask(testContext_GetTaskerNode.Name, pipeline).
		Wait().Success()
	require.True(t, got)
}

type testContextCloneAct struct {
	t *testing.T
}

func (t testContextCloneAct) Run(ctx *Context, _ *CustomActionArg) bool {
	cloneCtx := ctx.Clone()
	require.NotNil(t.t, cloneCtx)
	return true
}

func TestContext_Clone(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	err := res.RegisterCustomAction("TestContext_GetTaskerAct", &testContextCloneAct{t})
	require.NoError(t, err)

	pipeline := NewPipeline()
	testContext_CloneNode := NewNode("TestContext_Clone").
		SetAction(ActCustom(CustomActionParam{CustomAction: "TestContext_GetTaskerAct"}))
	pipeline.AddNode(testContext_CloneNode)

	got := tasker.PostTask(testContext_CloneNode.Name, pipeline).
		Wait().Success()
	require.True(t, got)
}

type testContextRunRecognitionDirectAct struct {
	t *testing.T
}

func (a *testContextRunRecognitionDirectAct) Run(ctx *Context, _ *CustomActionArg) bool {
	img, err := ctx.GetTasker().GetController().CacheImage()
	require.NoError(a.t, err)
	require.NotNil(a.t, img)

	// Test RunRecognitionDirect with DirectHit recognition type
	detail, err := ctx.RunRecognitionDirect(RecognitionTypeDirectHit, &DirectHitParam{}, img)
	require.NoError(a.t, err)
	require.NotNil(a.t, detail)
	require.True(a.t, detail.Hit)
	return true
}

func TestContext_RunRecognitionDirect(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	err := res.RegisterCustomAction("TestContext_RunRecognitionDirectAct", &testContextRunRecognitionDirectAct{t})
	require.NoError(t, err)

	pipeline := NewPipeline()
	testContext_RunRecognitionDirectNode := NewNode("TestContext_RunRecognitionDirect").
		SetAction(ActCustom(CustomActionParam{CustomAction: "TestContext_RunRecognitionDirectAct"}))
	pipeline.AddNode(testContext_RunRecognitionDirectNode)

	got := tasker.PostTask(testContext_RunRecognitionDirectNode.Name, pipeline).
		Wait().Success()
	require.True(t, got)
}

type testContextRunActionDirectAct struct {
	t *testing.T
}

func (a *testContextRunActionDirectAct) Run(ctx *Context, arg *CustomActionArg) bool {
	// Test RunActionDirect with Click action type
	clickParam := &ClickParam{
		Target: NewTargetRect(Rect{100, 100, 10, 10}),
	}
	detail, err := ctx.RunActionDirect(ActionTypeClick, clickParam, arg.Box, arg.RecognitionDetail)
	require.NoError(a.t, err)
	require.NotNil(a.t, detail)
	return true
}

func TestContext_RunActionDirect(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	err := res.RegisterCustomAction("TestContext_RunActionDirectAct", &testContextRunActionDirectAct{t})
	require.NoError(t, err)

	pipeline := NewPipeline()
	testContext_RunActionDirectNode := NewNode("TestContext_RunActionDirect").
		SetAction(ActCustom(CustomActionParam{CustomAction: "TestContext_RunActionDirectAct"}))
	pipeline.AddNode(testContext_RunActionDirectNode)

	got := tasker.PostTask(testContext_RunActionDirectNode.Name, pipeline).
		Wait().Success()
	require.True(t, got)
}
