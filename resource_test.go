package maa

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createResource(t *testing.T) *Resource {
	res, err := NewResource()
	require.NoError(t, err)
	require.NotNil(t, res)
	return res
}

func TestNewResource(t *testing.T) {
	res := createResource(t)
	res.Destroy()
}

type testResourceTestRec struct {
	run func(*Context, *CustomRecognitionArg) (*CustomRecognitionResult, bool)
}

func (t *testResourceTestRec) Run(ctx *Context, arg *CustomRecognitionArg) (*CustomRecognitionResult, bool) {
	if t.run != nil {
		return t.run(ctx, arg)
	}
	return &CustomRecognitionResult{}, true
}

func TestResource_RegisterCustomRecognition(t *testing.T) {
	testCases := []struct {
		name      string
		newRunner func(*int32) CustomRecognitionRunner
	}{
		{
			name: "struct runner",
			newRunner: func(calls *int32) CustomRecognitionRunner {
				return &testResourceTestRec{
					run: func(_ *Context, _ *CustomRecognitionArg) (*CustomRecognitionResult, bool) {
						atomic.AddInt32(calls, 1)
						return &CustomRecognitionResult{
							Box:    Rect{10, 20, 30, 40},
							Detail: "struct runner",
						}, true
					},
				}
			},
		},
		{
			name: "func adapter",
			newRunner: func(calls *int32) CustomRecognitionRunner {
				return CustomRecognitionFunc(func(_ *Context, _ *CustomRecognitionArg) (*CustomRecognitionResult, bool) {
					atomic.AddInt32(calls, 1)
					return &CustomRecognitionResult{
						Box:    Rect{50, 60, 70, 80},
						Detail: "func adapter",
					}, true
				})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := createBlankController(t)
			defer ctrl.Destroy()
			isConnected := ctrl.PostConnect().Wait().Success()
			require.True(t, isConnected)

			res := createResource(t)
			defer res.Destroy()

			tasker := createTasker(t)
			defer tasker.Destroy()
			taskerBind(t, tasker, ctrl, res)

			var calls int32
			err := res.RegisterCustomRecognition("TestRec", tc.newRunner(&calls))
			require.NoError(t, err)

			pipeline := NewPipeline()
			node := NewNode("TestResource_RegisterCustomRecognition").
				SetRecognition(RecCustom(CustomRecognitionParam{CustomRecognition: "TestRec"}))
			pipeline.AddNode(node)

			got := tasker.PostTask(node.Name, pipeline).Wait().Success()
			require.True(t, got)
			require.NotZero(t, atomic.LoadInt32(&calls))
		})
	}
}

func TestResource_UnregisterCustomRecognition(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	err := res.RegisterCustomRecognition("TestRec", &testResourceTestRec{})
	require.NoError(t, err)

	pipeline := NewPipeline()
	testResource_UnregisterCustomRecognitionNode := NewNode("TestResource_UnregisterCustomRecognition").
		SetRecognition(RecCustom(CustomRecognitionParam{CustomRecognition: "TestRec"})).
		SetTimeout(0 * time.Second)
	pipeline.AddNode(testResource_UnregisterCustomRecognitionNode)

	got2 := tasker.PostTask(testResource_UnregisterCustomRecognitionNode.Name, pipeline).
		Wait().Success()
	require.True(t, got2)

	err = res.UnregisterCustomRecognition("TestRec")
	require.NoError(t, err)

	got4 := tasker.PostTask(testResource_UnregisterCustomRecognitionNode.Name, pipeline).
		Wait().Failure()
	require.True(t, got4)
}

