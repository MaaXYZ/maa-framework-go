package test

import (
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/stretchr/testify/require"
)

func TestPipelineSmoking(t *testing.T) {
	testingPath := "./data_set/PipelineSmoking/MaaRecording.jsonl"

	ctrl, err := maa.NewReplayController(testingPath)
	require.NoError(t, err)
	require.NotNil(t, ctrl)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res, err := maa.NewResource()
	require.NoError(t, err)
	require.NotNil(t, res)
	defer res.Destroy()
	resDir := "./data_set/PipelineSmoking/resource"
	isPathSet := res.PostBundle(resDir).Wait().Success()
	require.True(t, isPathSet)

	tasker, err := maa.NewTasker()
	require.NoError(t, err)
	require.NotNil(t, tasker)
	defer tasker.Destroy()
	err = tasker.BindResource(res)
	require.NoError(t, err)
	err = tasker.BindController(ctrl)
	require.NoError(t, err)

	isInitialized := tasker.Initialized()
	require.True(t, isInitialized)

	got := tasker.PostTask("Wilderness").Wait().Success()
	require.True(t, got)
}
