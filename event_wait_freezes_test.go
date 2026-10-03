package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvent_NodeWaitFreezesDispatch(t *testing.T) {
	for _, status := range []EventStatus{EventStatusStarting, EventStatusSucceeded, EventStatusFailed} {
		name := map[EventStatus]string{EventStatusStarting: "Starting", EventStatusSucceeded: "Succeeded", EventStatusFailed: "Failed"}[status]
		t.Run(name, func(t *testing.T) {
			var captured *Context
			calls := 0
			sink := &contextEventSinkAdapter{onNodeWaitFreezes: func(ctx *Context, got EventStatus, detail NodeWaitFreezesDetail) {
				calls++
				captured = ctx
				require.Equal(t, status, got)
				require.Equal(t, uint64(17), detail.TaskID)
				require.Equal(t, int64(23), detail.WfID)
				require.Equal(t, "Click", detail.Name)
				require.Equal(t, "pre", detail.Phase)
				require.Equal(t, Rect{1, 2, 3, 4}, detail.ROI)
				require.Equal(t, float64(120), detail.Param["time"])
				require.Equal(t, []int64{31, 32}, detail.RecoIDs)
				require.Equal(t, uint64(150), detail.Elapsed)
				done, err := ctx.state.begin()
				require.NoError(t, err)
				done()
			}}
			callback := &eventCallback{sink: sink}
			callback.handleRaw(0, "Node.WaitFreezes."+name, []byte(`{"task_id":17,"wf_id":23,"name":"Click","phase":"pre","roi":[1,2,3,4],"param":{"time":120,"threshold":0.95},"reco_ids":[31,32],"elapsed":150,"focus":{"test":true}}`))
			require.Equal(t, 1, calls)
			require.NotNil(t, captured)
			_, err := captured.state.begin()
			require.ErrorIs(t, err, ErrClosed)
		})
	}
}

func TestEvent_NodeWaitFreezesInvalidPayloadAndOptionalSink(t *testing.T) {
	sink := &contextEventSinkAdapter{onNodeWaitFreezes: func(*Context, EventStatus, NodeWaitFreezesDetail) {
		t.Fatal("invalid payload reached the listener")
	}}
	callback := &eventCallback{sink: sink}
	callback.handleRaw(0, EventNodeWaitFreezes.Starting(), []byte(`{"roi":"invalid"}`))
	callback.handleRaw(0, EventNodeWaitFreezes.Starting(), []byte(`{`))
	// Listeners that do not implement the optional extension remain valid.
	callback.sink = struct{}{}
	callback.handleRaw(0, EventNodeWaitFreezes.Starting(), []byte(`{}`))
}

func TestTasker_OnNodeWaitFreezesInContext(t *testing.T) {
	tasker := createTasker(t)
	id := tasker.OnNodeWaitFreezesInContext(func(*Context, EventStatus, NodeWaitFreezesDetail) {})
	require.NotZero(t, id)
	tasker.RemoveContextSink(id)
	require.NoError(t, tasker.Destroy())
	require.Zero(t, tasker.OnNodeWaitFreezesInContext(nil))
}
