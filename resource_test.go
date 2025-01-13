package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func createResource(t *testing.T, notify Notification) *Resource {
	res := NewResource(notify)
	require.NotNil(t, res)
	return res
}

func TestNewResource(t *testing.T) {
	res := createResource(t, nil)
	res.Destroy()
}

type testResourceTestRec struct{}

func (t *testResourceTestRec) Run(_ *Context, _ *CustomRecognitionArg) (*CustomRecognitionResult, bool) {
	return &CustomRecognitionResult{}, true
}

func TestResource_RegisterCustomRecognition(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	got1 := res.RegisterCustomRecognition("TestRec", &testResourceTestRec{})
	require.True(t, got1)

	got2 := tasker.PostTask("TestResource_RegisterCustomRecognition", J{
		"TestResource_RegisterCustomRecognition": J{
			"recognition":        "custom",
			"custom_recognition": "TestRec",
		},
	}).Wait().Success()
	require.True(t, got2)
}

func TestResource_UnregisterCustomRecognition(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	got1 := res.RegisterCustomRecognition("TestRec", &testResourceTestRec{})
	require.True(t, got1)

	got2 := tasker.PostTask("TestResource_UnregisterCustomRecognition", J{
		"TestResource_UnregisterCustomRecognition": J{
			"recognition":        "custom",
			"custom_recognition": "TestRec",
		},
	}).Wait().Success()
	require.True(t, got2)

	got3 := res.UnregisterCustomRecognition("TestRec")
	require.True(t, got3)

	got4 := tasker.PostTask("TestResource_UnregisterCustomRecognition", J{
		"TestResource_UnregisterCustomRecognition": J{
			"recognition":        "custom",
			"custom_recognition": "TestRec",
		},
	}).Wait().Failure()
	require.True(t, got4)
}

func TestResource_ClearCustomRecognition(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	got1 := res.RegisterCustomRecognition("TestRec1", &testResourceTestRec{})
	require.True(t, got1)
	got2 := res.RegisterCustomRecognition("TestRec2", &testResourceTestRec{})
	require.True(t, got2)

	got3 := tasker.PostTask("TestResource_ClearCustomRecognition", J{
		"TestResource_ClearCustomRecognition": J{
			"recognition":        "custom",
			"custom_recognition": "TestRec1",
		},
	}).Wait().Success()
	require.True(t, got3)
	got4 := tasker.PostTask("TestResource_ClearCustomRecognition", J{
		"TestResource_ClearCustomRecognition": J{
			"recognition":        "custom",
			"custom_recognition": "TestRec2",
		},
	}).Wait().Success()
	require.True(t, got4)

	got5 := res.ClearCustomRecognition()
	require.True(t, got5)

	got6 := tasker.PostTask("TestResource_ClearCustomRecognition", J{
		"TestResource_ClearCustomRecognition": J{
			"recognition":        "custom",
			"custom_recognition": "TestRec1",
		},
	}).Wait().Failure()
	require.True(t, got6)
	got7 := tasker.PostTask("TestResource_ClearCustomRecognition", J{
		"TestResource_ClearCustomRecognition": J{
			"recognition":        "custom",
			"custom_recognition": "TestRec2",
		},
	}).Wait().Failure()
	require.True(t, got7)
}

type testResourceTestAct struct{}

func (t *testResourceTestAct) Run(_ *Context, _ *CustomActionArg) bool {
	return true
}

func TestResource_RegisterCustomAction(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	registered := res.RegisterCustomAction("TestAct", &testResourceTestAct{})
	require.True(t, registered)

	got := tasker.PostTask("TestResource_RegisterCustomAction", J{
		"TestResource_RegisterCustomAction": J{
			"action":        "custom",
			"custom_action": "TestAct",
		},
	}).Wait().Success()
	require.True(t, got)
}

func TestResource_UnregisterCustomAction(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	registered := res.RegisterCustomAction("TestAct", &testResourceTestAct{})
	require.True(t, registered)

	got1 := tasker.PostTask("TestResource_RegisterCustomAction", J{
		"TestResource_RegisterCustomAction": J{
			"action":        "custom",
			"custom_action": "TestAct",
		},
	}).Wait().Success()
	require.True(t, got1)

	unregistered := res.UnregisterCustomAction("TestAct")
	require.True(t, unregistered)

	got2 := tasker.PostTask("TestResource_RegisterCustomAction", J{
		"TestResource_RegisterCustomAction": J{
			"action":        "custom",
			"custom_action": "TestAct",
		},
	}).Wait().Failure()
	require.True(t, got2)
}

func TestResource_ClearCustomAction(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	registered1 := res.RegisterCustomAction("TestAct1", &testResourceTestAct{})
	require.True(t, registered1)
	registered2 := res.RegisterCustomAction("TestAct2", &testResourceTestAct{})
	require.True(t, registered2)

	got1 := tasker.PostTask("TestResource_RegisterCustomAction", J{
		"TestResource_RegisterCustomAction": J{
			"action":        "custom",
			"custom_action": "TestAct1",
		},
	}).Wait().Success()
	require.True(t, got1)
	got2 := tasker.PostTask("TestResource_RegisterCustomAction", J{
		"TestResource_RegisterCustomAction": J{
			"action":        "custom",
			"custom_action": "TestAct2",
		},
	}).Wait().Success()
	require.True(t, got2)

	cleared := res.ClearCustomAction()
	require.True(t, cleared)

	got3 := tasker.PostTask("TestResource_RegisterCustomAction", J{
		"TestResource_RegisterCustomAction": J{
			"action":        "custom",
			"custom_action": "TestAct1",
		},
	}).Wait().Failure()
	require.True(t, got3)
	got4 := tasker.PostTask("TestResource_RegisterCustomAction", J{
		"TestResource_RegisterCustomAction": J{
			"action":        "custom",
			"custom_action": "TestAct2",
		},
	}).Wait().Failure()
	require.True(t, got4)
}

func TestResource_PostPath(t *testing.T) {
	res := createResource(t, nil)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)
}

func TestResource_Clear(t *testing.T) {
	res := createResource(t, nil)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)
	cleared := res.Clear()
	require.True(t, cleared)
}

func TestResource_Loaded(t *testing.T) {
	res := createResource(t, nil)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)
	loaded := res.Loaded()
	require.True(t, loaded)
}

func TestResource_GetHash(t *testing.T) {
	res := createResource(t, nil)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)
	hash, ok := res.GetHash()
	require.True(t, ok)
	require.NotEqual(t, "0", hash)
}

func TestResource_GetTaskList(t *testing.T) {
	res := createResource(t, nil)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)
	taskList, ok := res.GetTaskList()
	require.True(t, ok)
	require.NotEmpty(t, taskList)
}
