package maa

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEvent_String(t *testing.T) {
	testCases := []struct {
		name   string
		event  Event
		expect string
	}{
		{
			name:   "ResourceLoading",
			event:  EventResourceLoading,
			expect: "Resource.Loading",
		},
		{
			name:   "ControllerAction",
			event:  EventControllerAction,
			expect: "Controller.Action",
		},
		{
			name:   "TaskerTask",
			event:  EventTaskerTask,
			expect: "Tasker.Task",
		},
		{
			name:   "NodePipelineNode",
			event:  EventNodePipelineNode,
			expect: "Node.PipelineNode",
		},
		{
			name:   "NodeRecognitionNode",
			event:  EventNodeRecognitionNode,
			expect: "Node.RecognitionNode",
		},
		{
			name:   "NodeActionNode",
			event:  EventNodeActionNode,
			expect: "Node.ActionNode",
		},
		{
			name:   "NodeNextList",
			event:  EventNodeNextList,
			expect: "Node.NextList",
		},
		{
			name:   "NodeRecognition",
			event:  EventNodeRecognition,
			expect: "Node.Recognition",
		},
		{
			name:   "NodeAction",
			event:  EventNodeAction,
			expect: "Node.Action",
		},
		{
			name:   "NodeWaitFreezes",
			event:  EventNodeWaitFreezes,
			expect: "Node.WaitFreezes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.event.String()
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestEvent_Starting(t *testing.T) {
	testCases := []struct {
		name   string
		event  Event
		expect string
	}{
		{
			name:   "ResourceLoading",
			event:  EventResourceLoading,
			expect: "Resource.Loading.Starting",
		},
		{
			name:   "NodeAction",
			event:  EventNodeAction,
			expect: "Node.Action.Starting",
		},
		{
			name:   "NodeWaitFreezes",
			event:  EventNodeWaitFreezes,
			expect: "Node.WaitFreezes.Starting",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.event.Starting()
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestEvent_Succeeded(t *testing.T) {
	testCases := []struct {
		name   string
		event  Event
		expect string
	}{
		{
			name:   "ResourceLoading",
			event:  EventResourceLoading,
			expect: "Resource.Loading.Succeeded",
		},
		{
			name:   "NodeAction",
			event:  EventNodeAction,
			expect: "Node.Action.Succeeded",
		},
		{
			name:   "NodeWaitFreezes",
			event:  EventNodeWaitFreezes,
			expect: "Node.WaitFreezes.Succeeded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.event.Succeeded()
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestEvent_Failed(t *testing.T) {
	testCases := []struct {
		name   string
		event  Event
		expect string
	}{
		{
			name:   "ResourceLoading",
			event:  EventResourceLoading,
			expect: "Resource.Loading.Failed",
		},
		{
			name:   "NodeAction",
			event:  EventNodeAction,
			expect: "Node.Action.Failed",
		},
		{
			name:   "NodeWaitFreezes",
			event:  EventNodeWaitFreezes,
			expect: "Node.WaitFreezes.Failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.event.Failed()
			require.Equal(t, tc.expect, got)
		})
	}
}
