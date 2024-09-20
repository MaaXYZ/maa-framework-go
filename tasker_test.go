package maa

import (
	"github.com/stretchr/testify/require"
	"testing"
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
