package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseEvent(t *testing.T) {
	testCases := []struct {
		name   string
		msg    string
		expect string
		status EventStatus
	}{
		{
			name:   "SucceededSuffix",
			msg:    "Node.PipelineNode.Succeeded",
			expect: "Node.PipelineNode",
			status: EventStatusSucceeded,
		},
		{
			name:   "FailedSuffix",
			msg:    "Node.NextList.Failed",
			expect: "Node.NextList",
			status: EventStatusFailed,
		},
		{
			name:   "StartingSuffix",
			msg:    "Node.WaitFreezes.Starting",
			expect: "Node.WaitFreezes",
			status: EventStatusStarting,
		},
		{
			name:   "FamilyNameWithoutStatusSuffix",
			msg:    "Resource.Loading",
			expect: "Resource.Loading",
			status: EventStatusUnknown,
		},
		{
			name:   "UnknownStatusSuffix",
			msg:    "Tasker.Task.Completed",
			expect: "Tasker.Task.Completed",
			status: EventStatusUnknown,
		},
		{
			name:   "NoDot",
			msg:    "Starting",
			expect: "Starting",
			status: EventStatusUnknown,
		},
		{
			name:   "EmptyMessage",
			msg:    "",
			expect: "",
			status: EventStatusUnknown,
		},
		{
			name:   "BareStatusSuffix",
			msg:    ".Failed",
			expect: "",
			status: EventStatusFailed,
		},
		{
			name:   "NameEndingInStatusWord",
			msg:    "Custom.Starting",
			expect: "Custom",
			status: EventStatusStarting,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			name, status := parseEvent(tc.msg)
			require.Equal(t, tc.expect, name)
			require.Equal(t, tc.status, status)
		})
	}
}

// Upstream attaches reco_ids and elapsed only to the WaitFreezes completion
// events; the Starting payload carries neither key.
func TestEvent_NodeWaitFreezesStartingPayloadShape(t *testing.T) {
	var got NodeWaitFreezesDetail
	calls := 0
	sink := &contextEventSinkAdapter{onNodeWaitFreezes: func(_ *Context, status EventStatus, detail NodeWaitFreezesDetail) {
		calls++
		require.Equal(t, EventStatusStarting, status)
		got = detail
	}}
	callback := &eventCallback{sink: sink}
	callback.handleRaw(0, EventNodeWaitFreezes.Starting(), []byte(
		`{"task_id":17,"wf_id":23,"name":"Click","phase":"pre","roi":[1,2,3,4],"param":{"time":120,"threshold":0.95},"focus":null}`))

	require.Equal(t, 1, calls)
	require.Equal(t, uint64(17), got.TaskID)
	require.Nil(t, got.RecoIDs)
	require.Zero(t, got.Elapsed)
}

// The external-handle fallback used for Tasker family events requires a
// non-zero handle; zero handles are rejected as closed before the sink runs.
func TestEvent_TaskerTaskDetailDecode(t *testing.T) {
	var got TaskerTaskDetail
	calls := 0
	sink := &taskerEventSinkAdapter{onTaskerTask: func(status EventStatus, detail TaskerTaskDetail) {
		calls++
		require.Equal(t, EventStatusSucceeded, status)
		got = detail
	}}
	callback := &eventCallback{sink: sink}
	callback.handleRaw(98765, EventTaskerTask.Succeeded(), []byte(
		`{"task_id":101,"entry":"Entry","uuid":"6b3d2f01-9c44-4e2f-8e97-1b2b3c4d5e6f","hash":"0123abcd"}`))

	require.Equal(t, 1, calls)
	require.Equal(t, uint64(101), got.TaskID)
	require.Equal(t, "Entry", got.Entry)
	require.Equal(t, "6b3d2f01-9c44-4e2f-8e97-1b2b3c4d5e6f", got.UUID)
	require.Equal(t, "0123abcd", got.Hash)
}

