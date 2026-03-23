package maa

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createTasker(t *testing.T) *Tasker {
	tasker, err := NewTasker()
	require.NoError(t, err)
	require.NotNil(t, tasker)
	return tasker
}

func taskerBind(t *testing.T, tasker *Tasker, ctrl *Controller, res *Resource) {
	err := tasker.BindResource(res)
	require.NoError(t, err)
	err = tasker.BindController(ctrl)
	require.NoError(t, err)
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
	err := tasker.BindResource(res)
	require.NoError(t, err)
}

func TestTasker_BindController(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	err := tasker.BindController(ctrl)
	require.NoError(t, err)
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
	testTasker_PostPipelineNode := NewNode("TestTasker_PostPipeline").
		SetAction(ActClick(ClickParam{
			Target: NewTargetRect(Rect{100, 200, 100, 100}),
		}))
	pipeline.AddNode(testTasker_PostPipelineNode)

	taskJob := tasker.PostTask(testTasker_PostPipelineNode.Name, pipeline)
	got := taskJob.Wait().Success()
	require.True(t, got)
	detail, err := taskJob.GetDetail()
	require.NoError(t, err)
	t.Logf("%#v", detail)
}

func TestTasker_GetTaskDetail_NodesAndGetNodeDetail(t *testing.T) {
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
	testNode := NewNode("TestTasker_GetTaskDetail_NodeIDs").
		SetAction(ActClick(ClickParam{
			Target: NewTargetRect(Rect{100, 200, 100, 100}),
		}))
	pipeline.AddNode(testNode)

	taskJob := tasker.PostTask(testNode.Name, pipeline)
	require.True(t, taskJob.Wait().Success())

	detail, err := taskJob.GetDetail()
	require.NoError(t, err)
	require.NotNil(t, detail)
	require.Len(t, detail.Nodes, 1)
	require.NotZero(t, detail.Nodes[0].ID())

	nodeDetail, err := detail.Nodes[0].GetDetail()
	require.NoError(t, err)
	require.NotNil(t, nodeDetail)
	require.Equal(t, detail.Nodes[0].ID(), nodeDetail.ID)
	require.Equal(t, testNode.Name, nodeDetail.Name)
}

func TestTasker_handleOverride(t *testing.T) {
	tasker := &Tasker{}

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
			called := false
			var gotEntry, gotOverride string
			postFunc := func(entry, override string) *TaskJob {
				called = true
				gotEntry = entry
				gotOverride = override
				return &TaskJob{}
			}

			taskJob := tasker.handleOverride("Entry", postFunc, tc.override...)
			require.NotNil(t, taskJob)
			require.True(t, called)
			require.Equal(t, "Entry", gotEntry)
			require.Equal(t, tc.want, gotOverride)
		})
	}
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

	ok := tasker.PostStop().Wait().Success()
	require.True(t, ok)
}

func TestTasker_GetResource(t *testing.T) {
	res1 := createResource(t)
	defer res1.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	err := tasker.BindResource(res1)
	require.NoError(t, err)

	res2 := tasker.GetResource()
	require.NotNil(t, res2)
	require.Equal(t, res1.handle, res2.handle)
}

func TestTasker_GetController(t *testing.T) {
	ctrl1 := createCarouselImageController(t)
	defer ctrl1.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	err := tasker.BindController(ctrl1)
	require.NoError(t, err)

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

	err := tasker.ClearCache()
	require.NoError(t, err)
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
	detail, err := tasker.GetLatestNode("Wilderness")
	require.NoError(t, err)
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
	testNode := NewNode("TestTasker_OverridePipeline").
		SetAction(ActClick(ClickParam{
			Target: NewTargetRect(Rect{100, 200, 100, 100}),
		}))
	pipeline.AddNode(testNode)

	// Start a task
	taskJob := tasker.PostTask(testNode.Name, pipeline)

	// Override the pipeline while task is running
	overridePipeline := NewPipeline()
	overrideNode := NewNode("TestTasker_OverridePipeline").
		SetAction(ActClick(ClickParam{
			Target: NewTargetRect(Rect{200, 300, 100, 100}),
		}))
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
