package maa

import (
	"fmt"
	"image"
	"image/color"
	"testing"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

// gradientPattern is a small opaque grayscale ramp. Upscaling it by two keeps
// nearest-neighbor output blocky while linear interpolation introduces
// intermediate values, so the same pattern distinguishes both methods.
var gradientPattern = repeatGrayRow([]uint8{0, 85, 170, 255}, 4)

// gradientNearest8x8 doubles the pattern by repeating every source pixel.
var gradientNearest8x8 = repeatGrayRow([]uint8{0, 0, 85, 85, 170, 170, 255, 255}, 8)

// gradientLinear8x8 is the same upscale through cv::INTER_LINEAR, observed
// through the real custom-controller screenshot pipeline.
var gradientLinear8x8 = repeatGrayRow([]uint8{0, 21, 64, 106, 149, 191, 234, 255}, 8)

func repeatGrayRow(row []uint8, height int) [][]uint8 {
	rows := make([][]uint8, height)
	for i := range rows {
		rows[i] = row
	}
	return rows
}

// screenshotOptionPermutations returns every ordering of three composed options.
func screenshotOptionPermutations(a, b, c ScreenshotOption) [][3]ScreenshotOption {
	opts := [3]ScreenshotOption{a, b, c}
	orders := [][3]int{
		{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0},
	}
	perms := make([][3]ScreenshotOption, 0, len(orders))
	for _, order := range orders {
		perms = append(perms, [3]ScreenshotOption{opts[order[0]], opts[order[1]], opts[order[2]]})
	}
	return perms
}

func createPatternScreenshotController(t *testing.T, img image.Image) *Controller {
	t.Helper()
	ctrl, err := NewCustomController(&fixedScreenshotController{img: img})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, ctrl.Destroy()) })
	job, err := ctrl.PostConnect()
	require.NoError(t, err)
	require.True(t, job.Wait().Success())
	return ctrl
}

func captureScreenshot(t *testing.T, ctrl *Controller) image.Image {
	t.Helper()
	job, err := ctrl.PostScreencap()
	require.NoError(t, err)
	require.True(t, job.Wait().Success())
	img, err := ctrl.CacheImage()
	require.NoError(t, err)
	require.NotNil(t, img)
	return img
}

// grayRows builds an opaque grayscale image from row-major 8-bit values.
func grayRows(rows [][]uint8) image.Image {
	height := len(rows)
	width := len(rows[0])
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y, row := range rows {
		for x, value := range row {
			img.SetRGBA(x, y, color.RGBA{R: value, G: value, B: value, A: 255})
		}
	}
	return img
}

func requireGrayPixels(t *testing.T, img image.Image, want [][]uint8) {
	t.Helper()
	require.Equal(t, image.Pt(len(want[0]), len(want)), img.Bounds().Size())
	got := make([][]uint8, len(want))
	for y := range want {
		got[y] = make([]uint8, len(want[y]))
		for x := range want[y] {
			r, _, _, _ := img.At(x, y).RGBA()
			got[y][x] = uint8(r >> 8)
		}
	}
	require.Equal(t, want, got)
}

type setOptionCall struct {
	key  native.MaaCtrlOption
	vals [2]int32
}

func setOptionKeys(calls []setOptionCall) []native.MaaCtrlOption {
	keys := make([]native.MaaCtrlOption, len(calls))
	for i, call := range calls {
		keys[i] = call.key
	}
	return keys
}

