package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testContextRunTaskAct struct {
	t *testing.T
}

func (t *testContextRunTaskAct) Run(ctx *Context, _ *CustomActionArg) bool {
	pipeline := NewPipeline()
	testNode := NewNode("Test",
		WithAction(ActClick(
			WithClickTarget(NewTargetRect(Rect{100, 100, 10, 10})),
		)),
	)
	pipeline.AddNode(testNode)

	detail := ctx.RunTask(testNode.Name, pipeline)
	require.NotNil(t.t, detail)
	return true
}

func TestContext_RunTask(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomAction("TestContext_RunPipelineAct", &testContextRunTaskAct{t})
	require.True(t, ok)

	pipeline := NewPipeline()
	testContext_RunPipelineNode := NewNode("TestContext_RunPipeline",
		WithAction(ActCustom("TestContext_RunPipelineAct")),
	)
	pipeline.AddNode(testContext_RunPipelineNode)

	got := tasker.PostTask(testContext_RunPipelineNode.Name, pipeline).
		Wait().Success()
	require.True(t, got)
}

type testContextRunRecognitionAct struct {
	t *testing.T
}

func (t *testContextRunRecognitionAct) Run(ctx *Context, _ *CustomActionArg) bool {
	img := ctx.GetTasker().GetController().CacheImage()
	require.NotNil(t.t, img)

	pipeline := NewPipeline()
	testNode := NewNode("Test",
		WithRecognition(RecOCR(
			WithOCRExpected([]string{"Hello"}),
		)),
	)
	pipeline.AddNode(testNode)

	_ = ctx.RunRecognition("Test", img, pipeline)
	return true
}

func TestContext_RunRecognition(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomAction("TestContext_RunRecognitionAct", &testContextRunRecognitionAct{t})
	require.True(t, ok)

	pipeline := NewPipeline()
	testContext_RunRecognitionNode := NewNode("TestContext_RunRecognition").
		AddNext("RunRecognition").
		AddNext("Stop")
	pipeline.AddNode(testContext_RunRecognitionNode)
	runRecognitionNode := NewNode("RunRecognition",
		WithRecognition(RecCustom("TestContext_RunRecognitionAct")),
	)
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
	testNode := NewNode("Test",
		WithAction(ActClick(
			WithClickTarget(NewTargetRect(Rect{100, 100, 10, 10})),
		)),
	)
	pipeline.AddNode(testNode)

	detail := ctx.RunAction(testNode.Name, arg.Box, arg.RecognitionDetail.DetailJson, pipeline)
	require.NotNil(a.t, detail)
	return true
}

func TestContext_RunAction(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomAction("TestContext_RunActionAct", &testContextRunActionAct{t})
	require.True(t, ok)

	pipeline := NewPipeline()
	testContext_RunActionNode := NewNode("TestContext_RunAction",
		WithAction(ActCustom("TestContext_RunActionAct")),
	)
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
	testNode1 := NewNode("Test",
		WithAction(ActClick(
			WithClickTarget(NewTargetRect(Rect{100, 100, 10, 10})),
		)),
	)
	pipeline1.AddNode(testNode1)

	detail1 := ctx.RunTask(testNode1.Name, pipeline1)
	require.NotNil(t.t, detail1)

	pipeline2 := NewPipeline()
	testNode2 := NewNode("Test",
		WithAction(ActClick(
			WithClickTarget(NewTargetRect(Rect{200, 200, 10, 10})),
		)),
	)
	pipeline2.AddNode(testNode2)

	ok := ctx.OverridePipeline(pipeline2)
	require.True(t.t, ok)

	detail2 := ctx.RunTask("Test")
	require.NotNil(t.t, detail2)
	return true
}

