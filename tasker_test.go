package maa

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createTasker(t *testing.T) *Tasker {
	tasker := NewTasker()
	require.NotNil(t, tasker)
	return tasker
}

func taskerBind(t *testing.T, tasker *Tasker, ctrl *Controller, res *Resource) {
	isResBound := tasker.BindResource(res)
	require.True(t, isResBound)
	isCtrlBound := tasker.BindController(ctrl)
	require.True(t, isCtrlBound)
}

func TestNewTasker(t *testing.T) {
	tasker := createTasker(t)
	tasker.Destroy()
}

func TestTasker_BindResource(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	bound := tasker.BindResource(res)
	require.True(t, bound)
}

func TestTasker_BindController(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	bound := tasker.BindController(ctrl)
	require.True(t, bound)
}

func TestTasker_Initialized(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	connected := ctrl.PostConnect().Wait().Success()
	require.True(t, connected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	initialized := tasker.Initialized()
	require.True(t, initialized)
}

func TestTasker_PostPipeline(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	pipeline := NewPipeline()
	testTasker_PostPipelineNode := NewNode("TestTasker_PostPipeline",
		WithAction(ActClick(
			WithClickTarget(NewTargetRect(Rect{100, 200, 100, 100})),
		)),
	)
	pipeline.AddNode(testTasker_PostPipelineNode)

	taskJob := tasker.PostTask(testTasker_PostPipelineNode.Name, pipeline)
	got := taskJob.Wait().Success()
	require.True(t, got)
	detail := taskJob.GetDetail()
	t.Logf("%#v", detail)
}

func TestTasker_Running(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	got := tasker.Running()
	require.False(t, got)
}

func TestTasker_PostStop(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	got := tasker.PostStop()
	require.NotNil(t, got)
}

func TestTasker_GetResource(t *testing.T) {
	res1 := createResource(t)
	defer res1.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	bound := tasker.BindResource(res1)
	require.True(t, bound)

	res2 := tasker.GetResource()
	require.NotNil(t, res2)
	require.Equal(t, res1.handle, res2.handle)
}

func TestTasker_GetController(t *testing.T) {
	ctrl1 := createCarouselImageController(t)
	defer ctrl1.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	bound := tasker.BindController(ctrl1)
	require.True(t, bound)

	ctrl2 := tasker.GetController()
	require.NotNil(t, ctrl2)
	require.Equal(t, ctrl1.handle, ctrl2.handle)
}

func TestTasker_ClearCache(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	got := tasker.ClearCache()
	require.True(t, got)
}

func TestTasker_GetLatestNode(t *testing.T) {
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
	job := tasker.PostTask("Wilderness")
	require.NotNil(t, job)
	time.Sleep(2 * time.Second)
	detail := tasker.GetLatestNode("Wilderness")
	t.Log(detail)
	got := job.Wait().Success()
	require.True(t, got)
}

func TestTasker_OverridePipeline(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	pipeline := NewPipeline()
	testNode := NewNode("TestTasker_OverridePipeline",
		WithAction(ActClick(
			WithClickTarget(NewTargetRect(Rect{100, 200, 100, 100})),
		)),
	)
	pipeline.AddNode(testNode)

	// Start a task
	taskJob := tasker.PostTask(testNode.Name, pipeline)

	// Override the pipeline while task is running
	overridePipeline := NewPipeline()
	overrideNode := NewNode("TestTasker_OverridePipeline",
		WithAction(ActClick(
			WithClickTarget(NewTargetRect(Rect{200, 300, 100, 100})),
		)),
	)
	overridePipeline.AddNode(overrideNode)

	// Test OverridePipeline with a Pipeline object
	got := taskJob.OverridePipeline(overridePipeline)
	// Note: OverridePipeline may return false if the task has already completed
	// The important thing is that it doesn't panic and executes correctly
	t.Logf("OverridePipeline result: %v", got)

	// Wait for task to complete
	success := taskJob.Wait().Success()
	require.True(t, success)
}