// installSetOptionRecorder replaces the native setter with a spy that records
// calls and forwards them to the real setter. When failKey is not
// MaaCtrlOption_Invalid, that single key fails instead. The original function
// is always restored through cleanup; tests that use this helper must not run
// in parallel.
func installSetOptionRecorder(t *testing.T, failKey native.MaaCtrlOption) (*[]setOptionCall, func()) {
	t.Helper()
	old := native.MaaControllerSetOption
	calls := make([]setOptionCall, 0, 4)
	native.MaaControllerSetOption = func(ctrl uintptr, key native.MaaCtrlOption, value unsafe.Pointer, valSize uint64) bool {
		call := setOptionCall{key: key}
		if key == native.MaaCtrlOption_ScreenshotUseRawSize {
			if *(*bool)(value) {
				call.vals[0] = 1
			}
		} else {
			if valSize >= uint64(unsafe.Sizeof(int32(0))) {
				call.vals[0] = *(*int32)(value)
			}
			if valSize >= 2*uint64(unsafe.Sizeof(int32(0))) {
				call.vals[1] = *(*int32)(unsafe.Add(value, unsafe.Sizeof(int32(0))))
			}
		}
		calls = append(calls, call)
		if key == failKey {
			return false
		}
		return old(ctrl, key, value, valSize)
	}
	restored := false
	restore := func() {
		if !restored {
			native.MaaControllerSetOption = old
			restored = true
		}
	}
	t.Cleanup(restore)
	return &calls, restore
}

func TestController_ScreenshotOptions_TargetMethodRawPermutations(t *testing.T) {
	targets := []struct {
		name string
		opt  ScreenshotOption
		want image.Point
	}{
		{"long side", WithScreenshotTargetLongSide(300), image.Pt(300, 150)},
		{"short side", WithScreenshotTargetShortSide(100), image.Pt(200, 100)},
		{"expand", WithScreenshotTargetExpand(120, 60), image.Pt(120, 60)},
	}
	for _, target := range targets {
		t.Run(target.name, func(t *testing.T) {
			ctrl := createScreenshotController(t, 400, 200)
			perms := screenshotOptionPermutations(
				target.opt,
				WithScreenshotResizeMethod(ScreenshotResizeMethodLinear),
				WithScreenshotUseRawSize(false),
			)
			for i, perm := range perms {
				t.Run(fmt.Sprintf("order_%d", i+1), func(t *testing.T) {
					// Start each ordering with a different target and raw mode active,
					// so earlier successful calls cannot hide a skipped setting.
					require.NoError(t, ctrl.SetScreenshot(
						WithScreenshotTargetLongSide(80),
						WithScreenshotResizeMethod(ScreenshotResizeMethodNearestNeighbor),
					))
					require.NoError(t, ctrl.SetScreenshot(WithScreenshotUseRawSize(true)))
					require.NoError(t, ctrl.SetScreenshot(perm[0], perm[1], perm[2]))
					requireScreenshotSize(t, ctrl, target.want.X, target.want.Y)
				})
			}
		})
	}
}

func TestController_ScreenshotOptions_InterpolationPixels(t *testing.T) {
	ctrl := createPatternScreenshotController(t, grayRows(gradientPattern))
	for _, tc := range []struct {
		name           string
		opts           []ScreenshotOption
		want           [][]uint8
		baselineMethod ScreenshotResizeMethod
	}{
		{
			"nearest target first",
			[]ScreenshotOption{
				WithScreenshotTargetLongSide(8),
				WithScreenshotResizeMethod(ScreenshotResizeMethodNearestNeighbor),
				WithScreenshotUseRawSize(false),
			},
			gradientNearest8x8,
			ScreenshotResizeMethodLinear,
		},
		{
			"nearest method first",
			[]ScreenshotOption{
				WithScreenshotResizeMethod(ScreenshotResizeMethodNearestNeighbor),
				WithScreenshotTargetLongSide(8),
				WithScreenshotUseRawSize(false),
			},
			gradientNearest8x8,
			ScreenshotResizeMethodLinear,
		},
		{
			"linear target first",
			[]ScreenshotOption{
				WithScreenshotTargetLongSide(8),
				WithScreenshotResizeMethod(ScreenshotResizeMethodLinear),
				WithScreenshotUseRawSize(false),
			},
			gradientLinear8x8,
			ScreenshotResizeMethodNearestNeighbor,
		},
		{
			"linear raw size first",
			[]ScreenshotOption{
				WithScreenshotUseRawSize(false),
				WithScreenshotResizeMethod(ScreenshotResizeMethodLinear),
				WithScreenshotTargetLongSide(8),
			},
			gradientLinear8x8,
			ScreenshotResizeMethodNearestNeighbor,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Each case must change the target, method, and raw mode to produce
			// its expected pixels, even if an earlier case already succeeded.
			require.NoError(t, ctrl.SetScreenshot(
				WithScreenshotTargetLongSide(4),
				WithScreenshotResizeMethod(tc.baselineMethod),
			))
			require.NoError(t, ctrl.SetScreenshot(WithScreenshotUseRawSize(true)))
			require.NoError(t, ctrl.SetScreenshot(tc.opts...))
			requireGrayPixels(t, captureScreenshot(t, ctrl), tc.want)
		})
	}
}