func TestContext_OverridePipeline(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomAction("TestContext_OverridePipelineAct", &testContextOverriderPipelineAct{t})
	require.True(t, ok)

	pipeline := NewPipeline()
	testContext_OverridePipelineNode := NewNode("TestContext_OverridePipeline",
		WithAction(ActCustom("TestContext_OverridePipelineAct")),
	)
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
	testNode := NewNode("Test",
		WithNext([]NodeNextItem{
			{Name: "TaskA"},
		}),
	)
	pipeline.AddNode(testNode)
	taskANode := NewNode("TaskA")
	pipeline.AddNode(taskANode)
	taskBNode := NewNode("TaskB")
	pipeline.AddNode(taskBNode)

	ok1 := ctx.OverridePipeline(pipeline)
	require.True(t.t, ok1)

	ok2 := ctx.OverrideNext(testNode.Name, []string{"TaskB"})
	require.True(t.t, ok2)

	detail := ctx.RunTask(testNode.Name, pipeline)
	require.NotNil(t.t, detail)
	return true
}

func TestContext_OverrideNext(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomAction("TestContext_OverrideNextAct", &testContextOverrideNextAct{t})
	require.True(t, ok)

	pipeline := NewPipeline()
	testContext_OverrideNextNode := NewNode("TestContext_OverrideNext",
		WithAction(ActCustom("TestContext_OverrideNextAct")),
	)
	pipeline.AddNode(testContext_OverrideNextNode)

	got := tasker.PostTask(testContext_OverrideNextNode.Name, pipeline).
		Wait().Success()
	require.True(t, got)
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

	nodeData, err := ctx.GetNodeData("test_click")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeClick, nodeData.Action.Type)
	require.IsType(a.t, (*NodeClickParam)(nil), nodeData.Action.Param)

	clickParam := nodeData.Action.Param.(*NodeClickParam)
	require.Equal(a.t, 1, clickParam.Contact)
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

	nodeData, err := ctx.GetNodeData("test_direct_hit")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeRecognitionTypeDirectHit, nodeData.Recognition.Type)
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

	nodeData, err := ctx.GetNodeData("test_template")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeRecognitionTypeTemplateMatch, nodeData.Recognition.Type)
	require.IsType(a.t, (*NodeTemplateMatchParam)(nil), nodeData.Recognition.Param)

	param := nodeData.Recognition.Param.(*NodeTemplateMatchParam)
	require.Equal(a.t, []string{"test.png", "test2.png"}, param.Template)
	require.Equal(a.t, []float64{0.8}, param.Threshold)
	require.Equal(a.t, NodeTemplateMatchOrderByScore, param.OrderBy)
	require.Equal(a.t, 1, param.Index)
	require.Equal(a.t, NodeTemplateMatchMethodCCOEFF_NORMED, param.Method)
	require.True(a.t, param.GreenMask)
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

	nodeData, err := ctx.GetNodeData("test_feature")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeRecognitionTypeFeatureMatch, nodeData.Recognition.Type)
	require.IsType(a.t, (*NodeFeatureMatchParam)(nil), nodeData.Recognition.Param)

	param := nodeData.Recognition.Param.(*NodeFeatureMatchParam)
	require.Equal(a.t, []string{"feature.png"}, param.Template)
	require.Equal(a.t, 10, param.Count)
	require.Equal(a.t, NodeFeatureMatchOrderByArea, param.OrderBy)
	require.Equal(a.t, NodeFeatureMatchMethodSIFT, param.Detector)
	require.Equal(a.t, 0.7, param.Ratio)
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

	nodeData, err := ctx.GetNodeData("test_color")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeRecognitionTypeColorMatch, nodeData.Recognition.Type)
	require.IsType(a.t, (*NodeColorMatchParam)(nil), nodeData.Recognition.Param)

	param := nodeData.Recognition.Param.(*NodeColorMatchParam)
	require.Equal(a.t, NodeColorMatchMethodRGB, param.Method)
	require.Equal(a.t, [][]int{{0, 0, 0}}, param.Lower)
	require.Equal(a.t, [][]int{{255, 255, 255}}, param.Upper)
	require.Equal(a.t, 100, param.Count)
	require.True(a.t, param.Connected)
}