func TestEvent_NodeNextListDetailDecode(t *testing.T) {
	var got NodeNextListDetail
	calls := 0
	sink := &contextEventSinkAdapter{onNodeNextList: func(_ *Context, status EventStatus, detail NodeNextListDetail) {
		calls++
		require.Equal(t, EventStatusStarting, status)
		got = detail
	}}
	callback := &eventCallback{sink: sink}
	callback.handleRaw(0, EventNodeNextList.Starting(), []byte(
		`{"task_id":5,"name":"A","list":[{"name":"B"},{"name":"C","jump_back":true,"anchor":true}],"focus":null}`))

	require.Equal(t, 1, calls)
	require.Equal(t, uint64(5), got.TaskID)
	require.Equal(t, "A", got.Name)
	require.Equal(t, []NextItem{
		{Name: "B"},
		{Name: "C", JumpBack: true, Anchor: true},
	}, got.List)
}

// Pins the per-family identifier fields: node_id for the Node.*Node details,
// reco_id for Node.Recognition, and action_id for Node.Action.
func TestEvent_NodeDetailIDDecode(t *testing.T) {
	t.Run("PipelineNodeNodeID", func(t *testing.T) {
		var got NodePipelineNodeDetail
		sink := &contextEventSinkAdapter{onNodePipelineNode: func(_ *Context, status EventStatus, detail NodePipelineNodeDetail) {
			require.Equal(t, EventStatusStarting, status)
			got = detail
		}}
		(&eventCallback{sink: sink}).handleRaw(0, EventNodePipelineNode.Starting(), []byte(
			`{"task_id":7,"node_id":100000001,"name":"A","focus":null}`))
		require.Equal(t, uint64(7), got.TaskID)
		require.Equal(t, uint64(100000001), got.NodeID)
		require.Equal(t, "A", got.Name)
	})
	t.Run("RecognitionNodeNodeID", func(t *testing.T) {
		var got NodeRecognitionNodeDetail
		sink := &contextEventSinkAdapter{onNodeRecognitionNode: func(_ *Context, status EventStatus, detail NodeRecognitionNodeDetail) {
			require.Equal(t, EventStatusFailed, status)
			got = detail
		}}
		(&eventCallback{sink: sink}).handleRaw(0, EventNodeRecognitionNode.Failed(), []byte(
			`{"task_id":7,"node_id":100000002,"name":"R","focus":null}`))
		require.Equal(t, uint64(100000002), got.NodeID)
		require.Equal(t, "R", got.Name)
	})
	t.Run("ActionNodeNodeID", func(t *testing.T) {
		var got NodeActionNodeDetail
		sink := &contextEventSinkAdapter{onNodeActionNode: func(_ *Context, status EventStatus, detail NodeActionNodeDetail) {
			require.Equal(t, EventStatusSucceeded, status)
			got = detail
		}}
		(&eventCallback{sink: sink}).handleRaw(0, EventNodeActionNode.Succeeded(), []byte(
			`{"task_id":7,"node_id":100000003,"name":"Act","focus":null}`))
		require.Equal(t, uint64(100000003), got.NodeID)
		require.Equal(t, "Act", got.Name)
	})
	t.Run("RecognitionRecoID", func(t *testing.T) {
		var got NodeRecognitionDetail
		sink := &contextEventSinkAdapter{onNodeRecognition: func(_ *Context, status EventStatus, detail NodeRecognitionDetail) {
			require.Equal(t, EventStatusStarting, status)
			got = detail
		}}
		(&eventCallback{sink: sink}).handleRaw(0, EventNodeRecognition.Starting(), []byte(
			`{"task_id":7,"reco_id":300000001,"name":"Rec","focus":null}`))
		require.Equal(t, uint64(300000001), got.RecognitionID)
		require.Equal(t, "Rec", got.Name)
	})
	t.Run("ActionActionID", func(t *testing.T) {
		var got NodeActionDetail
		sink := &contextEventSinkAdapter{onNodeAction: func(_ *Context, status EventStatus, detail NodeActionDetail) {
			require.Equal(t, EventStatusFailed, status)
			got = detail
		}}
		(&eventCallback{sink: sink}).handleRaw(0, EventNodeAction.Failed(), []byte(
			`{"task_id":7,"action_id":400000001,"name":"Do","focus":null}`))
		require.Equal(t, uint64(400000001), got.ActionID)
		require.Equal(t, "Do", got.Name)
	})
}

