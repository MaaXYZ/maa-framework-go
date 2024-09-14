package test

import (
	"github.com/MaaXYZ/maa-framework-go"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPipelineSmoking(t *testing.T) {
	testingPath := "./data_set/PipelineSmoking/MaaRecording.txt"
	resultPath := "./data_set/debug"

	ctrl := maa.NewDbgController(testingPath, resultPath, maa.DbgControllerTypeReplayRecording, "{}", nil)
	require.NotNil(t, ctrl)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := maa.NewResource(nil)
	require.NotNil(t, res)
	defer res.Destroy()
	resDir := "./data_set/PipelineSmoking/resource"
	isPathSet := res.PostPath(resDir).Wait().Success()
	require.True(t, isPathSet)

	tasker := maa.NewTasker(nil)
	require.NotNil(t, tasker)
	defer tasker.Destroy()
	isResBound := tasker.BindResource(res)
	require.True(t, isResBound)
	isCtrlBound := tasker.BindController(ctrl)
	require.True(t, isCtrlBound)

	isInitialized := tasker.Inited()
	require.True(t, isInitialized)

	got := tasker.PostPipeline("Wilderness").Wait().Success()
	require.True(t, got)
}
