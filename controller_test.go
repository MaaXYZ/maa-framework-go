package maa

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createDbgController(t *testing.T, notify Notification) Controller {
	testingPath := "./test/data_set/PipelineSmoking/Screenshot"
	resultPath := "./test/data_set/debug"

	ctrl := NewDbgController(testingPath, resultPath, DbgControllerTypeCarouselImage, "{}", notify)
	require.NotNil(t, ctrl)
	return ctrl
}

func TestNewDbgController(t *testing.T) {
	ctrl := createDbgController(t, nil)
	ctrl.Destroy()
}

func TestController_Handle(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	handle := ctrl.Handle()
	require.NotNil(t, handle)
}

func TestController_SetScreenshotTargetLongSide(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	got := ctrl.SetScreenshotTargetLongSide(1280)
	require.True(t, got)
}

func TestController_SetScreenshotTargetShortSide(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	got := ctrl.SetScreenshotTargetShortSide(720)
	require.True(t, got)
}

func TestController_SetRecording(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	got := ctrl.SetRecording(true)
	require.True(t, got)
}

func TestController_PostConnect(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
}

func TestController_Connected(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	connected := ctrl.Connected()
	require.True(t, connected)
}

func TestController_PostClick(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	clicked := ctrl.PostClick(100, 200).Wait().Success()
	require.True(t, clicked)
}

func TestController_PostSwipe(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	swiped := ctrl.PostSwipe(100, 200, 400, 300, 2*time.Second).Wait().Success()
	require.True(t, swiped)
}

func TestController_PostPressKey(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	pressed := ctrl.PostPressKey(4).Wait().Success()
	require.True(t, pressed)
}

func TestController_PostInputText(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	inputted := ctrl.PostInputText("Hello World").Wait().Success()
	require.True(t, inputted)
}

func TestController_PostStartApp(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	started := ctrl.PostStartApp("com.android.settings").Wait().Success()
	require.True(t, started)
}

func TestController_PostStopApp(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	stopped := ctrl.PostStopApp("com.android.settings").Wait().Success()
	require.True(t, stopped)
}

func TestController_PostTouchDown(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	downed := ctrl.PostTouchDown(0, 100, 200, 1000).Wait().Success()
	require.True(t, downed)
}

func TestController_PostTouchMove(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	downed := ctrl.PostTouchDown(0, 100, 200, 1000).Wait().Success()
	require.True(t, downed)
	moved := ctrl.PostTouchMove(0, 200, 300, 1000).Wait().Success()
	require.True(t, moved)
}

func TestController_PostTouchUp(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	downed := ctrl.PostTouchDown(0, 100, 200, 1000).Wait().Success()
	require.True(t, downed)
	moved := ctrl.PostTouchMove(0, 200, 300, 1000).Wait().Success()
	require.True(t, moved)
	upped := ctrl.PostTouchUp(0).Wait().Success()
	require.True(t, upped)
}

func TestController_PostScreencap(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	screencaped := ctrl.PostScreencap().Wait().Success()
	require.True(t, screencaped)
}

func TestController_CacheImage(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	screencaped := ctrl.PostScreencap().Wait().Success()
	require.True(t, screencaped)
	img := ctrl.CacheImage()
	require.NotNil(t, img)
}

func TestController_GetUUID(t *testing.T) {
	ctrl := createDbgController(t, nil)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	uuid, oK := ctrl.GetUUID()
	require.True(t, oK)
	require.NotEmpty(t, uuid)
}
