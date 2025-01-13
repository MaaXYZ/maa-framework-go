package maa

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createTasker(t *testing.T, notify Notification) *Tasker {
	tasker := NewTasker(notify)
	require.NotNil(t, tasker)
	return tasker
}

func taskerBind(t *testing.T, tasker *Tasker, ctrl Controller, res *Resource) {
	isResBound := tasker.BindResource(res)
	require.True(t, isResBound)
	isCtrlBound := tasker.BindController(ctrl)
	require.True(t, isCtrlBound)
}

func TestNewTasker(t *testing.T) {
	tasker := createTasker(t, nil)
	tasker.Destroy()
}

func TestTasker_BindResource(t *testing.T) {
	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	bound := tasker.BindResource(res)
	require.True(t, bound)
}

func TestTasker_BindController(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	bound := tasker.BindController(ctrl)
	require.True(t, bound)
}

func TestTasker_Initialized(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	connected := ctrl.PostConnect().Wait().Success()
	require.True(t, connected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	initialized := tasker.Initialized()
	require.True(t, initialized)
}

func TestTasker_PostPipeline(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	taskJob := tasker.PostTask("TestTasker_PostPipeline", J{
		"TestTasker_PostPipeline": J{
			"action": "Click",
			"target": []int{100, 200, 100, 100},
		},
	})
	got := taskJob.Wait().Success()
	require.True(t, got)
	detail := taskJob.GetDetail()
	t.Logf("%#v", detail)
}

func TestTasker_Running(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	got := tasker.Running()
	require.False(t, got)
}

func TestTasker_PostStop(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	got := tasker.PostStop()
	require.NotNil(t, got)
}

func TestTasker_GetResource(t *testing.T) {
	res1 := createResource(t, nil)
	defer res1.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	bound := tasker.BindResource(res1)
	require.True(t, bound)

	res2 := tasker.GetResource()
	require.NotNil(t, res2)
	require.Equal(t, res1.handle, res2.handle)
}

func TestTasker_GetController(t *testing.T) {
	ctrl1 := createDbgController(t, nil)
	defer ctrl1.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	bound := tasker.BindController(ctrl1)
	require.True(t, bound)

	ctrl2 := tasker.GetController()
	require.NotNil(t, ctrl2)
	require.Equal(t, ctrl1.Handle(), ctrl2.Handle())
}

func TestTasker_ClearCache(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()

	tasker := createTasker(t, nil)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	got := tasker.ClearCache()
	require.True(t, got)
}

func TestTasker_GetLatestNode(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, nil)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)

	tasker := createTasker(t, nil)
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
