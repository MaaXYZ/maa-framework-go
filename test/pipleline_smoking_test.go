package test

// func TestPipelineSmoking(t *testing.T) {
// 	testingPath := "./data_set/PipelineSmoking/MaaRecording.txt"
// 	resultPath := "./data_set/debug"

// 	ctrl := maa.NewDbgController(testingPath, resultPath, maa.DbgControllerTypeReplayRecording, "{}")
// 	require.NotNil(t, ctrl)
// 	defer ctrl.Destroy()
// 	isConnected := ctrl.PostConnect().Wait().Success()
// 	require.True(t, isConnected)

// 	res := maa.NewResource()
// 	require.NotNil(t, res)
// 	defer res.Destroy()
// 	resDir := "./data_set/PipelineSmoking/resource"
// 	isPathSet := res.PostBundle(resDir).Wait().Success()
// 	require.True(t, isPathSet)

// 	tasker := maa.NewTasker()
// 	require.NotNil(t, tasker)
// 	defer tasker.Destroy()
// 	isResBound := tasker.BindResource(res)
// 	require.True(t, isResBound)
// 	isCtrlBound := tasker.BindController(ctrl)
// 	require.True(t, isCtrlBound)

// 	isInitialized := tasker.Initialized()
// 	require.True(t, isInitialized)

// 	got := tasker.PostTask("Wilderness").Wait().Success()
// 	require.True(t, got)
// }
