package maa

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createCarouselImageController(t *testing.T) *Controller {
	testingPath := "./test/data_set/PipelineSmoking/Screenshot"
	ctrl, err := NewCarouselImageController(testingPath)
	require.NoError(t, err)
	require.NotNil(t, ctrl)
	return ctrl
}

func TestNewCarouselImageController(t *testing.T) {
	ctrl := createCarouselImageController(t)
	ctrl.Destroy()
}

func TestController_Handle(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	require.NotNil(t, ctrl)
}

func TestController_SetScreenshotTargetLongSide(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	got := ctrl.SetScreenshotTargetLongSide(1280)
	require.True(t, got)
}

func TestController_SetScreenshotTargetShortSide(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	got := ctrl.SetScreenshotTargetShortSide(720)
	require.True(t, got)
}

func TestController_SetScreenshotUseRawSize(t *testing.T) {
	testCases := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:    "enabled true",
			enabled: true,
		},
		{
			name:    "enabled false",
			enabled: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := createCarouselImageController(t)
			defer ctrl.Destroy()
			got := ctrl.SetScreenshotUseRawSize(tc.enabled)
			require.True(t, got)
		})
	}
}

func TestController_PostConnect(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
}

func TestController_Connected(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	connected := ctrl.Connected()
	require.True(t, connected)
}

func TestController_PostClick(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	clicked := ctrl.PostClick(100, 200).Wait().Success()
	require.True(t, clicked)
}

func TestController_PostSwipe(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	swiped := ctrl.PostSwipe(100, 200, 400, 300, 2*time.Second).Wait().Success()
	require.True(t, swiped)
}

func TestController_PostClickKey(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	pressed := ctrl.PostClickKey(4).Wait().Success()
	require.True(t, pressed)
}

func TestController_PostInputText(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	inputted := ctrl.PostInputText("Hello World").Wait().Success()
	require.True(t, inputted)
}

func TestController_PostStartApp(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	started := ctrl.PostStartApp("com.android.settings").Wait().Success()
	require.True(t, started)
}

func TestController_PostStopApp(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	stopped := ctrl.PostStopApp("com.android.settings").Wait().Success()
	require.True(t, stopped)
}

func TestController_PostTouchDown(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	downed := ctrl.PostTouchDown(0, 100, 200, 1000).Wait().Success()
	require.True(t, downed)
}

func TestController_PostTouchMove(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	downed := ctrl.PostTouchDown(0, 100, 200, 1000).Wait().Success()
	require.True(t, downed)
	moved := ctrl.PostTouchMove(0, 200, 300, 1000).Wait().Success()
	require.True(t, moved)
}

func TestController_PostTouchUp(t *testing.T) {
	ctrl := createCarouselImageController(t)
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

func TestController_PostKeyDown(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	downed := ctrl.PostKeyDown(4).Wait().Success()
	require.True(t, downed)
}

func TestController_PostKeyUp(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	downed := ctrl.PostKeyDown(4).Wait().Success()
	require.True(t, downed)
	upped := ctrl.PostKeyUp(4).Wait().Success()
	require.True(t, upped)
}

func TestController_PostScreencap(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	screencaped := ctrl.PostScreencap().Wait().Success()
	require.True(t, screencaped)
}

func TestController_CacheImage(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	screencaped := ctrl.PostScreencap().Wait().Success()
	require.True(t, screencaped)
	img := ctrl.CacheImage()
	require.NotNil(t, img)
}

func TestController_GetUUID(t *testing.T) {
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	uuid, oK := ctrl.GetUUID()
	require.True(t, oK)
	require.NotEmpty(t, uuid)
}
