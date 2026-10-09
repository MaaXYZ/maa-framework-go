package maa

import (
	"encoding/json"
	"image"
	"image/color"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBlankController_UUID(t *testing.T) {
	uuid, ok := (&BlankController{}).RequestUUID()
	require.True(t, ok)
	require.Equal(t, "blank-controller", uuid)

	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	got, err := ctrl.GetUUID()
	require.NoError(t, err)
	require.Equal(t, "blank-controller", got)
}

func TestBlankController_ScreencapImage(t *testing.T) {
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
	require.Equal(t, image.Rect(0, 0, 1280, 720), img.Bounds())
	rgba, ok := img.(*image.RGBA)
	require.True(t, ok)
	require.Equal(t, color.RGBA{R: 0, G: 0, B: 0, A: 255}, rgba.RGBAAt(0, 0))
	require.Equal(t, color.RGBA{R: 0, G: 0, B: 0, A: 255}, rgba.RGBAAt(1279, 719))
}

func TestBlankController_GetInfo(t *testing.T) {
	info, ok := (&BlankController{}).GetInfo()
	require.True(t, ok)
	require.JSONEq(t, `{"type":"blank"}`, info)

	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)
	nativeInfo, err := ctrl.GetInfo()
	require.NoError(t, err)
	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(nativeInfo), &parsed))
	require.Equal(t, "custom", parsed["type"],
		"the framework overwrites the type key for every custom controller")
}

func TestBlankController_PostScrollRelativeMoveShell(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)

	scrollJob, err := ctrl.PostScroll(10, 20)
	require.NoError(t, err)
	scrolled := scrollJob.Wait().Success()
	require.True(t, scrolled)

	moveJob, err := ctrl.PostRelativeMove(-5, 5)
	require.NoError(t, err)
	moved := moveJob.Wait().Success()
	require.True(t, moved)

	shellJob, err := ctrl.PostShell("true", 2*time.Second)
	require.NoError(t, err)
	shellDone := shellJob.Wait().Success()
	require.True(t, shellDone)
}