func (a *testContextGetNodeDataAct) testOCRRecognition(ctx *Context) {
	raw := map[string]any{
		"test_ocr": map[string]any{
			"recognition": map[string]any{
				"type": "OCR",
				"param": map[string]any{
					"roi":       []int{0, 0, 300, 100},
					"expected":  []string{"Hello", "World"},
					"threshold": 0.5,
					"replace":   [][]string{{"0", "O"}, {"1", "l"}},
					"order_by":  "Length",
					"index":     0,
					"only_rec":  true,
					"model":     "ppocr_v4",
				},
			},
		},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNodeData("test_ocr")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeRecognitionTypeOCR, nodeData.Recognition.Type)
	require.IsType(a.t, (*NodeOCRParam)(nil), nodeData.Recognition.Param)

	param := nodeData.Recognition.Param.(*NodeOCRParam)
	require.Equal(a.t, []string{"Hello", "World"}, param.Expected)
	require.Equal(a.t, 0.5, param.Threshold)
	require.Equal(a.t, NodeOCROrderByLength, param.OrderBy)
	require.True(a.t, param.OnlyRec)
	require.Equal(a.t, "ppocr_v4", param.Model)
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

	nodeData, err := ctx.GetNodeData("test_nn_classify")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeRecognitionTypeNeuralNetworkClassify, nodeData.Recognition.Type)
	require.IsType(a.t, (*NodeNeuralNetworkClassifyParam)(nil), nodeData.Recognition.Param)

	param := nodeData.Recognition.Param.(*NodeNeuralNetworkClassifyParam)
	require.Equal(a.t, []string{"Cat", "Dog", "Mouse"}, param.Labels)
	require.Equal(a.t, "classifier.onnx", param.Model)
	require.Equal(a.t, []int{0, 2}, param.Expected)
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

	nodeData, err := ctx.GetNodeData("test_nn_detect")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeRecognitionTypeNeuralNetworkDetect, nodeData.Recognition.Type)
	require.IsType(a.t, (*NodeNeuralNetworkDetectParam)(nil), nodeData.Recognition.Param)

	param := nodeData.Recognition.Param.(*NodeNeuralNetworkDetectParam)
	require.Equal(a.t, []string{"person", "car", "bicycle"}, param.Labels)
	require.Equal(a.t, "yolov8.onnx", param.Model)
	require.Equal(a.t, []int{0, 1}, param.Expected)
	require.Equal(a.t, NodeNeuralNetworkDetectOrderByArea, param.OrderBy)
	require.Equal(a.t, -1, param.Index)
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

	nodeData, err := ctx.GetNodeData("test_custom_rec")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeRecognitionTypeCustom, nodeData.Recognition.Type)
	require.IsType(a.t, (*NodeCustomRecognitionParam)(nil), nodeData.Recognition.Param)

	param := nodeData.Recognition.Param.(*NodeCustomRecognitionParam)
	require.Equal(a.t, "MyCustomRecognizer", param.CustomRecognition)
	require.NotNil(a.t, param.CustomRecognitionParam)
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

	nodeData, err := ctx.GetNodeData("test_do_nothing")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeDoNothing, nodeData.Action.Type)
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

	nodeData, err := ctx.GetNodeData("test_long_press")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeLongPress, nodeData.Action.Type)
	require.IsType(a.t, (*NodeLongPressParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*NodeLongPressParam)
	require.Equal(a.t, int64(2000), param.Duration)
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

	nodeData, err := ctx.GetNodeData("test_swipe")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeSwipe, nodeData.Action.Type)
	require.IsType(a.t, (*NodeSwipeParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*NodeSwipeParam)
	require.Equal(a.t, []int64{500}, param.Duration)
	require.Equal(a.t, []int64{100}, param.EndHold)
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

	nodeData, err := ctx.GetNodeData("test_multi_swipe")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeMultiSwipe, nodeData.Action.Type)
	require.IsType(a.t, (*NodeMultiSwipeParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*NodeMultiSwipeParam)
	require.Len(a.t, param.Swipes, 2)
	require.Equal(a.t, int64(0), param.Swipes[0].Starting)
	require.Equal(a.t, int64(100), param.Swipes[1].Starting)
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

	nodeData, err := ctx.GetNodeData("test_touch_down")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeTouchDown, nodeData.Action.Type)
	require.IsType(a.t, (*NodeTouchDownParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*NodeTouchDownParam)
	require.Equal(a.t, 50, param.Pressure)
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

	nodeData, err := ctx.GetNodeData("test_touch_move")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeTouchMove, nodeData.Action.Type)
	require.IsType(a.t, (*NodeTouchMoveParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*NodeTouchMoveParam)
	require.Equal(a.t, 30, param.Pressure)
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

	nodeData, err := ctx.GetNodeData("test_touch_up")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeTouchUp, nodeData.Action.Type)
	require.IsType(a.t, (*NodeTouchUpParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*NodeTouchUpParam)
	require.Equal(a.t, 1, param.Contact)
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

	nodeData, err := ctx.GetNodeData("test_click_key")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeClickKey, nodeData.Action.Type)
	require.IsType(a.t, (*NodeClickKeyParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*NodeClickKeyParam)
	require.Equal(a.t, []int{4, 66}, param.Key)
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

	nodeData, err := ctx.GetNodeData("test_long_press_key")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeLongPressKey, nodeData.Action.Type)
	require.IsType(a.t, (*NodeLongPressKeyParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*NodeLongPressKeyParam)
	require.Equal(a.t, []int{4}, param.Key)
	require.Equal(a.t, int64(1500), param.Duration)
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

	nodeData, err := ctx.GetNodeData("test_key_down")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeKeyDown, nodeData.Action.Type)
	require.IsType(a.t, (*NodeKeyDownParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*NodeKeyDownParam)
	require.Equal(a.t, 29, param.Key)
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

	nodeData, err := ctx.GetNodeData("test_key_up")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeKeyUp, nodeData.Action.Type)
	require.IsType(a.t, (*NodeKeyUpParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*NodeKeyUpParam)
	require.Equal(a.t, 29, param.Key)
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

	nodeData, err := ctx.GetNodeData("test_input_text")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeInputText, nodeData.Action.Type)
	require.IsType(a.t, (*NodeInputTextParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*NodeInputTextParam)
	require.Equal(a.t, "Hello World", param.InputText)
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

	nodeData, err := ctx.GetNodeData("test_start_app")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeStartApp, nodeData.Action.Type)
	require.IsType(a.t, (*NodeStartAppParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*NodeStartAppParam)
	require.Equal(a.t, "com.example.app/com.example.MainActivity", param.Package)
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

	nodeData, err := ctx.GetNodeData("test_stop_app")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeStopApp, nodeData.Action.Type)
	require.IsType(a.t, (*NodeStopAppParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*NodeStopAppParam)
	require.Equal(a.t, "com.example.app", param.Package)
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

	nodeData, err := ctx.GetNodeData("test_stop_task")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeStopTask, nodeData.Action.Type)
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

	nodeData, err := ctx.GetNodeData("test_command")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeCommand, nodeData.Action.Type)
	require.IsType(a.t, (*NodeCommandParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*NodeCommandParam)
	require.Equal(a.t, "python", param.Exec)
	require.Equal(a.t, []string{"{RESOURCE_DIR}/script.py", "{NODE}", "{IMAGE}"}, param.Args)
	require.True(a.t, param.Detach)
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

	nodeData, err := ctx.GetNodeData("test_scroll")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeScroll, nodeData.Action.Type)
	require.IsType(a.t, (*NodeScrollParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*NodeScrollParam)
	require.Equal(a.t, 100, param.Dx)
	require.Equal(a.t, 200, param.Dy)
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

	nodeData, err := ctx.GetNodeData("test_custom_act")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)
	require.Equal(a.t, NodeActionTypeCustom, nodeData.Action.Type)
	require.IsType(a.t, (*NodeCustomActionParam)(nil), nodeData.Action.Param)

	param := nodeData.Action.Param.(*NodeCustomActionParam)
	require.Equal(a.t, "MyCustomAction", param.CustomAction)
	require.NotNil(a.t, param.CustomActionParam)
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
			"anchor": []string{"MyAnchor", "AnotherAnchor"},
		},
		"NodeA": map[string]any{},
		"NodeB": map[string]any{},
		"NodeC": map[string]any{
			"anchor": []string{"AnchorX"},
		},
		"ErrorHandler": map[string]any{},
	}
	ctx.OverridePipeline(raw)

	nodeData, err := ctx.GetNodeData("test_attributes")
	require.NoError(a.t, err)
	require.NotNil(a.t, nodeData)

	// Check recognition and action
	require.Equal(a.t, NodeRecognitionTypeDirectHit, nodeData.Recognition.Type)
	require.Equal(a.t, NodeActionTypeClick, nodeData.Action.Type)

	// Check next list
	require.Len(a.t, nodeData.Next, 3)
	require.Equal(a.t, "NodeA", nodeData.Next[0].Name)
	require.False(a.t, nodeData.Next[0].JumpBack)
	require.Equal(a.t, "NodeB", nodeData.Next[1].Name)
	require.True(a.t, nodeData.Next[1].JumpBack)
	require.Equal(a.t, "AnchorX", nodeData.Next[2].Name)
	require.True(a.t, nodeData.Next[2].Anchor)

	// Check timing properties
	require.NotNil(a.t, nodeData.RateLimit)
	require.Equal(a.t, int64(500), *nodeData.RateLimit)
	require.NotNil(a.t, nodeData.Timeout)
	require.Equal(a.t, int64(30000), *nodeData.Timeout)

	// Check on_error
	require.Len(a.t, nodeData.OnError, 1)
	require.Equal(a.t, "ErrorHandler", nodeData.OnError[0].Name)

	// Check boolean and numeric properties
	require.True(a.t, nodeData.Inverse)
	require.NotNil(a.t, nodeData.Enabled)
	require.True(a.t, *nodeData.Enabled)
	require.NotNil(a.t, nodeData.MaxHit)
	require.Equal(a.t, uint64(5), *nodeData.MaxHit)

	// Check delays
	require.NotNil(a.t, nodeData.PreDelay)
	require.Equal(a.t, int64(100), *nodeData.PreDelay)
	require.NotNil(a.t, nodeData.PostDelay)
	require.Equal(a.t, int64(150), *nodeData.PostDelay)

	// Check wait freezes
	require.NotNil(a.t, nodeData.PreWaitFreezes)
	require.Equal(a.t, int64(1000), nodeData.PreWaitFreezes.Time)
	require.Equal(a.t, 0.95, nodeData.PreWaitFreezes.Threshold)
	require.NotNil(a.t, nodeData.PostWaitFreezes)
	require.Equal(a.t, int64(500), nodeData.PostWaitFreezes.Time)

	// Check focus
	require.NotNil(a.t, nodeData.Focus)

	// Check attach
	require.NotNil(a.t, nodeData.Attach)
	require.Equal(a.t, "custom_value", nodeData.Attach["custom_key"])

	// Check anchor
	require.Equal(a.t, []string{"MyAnchor", "AnotherAnchor"}, nodeData.Anchor)
}

