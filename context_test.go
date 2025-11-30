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
			WithClickTarget(Rect{100, 100, 10, 10}),
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
		WithRecognition(RecOCR([]string{"Hello"})),
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
			WithClickTarget(Rect{100, 100, 10, 10}),
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
			WithClickTarget(Rect{100, 100, 10, 10}),
		)),
	)
	pipeline1.AddNode(testNode1)

	detail1 := ctx.RunTask(testNode1.Name, pipeline1)
	require.NotNil(t.t, detail1)

	pipeline2 := NewPipeline()
	testNode2 := NewNode("Test",
		WithAction(ActClick(
			WithClickTarget(Rect{200, 200, 10, 10}),
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
