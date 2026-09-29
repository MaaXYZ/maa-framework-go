package maa

import (
	"image"
	"testing"
	"time"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
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

func TestTasker_PostRecognition_CreateBufferFailure(t *testing.T) {
	tasker := createTasker(t)
	defer tasker.Destroy()

	oldCreate := native.MaaImageBufferCreate
	defer func() { native.MaaImageBufferCreate = oldCreate }()
	native.MaaImageBufferCreate = func() uintptr { return 0 }

	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	taskJob, err := tasker.PostRecognition(RecognitionTypeOCR, OCRParam{Expected: []string{"Hello"}}, img)
	require.Error(t, err)
	require.NotNil(t, taskJob)
	require.True(t, taskJob.Failure())
	require.True(t, taskJob.Done())
	require.ErrorIs(t, taskJob.Error(), err)
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
	ctrl := createBlankController(t)
	defer ctrl.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	err := tasker.BindController(ctrl)
	require.NoError(t, err)
}

func TestTasker_Initialized(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	connected := connectJob.Wait().Success()
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
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	screencapJob, err := ctrl.PostScreencap()
	require.NoError(t, err)
	require.True(t, screencapJob.Wait().Success())

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

	taskJob, err := tasker.PostTask(testTasker_PostPipelineNode.Name, pipeline)
	require.NoError(t, err)
	got := taskJob.Wait().Success()
	require.True(t, got)
	detail, err := taskJob.GetDetail()
	require.NoError(t, err)
	t.Logf("%#v", detail)
}

func TestTasker_GetTaskDetail_NodesAndGetNodeDetail(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	screencapJob, err := ctrl.PostScreencap()
	require.NoError(t, err)
	require.True(t, screencapJob.Wait().Success())

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

	taskJob, err := tasker.PostTask(testNode.Name, pipeline)
	require.NoError(t, err)
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
		wantErr  bool
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
			name: "marshal error is not posted",
			override: []any{map[string]any{
				"f": func() {},
			}},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			posted := false
			var gotEntry, gotOverride string
			postFunc := func(entry, override string) (*TaskJob, error) {
				posted = true
				gotEntry = entry
				gotOverride = override
				return &TaskJob{}, nil
			}

			taskJob, err := tasker.handleOverride("Entry", postFunc, tc.override...)
			require.NotNil(t, taskJob)
			if tc.wantErr {
				require.Error(t, err)
				require.False(t, posted)
				require.ErrorIs(t, taskJob.Error(), err)
				require.True(t, taskJob.Failure())
				return
			}
			require.NoError(t, err)
			require.True(t, posted)
			require.Equal(t, "Entry", gotEntry)
			require.Equal(t, tc.want, gotOverride)
		})
	}
}

func TestTasker_DetailQueryFailures(t *testing.T) {
	tasker := createTasker(t)
	defer tasker.Destroy()

	_, err := tasker.GetRecognitionDetail(999)
	require.Error(t, err)

	_, err = tasker.GetActionDetail(999)
	require.Error(t, err)

	_, err = tasker.GetWaitFreezesDetail(999)
	require.Error(t, err)
}

func TestTasker_GetNodeDetail_SkipsAbsentSubDetails(t *testing.T) {
	tasker := createTasker(t)
	defer tasker.Destroy()

	oldGetNodeDetail := native.MaaTaskerGetNodeDetail
	defer func() { native.MaaTaskerGetNodeDetail = oldGetNodeDetail }()
	native.MaaTaskerGetNodeDetail = func(_ uintptr, _ int64, _ uintptr, recId, actionId *int64, completed *bool) bool {
		*recId = 0
		*actionId = 0
		*completed = true
		return true
	}

	detail, err := tasker.GetNodeDetail(7)
	require.NoError(t, err)
	require.NotNil(t, detail)
	require.Nil(t, detail.Recognition)
	require.Nil(t, detail.Action)
	require.True(t, detail.RunCompleted)
}

func TestTasker_Running(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
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
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	stopJob, err := tasker.PostStop()
	require.NoError(t, err)
	ok := stopJob.Wait().Success()
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
	ctrl1 := createBlankController(t)
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
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	err = tasker.ClearCache()
	require.NoError(t, err)
}

func TestTasker_GetLatestNode(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	bundleJob, err := res.PostBundle(resDir)
	require.NoError(t, err)
	isPathSet := bundleJob.Wait().Success()
	require.True(t, isPathSet)

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)
	job, err := tasker.PostTask("Wilderness")
	require.NoError(t, err)
	require.NotNil(t, job)
	time.Sleep(2 * time.Second)
	detail, err := tasker.GetLatestNode("Wilderness")
	require.NoError(t, err)
	t.Log(detail)
	stopJob, err := tasker.PostStop()
	require.NoError(t, err)
	ok := stopJob.Wait().Success()
	require.True(t, ok)
}

func TestTasker_OverridePipeline(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	screencapJob, err := ctrl.PostScreencap()
	require.NoError(t, err)
	require.True(t, screencapJob.Wait().Success())

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
	taskJob, err := tasker.PostTask(testNode.Name, pipeline)
	require.NoError(t, err)

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