func TestController_ScreenshotOptions_RawRetainsTargetAndMethod(t *testing.T) {
	ctrl := createPatternScreenshotController(t, grayRows(gradientPattern))
	for i, opts := range [][]ScreenshotOption{
		{WithScreenshotUseRawSize(true), WithScreenshotResizeMethod(ScreenshotResizeMethodLinear)},
		{WithScreenshotResizeMethod(ScreenshotResizeMethodLinear), WithScreenshotUseRawSize(true)},
	} {
		t.Run(fmt.Sprintf("order_%d", i+1), func(t *testing.T) {
			require.NoError(t, ctrl.SetScreenshot(
				WithScreenshotTargetLongSide(8),
				WithScreenshotResizeMethod(ScreenshotResizeMethodNearestNeighbor),
			))
			requireGrayPixels(t, captureScreenshot(t, ctrl), gradientNearest8x8)

			// Enabling raw size keeps the target and accepts a new interpolation method.
			require.NoError(t, ctrl.SetScreenshot(opts...))
			requireGrayPixels(t, captureScreenshot(t, ctrl), gradientPattern)

			// Disabling raw size resumes scaling with the retained target and the
			// interpolation method stored while raw mode was active.
			require.NoError(t, ctrl.SetScreenshot(WithScreenshotUseRawSize(false)))
			requireGrayPixels(t, captureScreenshot(t, ctrl), gradientLinear8x8)
		})
	}
}