func TestResource_ClearCustomRecognition(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	err := res.RegisterCustomRecognition("TestRec1", &testResourceTestRec{})
	require.NoError(t, err)
	err = res.RegisterCustomRecognition("TestRec2", &testResourceTestRec{})
	require.NoError(t, err)

	pipeline1 := NewPipeline()
	testResource_ClearCustomRecognitionNode1 := NewNode("TestResource_ClearCustomRecognition").
		SetRecognition(RecCustom(CustomRecognitionParam{CustomRecognition: "TestRec1"})).
		SetTimeout(0 * time.Second)
	pipeline1.AddNode(testResource_ClearCustomRecognitionNode1)

	pipeline2 := NewPipeline()
	testResource_ClearCustomRecognitionNode2 := NewNode("TestResource_ClearCustomRecognition").
		SetRecognition(RecCustom(CustomRecognitionParam{CustomRecognition: "TestRec2"})).
		SetTimeout(0 * time.Second)
	pipeline2.AddNode(testResource_ClearCustomRecognitionNode2)
	got3 := tasker.PostTask(testResource_ClearCustomRecognitionNode1.Name, pipeline1).
		Wait().Success()
	require.True(t, got3)
	got4 := tasker.PostTask(testResource_ClearCustomRecognitionNode2.Name, pipeline2).
		Wait().Success()
	require.True(t, got4)

	err = res.ClearCustomRecognition()
	require.NoError(t, err)

	got6 := tasker.PostTask(testResource_ClearCustomRecognitionNode1.Name, pipeline1).
		Wait().Failure()
	require.True(t, got6)
	got7 := tasker.PostTask(testResource_ClearCustomRecognitionNode2.Name, pipeline2).
		Wait().Failure()
	require.True(t, got7)
}

type testResourceTestAct struct {
	run func(*Context, *CustomActionArg) bool
}

func (t *testResourceTestAct) Run(ctx *Context, arg *CustomActionArg) bool {
	if t.run != nil {
		return t.run(ctx, arg)
	}
	return true
}

func TestResource_RegisterCustomAction(t *testing.T) {
	testCases := []struct {
		name      string
		newRunner func(*int32) CustomActionRunner
	}{
		{
			name: "struct runner",
			newRunner: func(calls *int32) CustomActionRunner {
				return &testResourceTestAct{
					run: func(_ *Context, _ *CustomActionArg) bool {
						atomic.AddInt32(calls, 1)
						return true
					},
				}
			},
		},
		{
			name: "func adapter",
			newRunner: func(calls *int32) CustomActionRunner {
				return CustomActionFunc(func(_ *Context, _ *CustomActionArg) bool {
					atomic.AddInt32(calls, 1)
					return true
				})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := createBlankController(t)
			defer ctrl.Destroy()
			isConnected := ctrl.PostConnect().Wait().Success()
			require.True(t, isConnected)

			res := createResource(t)
			defer res.Destroy()

			tasker := createTasker(t)
			defer tasker.Destroy()
			taskerBind(t, tasker, ctrl, res)

			var calls int32
			err := res.RegisterCustomAction("TestAct", tc.newRunner(&calls))
			require.NoError(t, err)

			pipeline := NewPipeline()
			node := NewNode("TestResource_RegisterCustomAction").
				SetAction(ActCustom(CustomActionParam{CustomAction: "TestAct"}))
			pipeline.AddNode(node)

			got := tasker.PostTask(node.Name, pipeline).Wait().Success()
			require.True(t, got)
			require.NotZero(t, atomic.LoadInt32(&calls))
		})
	}
}

func TestResource_UnregisterCustomAction(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	err := res.RegisterCustomAction("TestAct", &testResourceTestAct{})
	require.NoError(t, err)

	pipeline := NewPipeline()
	testResource_UnregisterCustomActionNode := NewNode("TestResource_UnregisterCustomAction").
		SetAction(ActCustom(CustomActionParam{CustomAction: "TestAct"}))
	pipeline.AddNode(testResource_UnregisterCustomActionNode)

	got1 := tasker.PostTask(testResource_UnregisterCustomActionNode.Name, pipeline).
		Wait().Success()
	require.True(t, got1)

	err = res.UnregisterCustomAction("TestAct")
	require.NoError(t, err)

	got2 := tasker.PostTask(testResource_UnregisterCustomActionNode.Name, pipeline).
		Wait().Failure()
	require.True(t, got2)
}