// Upstream payloads carry keys the Go detail types do not decode: Resource.Loading
// adds "type", Node.Recognition may add "anchor" and terminal "reco_details", and
// the Node.*Node terminal events add node_details/reco_details/action_details.
// Dispatch must tolerate all of them with the default decoder.
func TestEvent_DetailUnknownKeysTolerated(t *testing.T) {
	t.Run("ResourceLoadingTypeKey", func(t *testing.T) {
		var got ResourceLoadingDetail
		calls := 0
		sink := &resourceEventSinkAdapter{onResourceLoading: func(status EventStatus, detail ResourceLoadingDetail) {
			calls++
			require.Equal(t, EventStatusStarting, status)
			got = detail
		}}
		(&eventCallback{sink: sink}).handleRaw(98766, EventResourceLoading.Starting(), []byte(
			`{"res_id":9,"path":"/res","type":"Bundle","hash":"abc"}`))
		require.Equal(t, 1, calls)
		require.Equal(t, uint64(9), got.ResID)
		require.Equal(t, "/res", got.Path)
		require.Equal(t, "abc", got.Hash)
	})
	t.Run("RecognitionAnchorAndRecoDetails", func(t *testing.T) {
		var got NodeRecognitionDetail
		calls := 0
		sink := &contextEventSinkAdapter{onNodeRecognition: func(_ *Context, status EventStatus, detail NodeRecognitionDetail) {
			calls++
			require.Equal(t, EventStatusSucceeded, status)
			got = detail
		}}
		(&eventCallback{sink: sink}).handleRaw(0, EventNodeRecognition.Succeeded(), []byte(
			`{"task_id":7,"reco_id":300000002,"name":"R","focus":null,"anchor":"A0","reco_details":{"box":[1,2,3,4]}}`))
		require.Equal(t, 1, calls)
		require.Equal(t, uint64(300000002), got.RecognitionID)
		require.Equal(t, "R", got.Name)
	})
	t.Run("PipelineNodeTerminalDetailKeys", func(t *testing.T) {
		var got NodePipelineNodeDetail
		calls := 0
		sink := &contextEventSinkAdapter{onNodePipelineNode: func(_ *Context, status EventStatus, detail NodePipelineNodeDetail) {
			calls++
			require.Equal(t, EventStatusSucceeded, status)
			got = detail
		}}
		(&eventCallback{sink: sink}).handleRaw(0, EventNodePipelineNode.Succeeded(), []byte(
			`{"task_id":7,"node_id":100000004,"name":"N","focus":null,"node_details":{"focus":true},"reco_details":null,"action_details":null}`))
		require.Equal(t, 1, calls)
		require.Equal(t, uint64(100000004), got.NodeID)
		require.Equal(t, "N", got.Name)
	})
}

func TestEvent_FocusDecoding(t *testing.T) {
	testCases := []struct {
		name   string
		focus  string
		expect any
	}{
		{name: "NullFocus", focus: `"focus":null,`, expect: nil},
		{name: "ObjectFocus", focus: `"focus":{"k":1},`, expect: map[string]any{"k": float64(1)}},
		{name: "MissingFocus", focus: "", expect: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got any
			calls := 0
			sink := &contextEventSinkAdapter{onNodePipelineNode: func(_ *Context, status EventStatus, detail NodePipelineNodeDetail) {
				calls++
				require.Equal(t, EventStatusSucceeded, status)
				got = detail.Focus
			}}
			payload := `{"task_id":1,"node_id":2,` + tc.focus + `"name":"N"}`
			(&eventCallback{sink: sink}).handleRaw(0, EventNodePipelineNode.Succeeded(), []byte(payload))
			require.Equal(t, 1, calls)
			require.Equal(t, tc.expect, got)
		})
	}
}