func TestController_ScreenshotOptions_RejectsInvalidInput(t *testing.T) {
	ctrl := createScreenshotController(t, 400, 200)
	require.NoError(t, ctrl.SetScreenshot(WithScreenshotTargetLongSide(300)))
	requireScreenshotSize(t, ctrl, 300, 150)

	calls, _ := installSetOptionRecorder(t, native.MaaCtrlOption_Invalid)

	for _, tc := range []struct {
		name string
		opts []ScreenshotOption
		want string
	}{
		{
			"long side zero",
			[]ScreenshotOption{WithScreenshotTargetLongSide(0)},
			"invalid screenshot target long side: 0",
		},
		{
			"long side negative",
			[]ScreenshotOption{WithScreenshotTargetLongSide(-1)},
			"invalid screenshot target long side: -1",
		},
		{
			"short side zero",
			[]ScreenshotOption{WithScreenshotTargetShortSide(0)},
			"invalid screenshot target short side: 0",
		},
		{
			"short side negative",
			[]ScreenshotOption{WithScreenshotTargetShortSide(-1)},
			"invalid screenshot target short side: -1",
		},
		{
			"expand zero width",
			[]ScreenshotOption{WithScreenshotTargetExpand(0, 72)},
			"invalid screenshot target expand: 0x72",
		},
		{
			"expand zero height",
			[]ScreenshotOption{WithScreenshotTargetExpand(128, 0)},
			"invalid screenshot target expand: 128x0",
		},
		{
			"expand negative width",
			[]ScreenshotOption{WithScreenshotTargetExpand(-1, 72)},
			"invalid screenshot target expand: -1x72",
		},
		{
			"expand negative height",
			[]ScreenshotOption{WithScreenshotTargetExpand(128, -1)},
			"invalid screenshot target expand: 128x-1",
		},
		{
			"interpolation below range",
			[]ScreenshotOption{WithScreenshotResizeMethod(-1)},
			"invalid screenshot resize method: -1",
		},
		{
			"interpolation above range",
			[]ScreenshotOption{WithScreenshotResizeMethod(5)},
			"invalid screenshot resize method: 5",
		},
		{
			"interpolation far above range",
			[]ScreenshotOption{WithScreenshotResizeMethod(6)},
			"invalid screenshot resize method: 6",
		},
		{
			"duplicate long side",
			[]ScreenshotOption{WithScreenshotTargetLongSide(100), WithScreenshotTargetLongSide(100)},
			"duplicate screenshot option: target long side",
		},
		{
			"duplicate short side",
			[]ScreenshotOption{WithScreenshotTargetShortSide(100), WithScreenshotTargetShortSide(100)},
			"duplicate screenshot option: target short side",
		},
		{
			"duplicate expand",
			[]ScreenshotOption{WithScreenshotTargetExpand(100, 50), WithScreenshotTargetExpand(100, 50)},
			"duplicate screenshot option: target expand",
		},
		{
			"duplicate interpolation",
			[]ScreenshotOption{
				WithScreenshotResizeMethod(ScreenshotResizeMethodLinear),
				WithScreenshotResizeMethod(ScreenshotResizeMethodLinear),
			},
			"duplicate screenshot option: resize method",
		},
		{
			"duplicate raw true",
			[]ScreenshotOption{WithScreenshotUseRawSize(true), WithScreenshotUseRawSize(true)},
			"duplicate screenshot option: use raw size",
		},
		{
			"duplicate raw false",
			[]ScreenshotOption{WithScreenshotUseRawSize(false), WithScreenshotUseRawSize(false)},
			"duplicate screenshot option: use raw size",
		},
		{
			"duplicate raw true then false",
			[]ScreenshotOption{WithScreenshotUseRawSize(true), WithScreenshotUseRawSize(false)},
			"duplicate screenshot option: use raw size",
		},
		{
			"duplicate raw false then true",
			[]ScreenshotOption{WithScreenshotUseRawSize(false), WithScreenshotUseRawSize(true)},
			"duplicate screenshot option: use raw size",
		},
		{
			"conflict long side then short side",
			[]ScreenshotOption{WithScreenshotTargetLongSide(200), WithScreenshotTargetShortSide(100)},
			"conflicting screenshot targets: target long side and target short side",
		},
		{
			"conflict short side then long side",
			[]ScreenshotOption{WithScreenshotTargetShortSide(100), WithScreenshotTargetLongSide(200)},
			"conflicting screenshot targets: target short side and target long side",
		},
		{
			"conflict long side then expand",
			[]ScreenshotOption{WithScreenshotTargetLongSide(200), WithScreenshotTargetExpand(128, 72)},
			"conflicting screenshot targets: target long side and target expand",
		},
		{
			"conflict expand then long side",
			[]ScreenshotOption{WithScreenshotTargetExpand(128, 72), WithScreenshotTargetLongSide(200)},
			"conflicting screenshot targets: target expand and target long side",
		},
		{
			"conflict short side then expand",
			[]ScreenshotOption{WithScreenshotTargetShortSide(100), WithScreenshotTargetExpand(128, 72)},
			"conflicting screenshot targets: target short side and target expand",
		},
		{
			"conflict expand then short side",
			[]ScreenshotOption{WithScreenshotTargetExpand(128, 72), WithScreenshotTargetShortSide(100)},
			"conflicting screenshot targets: target expand and target short side",
		},
		{
			"conflict raw true then long side",
			[]ScreenshotOption{WithScreenshotUseRawSize(true), WithScreenshotTargetLongSide(200)},
			"conflicting screenshot options: use raw size true and target long side",
		},
		{
			"conflict long side then raw true",
			[]ScreenshotOption{WithScreenshotTargetLongSide(200), WithScreenshotUseRawSize(true)},
			"conflicting screenshot options: use raw size true and target long side",
		},
		{
			"conflict raw true then short side",
			[]ScreenshotOption{WithScreenshotUseRawSize(true), WithScreenshotTargetShortSide(100)},
			"conflicting screenshot options: use raw size true and target short side",
		},
		{
			"conflict short side then raw true",
			[]ScreenshotOption{WithScreenshotTargetShortSide(100), WithScreenshotUseRawSize(true)},
			"conflicting screenshot options: use raw size true and target short side",
		},
		{
			"conflict raw true then expand",
			[]ScreenshotOption{WithScreenshotUseRawSize(true), WithScreenshotTargetExpand(128, 72)},
			"conflicting screenshot options: use raw size true and target expand",
		},
		{
			"conflict expand then raw true",
			[]ScreenshotOption{WithScreenshotTargetExpand(128, 72), WithScreenshotUseRawSize(true)},
			"conflicting screenshot options: use raw size true and target expand",
		},
		{
			"valid target then invalid interpolation",
			[]ScreenshotOption{WithScreenshotTargetLongSide(200), WithScreenshotResizeMethod(5)},
			"screenshot option 2: invalid screenshot resize method: 5",
		},
		{
			"valid interpolation then invalid target",
			[]ScreenshotOption{WithScreenshotResizeMethod(ScreenshotResizeMethodLinear), WithScreenshotTargetShortSide(0)},
			"screenshot option 2: invalid screenshot target short side: 0",
		},
		{
			"valid raw false then invalid expand",
			[]ScreenshotOption{WithScreenshotUseRawSize(false), WithScreenshotTargetExpand(0, 0)},
			"screenshot option 2: invalid screenshot target expand: 0x0",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := len(*calls)
			err := ctrl.SetScreenshot(tc.opts...)
			require.ErrorContains(t, err, tc.want)
			// Prevalidation must reject the whole call before any native setter runs.
			require.Len(t, *calls, before)
		})
	}

	requireScreenshotSize(t, ctrl, 300, 150)
}

