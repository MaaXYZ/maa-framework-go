package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func createResource(t *testing.T) *Resource {
	res := NewResource()
	require.NotNil(t, res)
	return res
}

func TestNewResource(t *testing.T) {
	res := createResource(t)
	res.Destroy()
}

type testResourceTestRec struct{}

func (t *testResourceTestRec) Run(_ *Context, _ *CustomRecognitionArg) (*CustomRecognitionResult, bool) {
	return &CustomRecognitionResult{}, true
}

func TestResource_RegisterCustomRecognition(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	got1 := res.RegisterCustomRecognition("TestRec", &testResourceTestRec{})
	require.True(t, got1)

	pipeline := NewPipeline()
	testResource_RegisterCustomRecognitionNode := NewNode("TestResource_RegisterCustomRecognition",
		WithRecognition(RecCustom("TestRec")),
	)
	pipeline.AddNode(testResource_RegisterCustomRecognitionNode)

	got2 := tasker.PostTask(testResource_RegisterCustomRecognitionNode.Name, pipeline).
		Wait().Success()
	require.True(t, got2)
}

func TestResource_UnregisterCustomRecognition(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	got1 := res.RegisterCustomRecognition("TestRec", &testResourceTestRec{})
	require.True(t, got1)

	pipeline := NewPipeline()
	testResource_UnregisterCustomRecognitionNode := NewNode("TestResource_UnregisterCustomRecognition",
		WithRecognition(RecCustom("TestRec")),
	)
	pipeline.AddNode(testResource_UnregisterCustomRecognitionNode)

	got2 := tasker.PostTask(testResource_UnregisterCustomRecognitionNode.Name, pipeline).
		Wait().Success()
	require.True(t, got2)

	got3 := res.UnregisterCustomRecognition("TestRec")
	require.True(t, got3)

	got4 := tasker.PostTask(testResource_UnregisterCustomRecognitionNode.Name, pipeline).
		Wait().Failure()
	require.True(t, got4)
}

func TestResource_ClearCustomRecognition(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	got1 := res.RegisterCustomRecognition("TestRec1", &testResourceTestRec{})
	require.True(t, got1)
	got2 := res.RegisterCustomRecognition("TestRec2", &testResourceTestRec{})
	require.True(t, got2)

	pipeline1 := NewPipeline()
	testResource_ClearCustomRecognitionNode1 := NewNode("TestResource_ClearCustomRecognition",
		WithRecognition(RecCustom("TestRec1")),
	)
	pipeline1.AddNode(testResource_ClearCustomRecognitionNode1)

	pipeline2 := NewPipeline()
	testResource_ClearCustomRecognitionNode2 := NewNode("TestResource_ClearCustomRecognition",
		WithRecognition(RecCustom("TestRec2")),
	)
	pipeline2.AddNode(testResource_ClearCustomRecognitionNode2)
	got3 := tasker.PostTask(testResource_ClearCustomRecognitionNode1.Name, pipeline1).
		Wait().Success()
	require.True(t, got3)
	got4 := tasker.PostTask(testResource_ClearCustomRecognitionNode2.Name, pipeline2).
		Wait().Success()
	require.True(t, got4)

	got5 := res.ClearCustomRecognition()
	require.True(t, got5)

	got6 := tasker.PostTask(testResource_ClearCustomRecognitionNode1.Name, pipeline1).
		Wait().Failure()
	require.True(t, got6)
	got7 := tasker.PostTask(testResource_ClearCustomRecognitionNode2.Name, pipeline2).
		Wait().Failure()
	require.True(t, got7)
}

type testResourceTestAct struct{}

func (t *testResourceTestAct) Run(_ *Context, _ *CustomActionArg) bool {
	return true
}

func TestResource_RegisterCustomAction(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	registered := res.RegisterCustomAction("TestAct", &testResourceTestAct{})
	require.True(t, registered)

	pipeline := NewPipeline()
	testResource_RegisterCustomActionNode := NewNode("TestResource_RegisterCustomAction",
		WithAction(ActCustom("TestAct")),
	)
	pipeline.AddNode(testResource_RegisterCustomActionNode)

	got := tasker.PostTask(testResource_RegisterCustomActionNode.Name, pipeline).
		Wait().Success()
	require.True(t, got)
}

func TestResource_UnregisterCustomAction(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	registered := res.RegisterCustomAction("TestAct", &testResourceTestAct{})
	require.True(t, registered)

	pipeline := NewPipeline()
	testResource_UnregisterCustomActionNode := NewNode("TestResource_UnregisterCustomAction",
		WithAction(ActCustom("TestAct")),
	)
	pipeline.AddNode(testResource_UnregisterCustomActionNode)

	got1 := tasker.PostTask(testResource_UnregisterCustomActionNode.Name, pipeline).
		Wait().Success()
	require.True(t, got1)

	unregistered := res.UnregisterCustomAction("TestAct")
	require.True(t, unregistered)

	got2 := tasker.PostTask(testResource_UnregisterCustomActionNode.Name, pipeline).
		Wait().Failure()
	require.True(t, got2)
}

func TestResource_ClearCustomAction(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	registered1 := res.RegisterCustomAction("TestAct1", &testResourceTestAct{})
	require.True(t, registered1)
	registered2 := res.RegisterCustomAction("TestAct2", &testResourceTestAct{})
	require.True(t, registered2)

	pipeline1 := NewPipeline()
	testResource_ClearCustomActionNode1 := NewNode("TestResource_ClearCustomAction",
		WithAction(ActCustom("TestAct1")),
	)
	pipeline1.AddNode(testResource_ClearCustomActionNode1)

	pipeline2 := NewPipeline()
	testResource_ClearCustomActionNode2 := NewNode("TestResource_ClearCustomAction",
		WithAction(ActCustom("TestAct2")),
	)
	pipeline2.AddNode(testResource_ClearCustomActionNode2)

	got1 := tasker.PostTask(testResource_ClearCustomActionNode1.Name, pipeline1).
		Wait().Success()
	require.True(t, got1)
	got2 := tasker.PostTask(testResource_ClearCustomActionNode2.Name, pipeline2).
		Wait().Success()
	require.True(t, got2)

	cleared := res.ClearCustomAction()
	require.True(t, cleared)

	got3 := tasker.PostTask(testResource_ClearCustomActionNode1.Name, pipeline1).
		Wait().Failure()
	require.True(t, got3)
	got4 := tasker.PostTask(testResource_ClearCustomActionNode2.Name, pipeline2).
		Wait().Failure()
	require.True(t, got4)
}

func TestResource_PostPath(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)
}

func TestResource_Clear(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)
	cleared := res.Clear()
	require.True(t, cleared)
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
	hash, ok := res.GetHash()
	require.True(t, ok)
	require.NotEqual(t, "0", hash)
}

func TestResource_GetNodeList(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)
	taskList, ok := res.GetNodeList()
	require.True(t, ok)
	require.NotEmpty(t, taskList)
}
