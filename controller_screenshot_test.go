package maa

import (
	"image"
	"testing"

	"github.com/stretchr/testify/require"
)

type fixedScreenshotController struct {
	BlankController
	img image.Image
}

func (c *fixedScreenshotController) Screencap() (image.Image, bool) {
	return c.img, true
}

func createScreenshotController(t *testing.T, width, height int) *Controller {
	t.Helper()
	ctrl, err := NewCustomController(&fixedScreenshotController{
		img: image.NewRGBA(image.Rect(0, 0, width, height)),
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, ctrl.Destroy()) })
	require.True(t, ctrl.PostConnect().Wait().Success())
	return ctrl
}

func requireScreenshotSize(t *testing.T, ctrl *Controller, width, height int) {
	t.Helper()
	require.True(t, ctrl.PostScreencap().Wait().Success())
	img, err := ctrl.CacheImage()
	require.NoError(t, err)
	require.NotNil(t, img)
	require.Equal(t, image.Pt(width, height), img.Bounds().Size())
}

func TestController_ScreenshotTargetExpand_Size(t *testing.T) {
	for _, tc := range []struct {
		name              string
		raw, target, want image.Point
	}{
		{"landscape", image.Pt(400, 200), image.Pt(128, 72), image.Pt(144, 72)},
		{"portrait", image.Pt(200, 400), image.Pt(72, 128), image.Pt(72, 144)},
		{"same aspect ratio", image.Pt(400, 200), image.Pt(120, 60), image.Pt(120, 60)},
		{"upscale", image.Pt(100, 50), image.Pt(128, 72), image.Pt(144, 72)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := createScreenshotController(t, tc.raw.X, tc.raw.Y)
			require.NoError(t, ctrl.SetScreenshot(WithScreenshotTargetExpand(int32(tc.target.X), int32(tc.target.Y))))
			requireScreenshotSize(t, ctrl, tc.want.X, tc.want.Y)
		})
	}
}

func TestController_ScreenshotTargetExpand_ModeSwitching(t *testing.T) {
	ctrl := createScreenshotController(t, 400, 200)
	// Each step uses the same controller to verify that previous targets are reset
	// and cached output dimensions are refreshed when the mode changes.
	for _, step := range []struct {
		name string
		opts []ScreenshotOption
		want image.Point
	}{
		{"long side", []ScreenshotOption{WithScreenshotTargetLongSide(200)}, image.Pt(200, 100)},
		{"expand replaces long side", []ScreenshotOption{WithScreenshotTargetExpand(128, 72)}, image.Pt(144, 72)},
		{"short side replaces expand", []ScreenshotOption{WithScreenshotTargetShortSide(50)}, image.Pt(100, 50)},
		{"expand replaces short side", []ScreenshotOption{WithScreenshotTargetExpand(128, 72)}, image.Pt(144, 72)},
		{"raw size", []ScreenshotOption{WithScreenshotUseRawSize(true)}, image.Pt(400, 200)},
		{"expand ignored while raw", []ScreenshotOption{WithScreenshotTargetExpand(80, 20)}, image.Pt(400, 200)},
		{"restore expand", []ScreenshotOption{WithScreenshotUseRawSize(false)}, image.Pt(80, 40)},
		{"last expand wins", []ScreenshotOption{WithScreenshotTargetLongSide(200), WithScreenshotTargetExpand(128, 72)}, image.Pt(144, 72)},
		{"last long side wins", []ScreenshotOption{WithScreenshotTargetExpand(128, 72), WithScreenshotTargetLongSide(200)}, image.Pt(200, 100)},
	} {
		t.Run(step.name, func(t *testing.T) {
			require.NoError(t, ctrl.SetScreenshot(step.opts...))
			requireScreenshotSize(t, ctrl, step.want.X, step.want.Y)
		})
	}
}

func TestController_ScreenshotTargetExpand_InvalidSize(t *testing.T) {
	ctrl := createScreenshotController(t, 400, 200)
	require.NoError(t, ctrl.SetScreenshot(WithScreenshotTargetExpand(128, 72)))
	for _, size := range [][2]int32{{0, 72}, {128, 0}, {-1, 72}, {128, -1}} {
		require.Error(t, ctrl.SetScreenshot(WithScreenshotTargetExpand(size[0], size[1])))
		// A rejected update must not replace the previously valid target.
		requireScreenshotSize(t, ctrl, 144, 72)
	}
}