func TestController_ScreenshotOptions_PrevalidationKeepsState(t *testing.T) {
	ctrl := createScreenshotController(t, 400, 200)
	require.NoError(t, ctrl.SetScreenshot(WithScreenshotTargetLongSide(300)))
	requireScreenshotSize(t, ctrl, 300, 150)

	for _, opts := range [][]ScreenshotOption{
		{WithScreenshotTargetLongSide(200), WithScreenshotResizeMethod(5)},
		{WithScreenshotTargetShortSide(0), WithScreenshotResizeMethod(ScreenshotResizeMethodLinear)},
		{WithScreenshotUseRawSize(true), WithScreenshotTargetLongSide(200)},
	} {
		require.Error(t, ctrl.SetScreenshot(opts...))
		requireScreenshotSize(t, ctrl, 300, 150)
	}
}

func TestController_ScreenshotOptions_AcceptsBoundaryValues(t *testing.T) {
	t.Run("minimum positive targets", func(t *testing.T) {
		ctrl := createScreenshotController(t, 100, 100)
		for _, tc := range []struct {
			name string
			opt  ScreenshotOption
		}{
			{"long side", WithScreenshotTargetLongSide(1)},
			{"short side", WithScreenshotTargetShortSide(1)},
			{"expand", WithScreenshotTargetExpand(1, 1)},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.NoError(t, ctrl.SetScreenshot(tc.opt))
				requireScreenshotSize(t, ctrl, 1, 1)
			})
		}
	})

	t.Run("all interpolation values", func(t *testing.T) {
		ctrl := createScreenshotController(t, 100, 100)
		for method := ScreenshotResizeMethod(0); method <= 4; method++ {
			t.Run(fmt.Sprintf("method_%d", method), func(t *testing.T) {
				require.NoError(t, ctrl.SetScreenshot(
					WithScreenshotTargetLongSide(50),
					WithScreenshotResizeMethod(method),
				))
				requireScreenshotSize(t, ctrl, 50, 50)
			})
		}
	})
}

