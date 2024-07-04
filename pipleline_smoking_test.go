package maa

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPipelineSmoking(t *testing.T) {
	testingPath := "./TestingDataSet/PipelineSmoking/MaaRecording.txt"
	resultPath := "./TestingDataSet/debug"

	ctrl := NewDbgController(testingPath, resultPath, DbgControllerTypeReplayRecording, "{}", nil)
	defer ctrl.Destroy()
	ctrlId := ctrl.PostConnect()

	res := NewResource(nil)
	defer res.Destroy()
	resDir := "./TestingDataSet/PipelineSmoking/resource"
	resId := res.PostPath(resDir)

	ctrl.Wait(ctrlId)
	res.Wait(resId)

	inst := New(nil)
	defer inst.Destroy()
	inst.BindResource(res)
	inst.BindController(ctrl)

	require.True(t, inst.Inited())

	taskId := inst.PostTask("Wilderness", "{}")
	status := inst.WaitTask(taskId)
	require.Equal(t, status, Success)
}
