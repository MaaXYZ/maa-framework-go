package maa

import (
	"image"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createBlankController(t *testing.T) *Controller {
	ctrl, err := NewBlankController()
	require.NoError(t, err)
	require.NotNil(t, ctrl)
	return ctrl
}

func TestNewBlankController(t *testing.T) {
	ctrl := createBlankController(t)
	ctrl.Destroy()
}

func TestNewLinuxController_InvalidConfig(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux controller is only available on Linux")
	}

	ctrl, err := NewLinuxController(`{}`)
	require.Error(t, err)
	require.Nil(t, ctrl)
}

func TestController_Handle(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	require.NotNil(t, ctrl)
}

func TestController_SetScreenshotTargetLongSide(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	err := ctrl.SetScreenshot(WithScreenshotTargetLongSide(1280))
	require.NoError(t, err)
}

func TestController_SetScreenshotTargetShortSide(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	err := ctrl.SetScreenshot(WithScreenshotTargetShortSide(720))
	require.NoError(t, err)
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
			ctrl := createBlankController(t)
			defer ctrl.Destroy()
			err := ctrl.SetScreenshot(WithScreenshotUseRawSize(tc.enabled))
			require.NoError(t, err)
		})
	}
}

func TestController_PostConnect(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
}

func TestController_Connected(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	connected := ctrl.Connected()
	require.True(t, connected)
}

func TestController_PostClick(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	clickJob, err := ctrl.PostClick(100, 200)
	require.NoError(t, err)
	clicked := clickJob.Wait().Success()
	require.True(t, clicked)
}

func TestController_PostSwipe(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	swipeJob, err := ctrl.PostSwipe(100, 200, 400, 300, 2*time.Second)
	require.NoError(t, err)
	swiped := swipeJob.Wait().Success()
	require.True(t, swiped)
}

func TestController_PostClickKey(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	clickKeyJob, err := ctrl.PostClickKey(4)
	require.NoError(t, err)
	pressed := clickKeyJob.Wait().Success()
	require.True(t, pressed)
}

func TestController_PostInputText(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	inputTextJob, err := ctrl.PostInputText("Hello World")
	require.NoError(t, err)
	inputted := inputTextJob.Wait().Success()
	require.True(t, inputted)
}

func TestController_PostStartApp(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	startAppJob, err := ctrl.PostStartApp("com.android.settings")
	require.NoError(t, err)
	started := startAppJob.Wait().Success()
	require.True(t, started)
}

func TestController_PostStopApp(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	stopAppJob, err := ctrl.PostStopApp("com.android.settings")
	require.NoError(t, err)
	stopped := stopAppJob.Wait().Success()
	require.True(t, stopped)
}

func TestController_PostTouchDown(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	touchDownJob, err := ctrl.PostTouchDown(0, 100, 200, 1000)
	require.NoError(t, err)
	downed := touchDownJob.Wait().Success()
	require.True(t, downed)
}

func TestController_PostTouchMove(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	touchDownJob, err := ctrl.PostTouchDown(0, 100, 200, 1000)
	require.NoError(t, err)
	downed := touchDownJob.Wait().Success()
	require.True(t, downed)
	touchMoveJob, err := ctrl.PostTouchMove(0, 200, 300, 1000)
	require.NoError(t, err)
	moved := touchMoveJob.Wait().Success()
	require.True(t, moved)
}

func TestController_PostTouchUp(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	touchDownJob, err := ctrl.PostTouchDown(0, 100, 200, 1000)
	require.NoError(t, err)
	downed := touchDownJob.Wait().Success()
	require.True(t, downed)
	touchMoveJob, err := ctrl.PostTouchMove(0, 200, 300, 1000)
	require.NoError(t, err)
	moved := touchMoveJob.Wait().Success()
	require.True(t, moved)
	touchUpJob, err := ctrl.PostTouchUp(0)
	require.NoError(t, err)
	upped := touchUpJob.Wait().Success()
	require.True(t, upped)
}

func TestController_PostKeyDown(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	keyDownJob, err := ctrl.PostKeyDown(4)
	require.NoError(t, err)
	downed := keyDownJob.Wait().Success()
	require.True(t, downed)
}

func TestController_PostKeyUp(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	keyDownJob, err := ctrl.PostKeyDown(4)
	require.NoError(t, err)
	downed := keyDownJob.Wait().Success()
	require.True(t, downed)
	keyUpJob, err := ctrl.PostKeyUp(4)
	require.NoError(t, err)
	upped := keyUpJob.Wait().Success()
	require.True(t, upped)
}

func TestController_PostScreencap(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	screencapJob, err := ctrl.PostScreencap()
	require.NoError(t, err)
	screencaped := screencapJob.Wait().Success()
	require.True(t, screencaped)
}

func TestController_PostInactive(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	inactiveJob, err := ctrl.PostInactive()
	require.NoError(t, err)
	inactiveOk := inactiveJob.Wait().Success()
	require.True(t, inactiveOk)
}

func TestController_CacheImage(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	screencapJob, err := ctrl.PostScreencap()
	require.NoError(t, err)
	screencaped := screencapJob.Wait().Success()
	require.True(t, screencaped)
	img, err := ctrl.CacheImage()
	require.NoError(t, err)
	require.NotNil(t, img)
}

func TestController_CacheImageInto(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	screencapJob, err := ctrl.PostScreencap()
	require.NoError(t, err)
	screencaped := screencapJob.Wait().Success()
	require.True(t, screencaped)

	img1, err := ctrl.CacheImageInto(nil)
	require.NoError(t, err)
	require.NotNil(t, img1)

	reused := image.NewRGBA(img1.Bounds())
	img2, err := ctrl.CacheImageInto(reused)
	require.NoError(t, err)
	require.NotNil(t, img2)
	require.Same(t, reused, img2)

	mismatch := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img3, err := ctrl.CacheImageInto(mismatch)
	require.NoError(t, err)
	require.NotNil(t, img3)
	require.NotSame(t, mismatch, img3)
}

func TestController_GetUUID(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	uuid, err := ctrl.GetUUID()
	require.NoError(t, err)
	require.NotEmpty(t, uuid)
}

func TestController_GetInfo(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	info, err := ctrl.GetInfo()
	require.NoError(t, err)
	require.NotEmpty(t, info)
}
