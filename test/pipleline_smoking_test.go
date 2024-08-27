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
	defer ctrl.Destroy()
	ctrl.PostConnect().Wait()

	res := maa.NewResource(nil)
	defer res.Destroy()
	resDir := "./data_set/PipelineSmoking/resource"
	res.PostPath(resDir).Wait()

	inst := maa.New(nil)
	defer inst.Destroy()
	inst.BindResource(res)
	inst.BindController(ctrl)

	require.True(t, inst.Inited())

	got := inst.PostTask("Wilderness", "{}").Wait()
	require.True(t, got)
}
