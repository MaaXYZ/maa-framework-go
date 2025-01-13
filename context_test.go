package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testContextRunTaskAct struct {
	t *testing.T
}

func (t *testContextRunTaskAct) Run(ctx *Context, _ *CustomActionArg) bool {
	detail := ctx.RunTask("Test", J{
		"Test": J{
			"action": "Click",
			"target": []int{100, 100, 10, 10},
		},
	})
	require.NotNil(t.t, detail)
	return true
}

func TestContext_RunTask(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomAction("TestContext_RunPipelineAct", &testContextRunTaskAct{t})
	require.True(t, ok)

	got := tasker.PostTask("TestContext_RunPipeline", J{
		"TestContext_RunPipeline": J{
			"action":        "Custom",
			"custom_action": "TestContext_RunPipelineAct",
		},
	}).Wait().Success()
	require.True(t, got)
}

type testContextRunRecognitionAct struct {
	t *testing.T
}

func (t *testContextRunRecognitionAct) Run(ctx *Context, _ *CustomActionArg) bool {
	img := ctx.GetTasker().GetController().CacheImage()
	require.NotNil(t.t, img)
	_ = ctx.RunRecognition("Test", img, J{
		"Test": J{
			"recognition": "OCR",
			"expected":    "Hello",
		},
	})
	return true
}

func TestContext_RunRecognition(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomAction("TestContext_RunRecognitionAct", &testContextRunRecognitionAct{t})
	require.True(t, ok)

	got := tasker.PostTask("TestContext_RunRecognition", J{
		"TestContext_RunRecognition": J{
			"next": []string{
				"RunRecognition",
				"Stop",
			},
		},
		"RunRecognition": J{
			"action":        "Custom",
			"custom_action": "TestContext_RunRecognitionAct",
		},
		"Stop": J{},
	}).Wait().Success()
	require.True(t, got)
}

type testContextRunActionAct struct {
	t *testing.T
}

func (a testContextRunActionAct) Run(ctx *Context, arg *CustomActionArg) bool {
	detail := ctx.RunAction("Test", arg.Box, arg.RecognitionDetail.DetailJson, J{
		"Test": J{
			"action": "Click",
			"target": []int{100, 100, 10, 10},
		},
	})
	require.NotNil(a.t, detail)
	return true
}

func TestContext_RunAction(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomAction("TestContext_RunActionAct", &testContextRunActionAct{t})
	require.True(t, ok)

	got := tasker.PostTask("TestContext_RunAction", J{
		"TestContext_RunAction": J{
			"action":        "Custom",
			"custom_action": "TestContext_RunActionAct",
		},
	}).Wait().Success()
	require.True(t, got)
}

type testContextOverriderPipelineAct struct {
	t *testing.T
}

func (t *testContextOverriderPipelineAct) Run(ctx *Context, _ *CustomActionArg) bool {
	detail1 := ctx.RunTask("Test", J{
		"Test": J{
			"action": "Click",
			"target": []int{100, 100, 10, 10},
		},
	})
	require.NotNil(t.t, detail1)

	ok := ctx.OverridePipeline(J{
		"Test": J{
			"target": []int{200, 200, 10, 10},
		},
	})
	require.True(t.t, ok)

	detail2 := ctx.RunTask("Test")
	require.NotNil(t.t, detail2)
	return true
}

func TestContext_OverridePipeline(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomAction("TestContext_OverridePipelineAct", &testContextOverriderPipelineAct{t})
	require.True(t, ok)

	got := tasker.PostTask("TestContext_OverridePipeline", J{
		"TestContext_OverridePipeline": J{
			"action":        "Custom",
			"custom_action": "TestContext_OverridePipelineAct",
		},
	}).Wait().Success()
	require.True(t, got)
}

type testContextOverrideNextAct struct {
	t *testing.T
}

func (t *testContextOverrideNextAct) Run(ctx *Context, _ *CustomActionArg) bool {
	ok1 := ctx.OverridePipeline(J{
		"Test": J{
			"next": "TaskA",
		},
		"TaskA": J{},
		"TaskB": J{},
	})
	require.True(t.t, ok1)

	ok2 := ctx.OverrideNext("Test", []string{"TaskB"})
	require.True(t.t, ok2)

	detail := ctx.RunTask("Test")
	require.NotNil(t.t, detail)
	return true
}

func TestContext_OverrideNext(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomAction("TestContext_OverrideNextAct", &testContextOverrideNextAct{t})
	require.True(t, ok)

	got := tasker.PostTask("TestContext_OverrideNext", J{
		"TestContext_OverrideNext": J{
			"action":        "Custom",
			"custom_action": "TestContext_OverrideNextAct",
		},
	}).Wait().Success()
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
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomAction("TestContext_GetTaskJobAct", &testContextGetTaskJobAct{t})
	require.True(t, ok)

	got := tasker.PostTask("TestContext_GetTaskJob", J{
		"TestContext_GetTaskJob": J{
			"action":        "Custom",
			"custom_action": "TestContext_GetTaskJobAct",
		},
	}).Wait().Success()
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
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomAction("TestContext_GetTaskerAct", &testContextGetTaskerAct{t})
	require.True(t, ok)

	got := tasker.PostTask("TestContext_GetTasker", J{
		"TestContext_GetTasker": J{
			"action":        "Custom",
			"custom_action": "TestContext_GetTaskerAct",
		},
	}).Wait().Success()
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
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	ok := res.RegisterCustomAction("TestContext_GetTaskerAct", &testContextCloneAct{t})
	require.True(t, ok)

	got := tasker.PostTask("TestContext_GetTasker", J{
		"TestContext_GetTasker": J{
			"action":        "Custom",
			"custom_action": "TestContext_GetTaskerAct",
		},
	}).Wait().Success()
	require.True(t, got)
}