func TestController_ScreenshotOptions_EmptyAndNilAreNoops(t *testing.T) {
	ctrl := createScreenshotController(t, 400, 200)
	require.NoError(t, ctrl.SetScreenshot(WithScreenshotTargetLongSide(300)))
	requireScreenshotSize(t, ctrl, 300, 150)

	require.NoError(t, ctrl.SetScreenshot())
	require.NoError(t, ctrl.SetScreenshot(nil))
	requireScreenshotSize(t, ctrl, 300, 150)

	// Nil options interspersed with valid ones do not block application.
	require.NoError(t, ctrl.SetScreenshot(
		nil,
		WithScreenshotTargetExpand(120, 60),
		nil,
		WithScreenshotResizeMethod(ScreenshotResizeMethodLinear),
		WithScreenshotUseRawSize(false),
		nil,
	))
	requireScreenshotSize(t, ctrl, 120, 60)

	calls, _ := installSetOptionRecorder(t, native.MaaCtrlOption_Invalid)
	require.NoError(t, ctrl.SetScreenshot())
	require.NoError(t, ctrl.SetScreenshot(nil, nil))
	require.Empty(t, *calls)

	require.NoError(t, ctrl.SetScreenshot(nil, WithScreenshotTargetLongSide(300), nil))
	require.Equal(t, []native.MaaCtrlOption{native.MaaCtrlOption_ScreenshotTargetLongSide}, setOptionKeys(*calls))
}

func TestController_ScreenshotOptions_ClosedController(t *testing.T) {
	calls, _ := installSetOptionRecorder(t, native.MaaCtrlOption_Invalid)

	var zero Controller
	require.ErrorIs(t, zero.SetScreenshot(WithScreenshotTargetLongSide(100)), ErrClosed)
	require.ErrorIs(t, zero.SetScreenshot(), ErrClosed)
	require.ErrorIs(t, zero.SetScreenshot(nil), ErrClosed)

	ctrl, err := NewCustomController(&fixedScreenshotController{
		img: image.NewRGBA(image.Rect(0, 0, 400, 200)),
	})
	require.NoError(t, err)
	require.NoError(t, ctrl.Destroy())
	require.ErrorIs(t, ctrl.SetScreenshot(WithScreenshotTargetLongSide(100)), ErrClosed)
	require.ErrorIs(t, ctrl.SetScreenshot(), ErrClosed)
	require.ErrorIs(t, ctrl.SetScreenshot(nil), ErrClosed)

	require.Empty(t, *calls)
}

func TestController_ScreenshotOptions_NativeApplicationOrder(t *testing.T) {
	ctrl := createScreenshotController(t, 400, 200)
	calls, _ := installSetOptionRecorder(t, native.MaaCtrlOption_Invalid)

	// Options are applied natively as target, interpolation, then raw-size mode,
	// no matter how they are ordered by the caller.
	require.NoError(t, ctrl.SetScreenshot(
		WithScreenshotUseRawSize(false),
		WithScreenshotResizeMethod(ScreenshotResizeMethodLinear),
		WithScreenshotTargetLongSide(300),
	))
	require.Equal(t, []setOptionCall{
		{key: native.MaaCtrlOption_ScreenshotTargetLongSide, vals: [2]int32{300, 0}},
		{key: native.MaaCtrlOption_ScreenshotResizeMethod, vals: [2]int32{1, 0}},
		{key: native.MaaCtrlOption_ScreenshotUseRawSize, vals: [2]int32{0, 0}},
	}, *calls)

	before := len(*calls)
	require.NoError(t, ctrl.SetScreenshot(
		WithScreenshotTargetLongSide(200),
		WithScreenshotUseRawSize(false),
		WithScreenshotResizeMethod(ScreenshotResizeMethodArea),
	))
	require.Equal(t, []native.MaaCtrlOption{
		native.MaaCtrlOption_ScreenshotTargetLongSide,
		native.MaaCtrlOption_ScreenshotResizeMethod,
		native.MaaCtrlOption_ScreenshotUseRawSize,
	}, setOptionKeys((*calls)[before:]))
}

