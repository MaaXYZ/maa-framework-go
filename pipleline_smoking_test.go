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
	ctrlJob := ctrl.PostConnect()

	res := NewResource(nil)
	defer res.Destroy()
	resDir := "./TestingDataSet/PipelineSmoking/resource"
	resJob := res.PostPath(resDir)

	ctrlJob.Wait()
	resJob.Wait()

	inst := New(nil)
	defer inst.Destroy()
	inst.BindResource(res)
	inst.BindController(ctrl)

	require.True(t, inst.Inited())

	taskJob := inst.PostTask("Wilderness", "{}")
	got := taskJob.Wait()
	require.True(t, got)
}