func TestResource_ClearCustomAction(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	err := res.RegisterCustomAction("TestAct1", &testResourceTestAct{})
	require.NoError(t, err)
	err = res.RegisterCustomAction("TestAct2", &testResourceTestAct{})
	require.NoError(t, err)

	pipeline1 := NewPipeline()
	testResource_ClearCustomActionNode1 := NewNode("TestResource_ClearCustomAction").
		SetAction(ActCustom(CustomActionParam{CustomAction: "TestAct1"}))
	pipeline1.AddNode(testResource_ClearCustomActionNode1)

	pipeline2 := NewPipeline()
	testResource_ClearCustomActionNode2 := NewNode("TestResource_ClearCustomAction").
		SetAction(ActCustom(CustomActionParam{CustomAction: "TestAct2"}))
	pipeline2.AddNode(testResource_ClearCustomActionNode2)

	got1 := tasker.PostTask(testResource_ClearCustomActionNode1.Name, pipeline1).
		Wait().Success()
	require.True(t, got1)
	got2 := tasker.PostTask(testResource_ClearCustomActionNode2.Name, pipeline2).
		Wait().Success()
	require.True(t, got2)

	err = res.ClearCustomAction()
	require.NoError(t, err)

	got3 := tasker.PostTask(testResource_ClearCustomActionNode1.Name, pipeline1).
		Wait().Failure()
	require.True(t, got3)
	got4 := tasker.PostTask(testResource_ClearCustomActionNode2.Name, pipeline2).
		Wait().Failure()
	require.True(t, got4)
}

func TestResource_PostBundle(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)
}

func TestResource_OverrideNext(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)

	override := []NextItem{
		{Name: "StartGame"},
		{Name: "Sub_BackButton", JumpBack: true},
		{Name: "HomeFlag", Anchor: true},
	}
	err := res.OverrideNext("StartUp", override)
	require.NoError(t, err)

	node, err := res.GetNode("StartUp")
	require.NoError(t, err)
	require.NotNil(t, node)
	require.Len(t, node.Next, len(override))

	findNextItem := func(name string) NextItem {
		for _, item := range node.Next {
			if item.Name == name {
				return item
			}
		}
		require.FailNowf(t, "next item not found", "name=%s", name)
		return NextItem{}
	}

	startGame := findNextItem("StartGame")
	require.False(t, startGame.JumpBack)
	require.False(t, startGame.Anchor)

	backButton := findNextItem("Sub_BackButton")
	require.True(t, backButton.JumpBack)
	require.False(t, backButton.Anchor)

	homeFlag := findNextItem("HomeFlag")
	require.False(t, homeFlag.JumpBack)
	require.True(t, homeFlag.Anchor)
}

func TestResource_Clear(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)
	err := res.Clear()
	require.NoError(t, err)
}

func TestResource_Loaded(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)
	loaded := res.Loaded()
	require.True(t, loaded)
}

func TestResource_GetHash(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)
	hash, err := res.GetHash()
	require.NoError(t, err)
	require.NotEqual(t, "0", hash)
}

func TestResource_GetNodeList(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)
	taskList, err := res.GetNodeList()
	require.NoError(t, err)
	require.NotEmpty(t, taskList)
}

func TestResource_GetNode(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)

	node, err := res.GetNode("StartGame")
	require.NoError(t, err)
	require.NotNil(t, node)
	require.NotNil(t, node.Recognition)
	require.NotNil(t, node.Action)
}

func TestResource_GetDefaultRecognitionParam(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()

	// Test getting default recognition parameters for DirectHit type
	param, err := res.GetDefaultRecognitionParam(RecognitionTypeDirectHit)
	require.NoError(t, err)
	require.NotNil(t, param)
	_, isDirectHit := param.(*DirectHitParam)
	require.True(t, isDirectHit, "param should be *NodeDirectHitParam")

	// Test getting default recognition parameters for OCR type
	param2, err := res.GetDefaultRecognitionParam(RecognitionTypeOCR)
	require.NoError(t, err)
	require.NotNil(t, param2)
	_, isOCR := param2.(*OCRParam)
	require.True(t, isOCR, "param2 should be *OCRParam")
}

func TestResource_GetDefaultActionParam(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()

	// Test getting default action parameters for Click type
	param, err := res.GetDefaultActionParam(ActionTypeClick)
	require.NoError(t, err)
	require.NotNil(t, param)
	_, isClick := param.(*ClickParam)
	require.True(t, isClick, "param should be *ClickParam")

	// Test getting default action parameters for DoNothing type
	param2, err := res.GetDefaultActionParam(ActionTypeDoNothing)
	require.NoError(t, err)
	require.NotNil(t, param2)
	_, isDoNothing := param2.(*DoNothingParam)
	require.True(t, isDoNothing, "param2 should be *DoNothingParam")

	// Test getting default action parameters for Screencap type
	param3, err := res.GetDefaultActionParam(ActionTypeScreencap)
	require.NoError(t, err)
	require.NotNil(t, param3)
	_, isScreencap := param3.(*ScreencapParam)
	require.True(t, isScreencap, "param3 should be *ScreencapParam")
}