func TestController_ScreenshotOptions_NativeFailureStops(t *testing.T) {
	composed := []ScreenshotOption{
		WithScreenshotTargetLongSide(8),
		WithScreenshotResizeMethod(ScreenshotResizeMethodLinear),
		WithScreenshotUseRawSize(false),
	}
	createRawController := func(t *testing.T) *Controller {
		t.Helper()
		ctrl := createPatternScreenshotController(t, grayRows(gradientPattern))
		require.NoError(t, ctrl.SetScreenshot(WithScreenshotUseRawSize(true)))
		requireGrayPixels(t, captureScreenshot(t, ctrl), gradientPattern)
		return ctrl
	}

	t.Run("first setting fails", func(t *testing.T) {
		ctrl := createRawController(t)
		calls, _ := installSetOptionRecorder(t, native.MaaCtrlOption_ScreenshotTargetLongSide)

		err := ctrl.SetScreenshot(composed...)
		require.ErrorContains(t, err, "failed to set screenshot target long side")
		require.Equal(t, []native.MaaCtrlOption{native.MaaCtrlOption_ScreenshotTargetLongSide}, setOptionKeys(*calls))
		// The raw-size baseline is untouched because the failing setter is first.
		requireGrayPixels(t, captureScreenshot(t, ctrl), gradientPattern)
	})

	t.Run("middle setting fails", func(t *testing.T) {
		ctrl := createRawController(t)
		calls, restore := installSetOptionRecorder(t, native.MaaCtrlOption_ScreenshotResizeMethod)

		err := ctrl.SetScreenshot(composed...)
		require.ErrorContains(t, err, "failed to set screenshot resize method")
		require.Equal(t, []native.MaaCtrlOption{
			native.MaaCtrlOption_ScreenshotTargetLongSide,
			native.MaaCtrlOption_ScreenshotResizeMethod,
		}, setOptionKeys(*calls))
		requireGrayPixels(t, captureScreenshot(t, ctrl), gradientPattern)

		// The target applied before the failure is retained; raw-size mode was
		// never disabled because the failing setter stopped the call.
		restore()
		require.NoError(t, ctrl.SetScreenshot(WithScreenshotUseRawSize(false)))
		requireScreenshotSize(t, ctrl, 8, 8)
	})

	t.Run("last setting fails", func(t *testing.T) {
		ctrl := createRawController(t)
		calls, restore := installSetOptionRecorder(t, native.MaaCtrlOption_ScreenshotUseRawSize)

		err := ctrl.SetScreenshot(composed...)
		require.ErrorContains(t, err, "failed to set screenshot use raw size")
		require.Equal(t, []native.MaaCtrlOption{
			native.MaaCtrlOption_ScreenshotTargetLongSide,
			native.MaaCtrlOption_ScreenshotResizeMethod,
			native.MaaCtrlOption_ScreenshotUseRawSize,
		}, setOptionKeys(*calls))
		requireGrayPixels(t, captureScreenshot(t, ctrl), gradientPattern)

		// Earlier target and interpolation changes remain applied; only raw-size
		// mode was skipped.
		restore()
		require.NoError(t, ctrl.SetScreenshot(WithScreenshotUseRawSize(false)))
		requireGrayPixels(t, captureScreenshot(t, ctrl), gradientLinear8x8)
	})
}
