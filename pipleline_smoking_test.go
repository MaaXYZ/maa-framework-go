package maa

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPipelineSmoking(t *testing.T) {
	testingPath := "./test/data_set/PipelineSmoking/MaaRecording.txt"
	resultPath := "./test/data_set/debug"

	ctrl := NewDbgController(testingPath, resultPath, DbgControllerTypeReplayRecording, "{}", nil)
	defer ctrl.Destroy()
	ctrlJob := ctrl.PostConnect()

	res := NewResource(nil)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
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
