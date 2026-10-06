package maa

import (
	"encoding/json"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

// testCustomActionArgRecord snapshots the fields of a CustomActionArg inside
// the callback, so assertions run on the test goroutine after the task ends.
type testCustomActionArgRecord struct {
	taskID            int64
	currentTaskName   string
	customActionName  string
	customActionParam string
	recognitionDetail *RecognitionDetail
	box               Rect
}

type testCustomActionRecorder struct {
	called atomic.Bool
	record atomic.Pointer[testCustomActionArgRecord]
}

func (r *testCustomActionRecorder) Run(_ *Context, arg *CustomActionArg) bool {
	r.called.Store(true)
	r.record.Store(&testCustomActionArgRecord{
		taskID:            arg.TaskID,
		currentTaskName:   arg.CurrentTaskName,
		customActionName:  arg.CustomActionName,
		customActionParam: arg.CustomActionParam,
		recognitionDetail: arg.RecognitionDetail,
		box:               arg.Box,
	})
	return true
}

// testCustomActionRunActionOuter invokes the inner custom action through the
// action-only entries Context.RunAction or Context.RunActionDirect.
type testCustomActionRunActionOuter struct {
	direct bool
}

const (
	testCustomActionInnerName = "TestCustomAction_Inner"
	testCustomActionOuterName = "TestCustomAction_RunAction_Outer"
)

var testCustomActionInnerBox = Rect{11, 22, 33, 44}

func (o *testCustomActionRunActionOuter) Run(ctx *Context, arg *CustomActionArg) bool {
	// The outer node runs from a pipeline node with recognition, so its
	// RecognitionDetail is set.
	if arg.RecognitionDetail == nil {
		return false
	}
	if o.direct {
		detail, err := ctx.RunActionDirect(
			ActionTypeCustom,
			CustomActionParam{CustomAction: testCustomActionInnerName},
			testCustomActionInnerBox,
			nil,
		)
		return err == nil && detail != nil && detail.Success
	}
	pipeline := NewPipeline()
	node := NewNode("TestCustomAction_RunAction_Inner").
		SetAction(ActCustom(CustomActionParam{CustomAction: testCustomActionInnerName}))
	pipeline.AddNode(node)
	detail, err := ctx.RunAction(node.Name, testCustomActionInnerBox, "", pipeline)
	return err == nil && detail != nil && detail.Success
}

// TestCustomAction_ActionOnlyRecognitionDetailNil pins the wrapper's handling
// of the invalid reco id reported by action-only runs: a custom action invoked
// through Context.RunAction or RunActionDirect must receive a nil
// RecognitionDetail and still succeed, instead of failing on
// GetRecognitionDetail with the invalid id.
func TestCustomAction_ActionOnlyRecognitionDetailNil(t *testing.T) {
	for _, tc := range []struct {
		name   string
		direct bool
	}{
		{name: "RunAction"},
		{name: "RunActionDirect", direct: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := createBlankController(t)
			defer ctrl.Destroy()
			connectJob, err := ctrl.PostConnect()
			require.NoError(t, err)
			require.True(t, connectJob.Wait().Success())

			res := createResource(t)
			defer res.Destroy()

			recorder := &testCustomActionRecorder{}
			require.NoError(t, res.RegisterCustomAction(testCustomActionInnerName, recorder))
			require.NoError(t, res.RegisterCustomAction(
				testCustomActionOuterName,
				&testCustomActionRunActionOuter{direct: tc.direct},
			))

			tasker := createTasker(t)
			defer tasker.Destroy()
			taskerBind(t, tasker, ctrl, res)

			pipeline := NewPipeline()
			node := NewNode(testCustomActionOuterName).
				SetAction(ActCustom(CustomActionParam{CustomAction: testCustomActionOuterName}))
			pipeline.AddNode(node)

			taskJob, err := tasker.PostTask(node.Name, pipeline)
			require.NoError(t, err)
			require.True(t, taskJob.Wait().Success())
			require.True(t, recorder.called.Load(), "inner custom action was not invoked")

			record := recorder.record.Load()
			require.NotNil(t, record)
			require.Nil(t, record.recognitionDetail,
				"RecognitionDetail must be nil for an action-only custom action")
			require.Equal(t, testCustomActionInnerBox, record.box,
				"Box must be the box passed to the action-only entry")
			if !tc.direct {
				require.Equal(t, "TestCustomAction_RunAction_Inner", record.currentTaskName)
			}
		})
	}
}

// TestCustomAction_ArgFields pins that every CustomActionArg field reaches the
// runner as documented: identity fields, the task id usable with
// Tasker.GetTaskDetail, the custom param as JSON, and Box as the action's
// resolved target rect, which for the default Self target equals the
// recognition hit box.
func TestCustomAction_ArgFields(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	require.True(t, connectJob.Wait().Success())

	res := createResource(t)
	defer res.Destroy()

	const (
		actName  = "TestCustomAction_ArgFields_Act"
		nodeName = "TestCustomAction_ArgFields"
	)
	recorder := &testCustomActionRecorder{}
	require.NoError(t, res.RegisterCustomAction(actName, recorder))

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	roi := Rect{100, 120, 60, 40}
	recognition := RecDirectHit()
	recognition.Param = &DirectHitParam{ROI: NewTargetRect(roi)}

	pipeline := NewPipeline()
	node := NewNode(nodeName).
		SetRecognition(recognition).
		SetAction(ActCustom(CustomActionParam{
			CustomAction:      actName,
			CustomActionParam: map[string]any{"key": "value"},
		}))
	pipeline.AddNode(node)

	taskJob, err := tasker.PostTask(node.Name, pipeline)
	require.NoError(t, err)
	require.True(t, taskJob.Wait().Success())
	require.True(t, recorder.called.Load(), "custom action was not invoked")

	record := recorder.record.Load()
	require.NotNil(t, record)
	require.Equal(t, actName, record.customActionName)
	require.Equal(t, nodeName, record.currentTaskName)
	require.NotZero(t, record.taskID)

	var gotParam, wantParam any
	require.NoError(t, json.Unmarshal([]byte(record.customActionParam), &gotParam))
	require.NoError(t, json.Unmarshal([]byte(`{"key":"value"}`), &wantParam))
	require.Equal(t, wantParam, gotParam)

	taskDetail, err := tasker.GetTaskDetail(record.taskID)
	require.NoError(t, err)
	require.Equal(t, record.taskID, taskDetail.ID)
	require.Equal(t, nodeName, taskDetail.Entry)

	require.NotNil(t, record.recognitionDetail)
	require.True(t, record.recognitionDetail.Hit)
	require.Equal(t, roi, record.recognitionDetail.Box)
	require.Equal(t, roi, record.box,
		"Box must equal the recognition hit box for the default Self target")
}