func TestContext_GetNodeData(t *testing.T) {
	ctrl := NewBlankController()
	require.NotNil(t, ctrl)
	defer ctrl.Destroy()
	connected := ctrl.PostConnect().Wait().Success()
	require.True(t, connected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()

	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomAction("TestContext_GetNodeDataAct", &testContextGetNodeDataAct{t})
	require.True(t, ok)

	pipeline := NewPipeline()
	launchNode := NewNode("launch",
		WithAction(ActCustom("TestContext_GetNodeDataAct")),
	)
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
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomAction("TestContext_GetTaskJobAct", &testContextGetTaskJobAct{t})
	require.True(t, ok)

	pipeline := NewPipeline()
	testContext_GetTaskJobNode := NewNode("TestContext_GetTaskJob",
		WithAction(ActCustom("TestContext_GetTaskJobAct")),
	)
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
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomAction("TestContext_GetTaskerAct", &testContextGetTaskerAct{t})
	require.True(t, ok)

	pipeline := NewPipeline()
	testContext_GetTaskerNode := NewNode("TestContext_GetTasker",
		WithAction(ActCustom("TestContext_GetTaskerAct")),
	)
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
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomAction("TestContext_GetTaskerAct", &testContextCloneAct{t})
	require.True(t, ok)

	pipeline := NewPipeline()
	testContext_CloneNode := NewNode("TestContext_Clone",
		WithAction(ActCustom("TestContext_GetTaskerAct")),
	)
	pipeline.AddNode(testContext_CloneNode)

	got := tasker.PostTask(testContext_CloneNode.Name, pipeline).
		Wait().Success()
	require.True(t, got)
}
