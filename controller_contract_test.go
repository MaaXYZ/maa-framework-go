package maa

import (
	"image"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v4/controller/adb"
	"github.com/MaaXYZ/maa-framework-go/v4/controller/macos"
	"github.com/MaaXYZ/maa-framework-go/v4/controller/win32"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

// waitFor polls cond until it holds or the deadline expires.
func waitFor(t *testing.T, timeout time.Duration, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return cond()
}

// recordingSink records controller action events for dispatch assertions.
type recordingSink struct {
	mu      sync.Mutex
	events  []EventStatus
	details []ControllerActionDetail
}

func (s *recordingSink) OnControllerAction(ctrl *Controller, event EventStatus, detail ControllerActionDetail) {
	// The controller view handed to the sink must be usable: reading the
	// UUID inside the callback goes through the same borrowed handle.
	_, _ = ctrl.GetUUID()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
	s.details = append(s.details, detail)
}

func (s *recordingSink) snapshot() (events []EventStatus, details []ControllerActionDetail) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]EventStatus{}, s.events...), append([]ControllerActionDetail{}, s.details...)
}

func (s *recordingSink) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.events)
}

func (s *recordingSink) find(action string, event EventStatus) (ControllerActionDetail, bool) {
	events, details := s.snapshot()
	for i, e := range events {
		if e == event && details[i].Action == action {
			return details[i], true
		}
	}
	return ControllerActionDetail{}, false
}

// TestController_SinkDispatch pins the Controller.Action event contract: a
// sink registered via AddSink receives Starting and Succeeded events for a
// posted action, with the detail JSON decoded into ControllerActionDetail and
// a usable borrowed controller view; RemoveSink and ClearSinks stop delivery.
func TestController_SinkDispatch(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	require.True(t, connectJob.Wait().Success())

	sink := &recordingSink{}
	sinkID := ctrl.AddSink(sink)
	require.NotZero(t, sinkID)

	clickJob, err := ctrl.PostClick(11, 22)
	require.NoError(t, err)
	require.True(t, clickJob.Wait().Success())

	require.True(t, waitFor(t, 5*time.Second, func() bool {
		_, okStart := sink.find("click", EventStatusStarting)
		_, okDone := sink.find("click", EventStatusSucceeded)
		return okStart && okDone
	}), "expected click Starting and Succeeded events, got statuses %v", func() []EventStatus {
		events, _ := sink.snapshot()
		return events
	}())

	startDetail, ok := sink.find("click", EventStatusStarting)
	require.True(t, ok)
	require.Equal(t, "blank-controller", startDetail.UUID)
	require.Equal(t, "custom", startDetail.Info["type"])
	// ClickParam serializes as {"point": [x, y], "contact": n, "pressure": n}.
	require.Equal(t, []any{float64(11), float64(22)}, startDetail.Param["point"])
	require.Equal(t, float64(0), startDetail.Param["contact"])
	require.Equal(t, float64(1), startDetail.Param["pressure"])
	require.NotZero(t, startDetail.CtrlID)

	// RemoveSink stops delivery of later actions. Dispatch happens while the
	// action executes, so a completed job plus a grace window proves the
	// removed sink received nothing.
	ctrl.RemoveSink(sinkID)
	countAfterRemove := sink.count()
	scrollJob, err := ctrl.PostScroll(120, -120)
	require.NoError(t, err)
	require.True(t, scrollJob.Wait().Success())
	time.Sleep(200 * time.Millisecond)
	require.Equal(t, countAfterRemove, sink.count(), "events must stop after RemoveSink")

	// ClearSinks also stops delivery.
	sink2 := &recordingSink{}
	require.NotZero(t, ctrl.AddSink(sink2))
	ctrl.ClearSinks()
	countAfterClear := sink2.count()
	keyJob, err := ctrl.PostClickKey(4)
	require.NoError(t, err)
	require.True(t, keyJob.Wait().Success())
	time.Sleep(200 * time.Millisecond)
	require.Equal(t, countAfterClear, sink2.count(), "events must stop after ClearSinks")

	// OnControllerAction routes the same events through the function adapter.
	var adapterMu sync.Mutex
	adapterSeen := 0
	adapterID := ctrl.OnControllerAction(func(EventStatus, ControllerActionDetail) {
		adapterMu.Lock()
		defer adapterMu.Unlock()
		adapterSeen++
	})
	require.NotZero(t, adapterID)
	inactiveJob, err := ctrl.PostInactive()
	require.NoError(t, err)
	require.True(t, inactiveJob.Wait().Success())
	require.True(t, waitFor(t, 5*time.Second, func() bool {
		adapterMu.Lock()
		defer adapterMu.Unlock()
		return adapterSeen > 0
	}), "OnControllerAction adapter must receive events")
}

// TestController_SetMouseLockFollow_NonWin32Fails pins that mouse-lock-follow
// is rejected for controllers that are not Win32 message-input controllers:
// the blank controller is a custom controller, so the native option call fails.
func TestController_SetMouseLockFollow_NonWin32Fails(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	require.Error(t, ctrl.SetMouseLockFollow(true))
	require.Error(t, ctrl.SetMouseLockFollow(false))
}

// TestNewControllers_NativeCreateFailure pins the constructor failure
// contract: when native creation returns a null handle, every constructor
// reports an error and returns a nil controller.
func TestNewControllers_NativeCreateFailure(t *testing.T) {
	t.Run("Adb", func(t *testing.T) {
		old := native.MaaAdbControllerCreate
		native.MaaAdbControllerCreate = func(string, string, uint64, uint64, string, string) uintptr { return 0 }
		defer func() { native.MaaAdbControllerCreate = old }()
		ctrl, err := NewAdbController("adb", "127.0.0.1:5555", adb.ScreencapDefault, adb.InputDefault, "{}", "")
		require.Error(t, err)
		require.Nil(t, ctrl)
	})

	t.Run("PlayCover", func(t *testing.T) {
		old := native.MaaPlayCoverControllerCreate
		native.MaaPlayCoverControllerCreate = func(string, string) uintptr { return 0 }
		defer func() { native.MaaPlayCoverControllerCreate = old }()
		ctrl, err := NewPlayCoverController("127.0.0.1:5555", "com.example.app")
		require.Error(t, err)
		require.Nil(t, ctrl)
	})

	t.Run("Win32", func(t *testing.T) {
		old := native.MaaWin32ControllerCreate
		native.MaaWin32ControllerCreate = func(unsafe.Pointer, uint64, uint64, uint64) uintptr { return 0 }
		defer func() { native.MaaWin32ControllerCreate = old }()
		ctrl, err := NewWin32Controller(nil, win32.ScreencapForeground, win32.InputSeize, win32.InputSeize)
		require.Error(t, err)
		require.Nil(t, ctrl)
	})

	t.Run("MacOS", func(t *testing.T) {
		old := native.MaaMacOSControllerCreate
		native.MaaMacOSControllerCreate = func(uint32, native.MaaMacOSScreencapMethod, native.MaaMacOSInputMethod) uintptr { return 0 }
		defer func() { native.MaaMacOSControllerCreate = old }()
		ctrl, err := NewMacOSController(0, macos.ScreencapScreenCaptureKit, macos.InputGlobalEvent)
		require.Error(t, err)
		require.Nil(t, ctrl)
	})

	t.Run("Gamepad", func(t *testing.T) {
		old := native.MaaGamepadControllerCreate
		native.MaaGamepadControllerCreate = func(unsafe.Pointer, native.MaaGamepadType, uint64) uintptr { return 0 }
		defer func() { native.MaaGamepadControllerCreate = old }()
		ctrl, err := NewGamepadController(nil, GamepadTypeXbox360, 0)
		require.Error(t, err)
		require.Nil(t, ctrl)
	})

	t.Run("AndroidNative", func(t *testing.T) {
		old := native.MaaAndroidNativeControllerCreate
		native.MaaAndroidNativeControllerCreate = func(string) uintptr { return 0 }
		defer func() { native.MaaAndroidNativeControllerCreate = old }()
		ctrl, err := NewAndroidNativeController("{}")
		require.Error(t, err)
		require.Nil(t, ctrl)
	})

	t.Run("Replay", func(t *testing.T) {
		old := native.MaaReplayControllerCreate
		native.MaaReplayControllerCreate = func(string) uintptr { return 0 }
		defer func() { native.MaaReplayControllerCreate = old }()
		ctrl, err := NewReplayController(t.TempDir() + "/replay.jsonl")
		require.Error(t, err)
		require.Nil(t, ctrl)
	})

	t.Run("Record", func(t *testing.T) {
		inner := createBlankController(t)
		defer inner.Destroy()
		old := native.MaaRecordControllerCreate
		native.MaaRecordControllerCreate = func(uintptr, string) uintptr { return 0 }
		defer func() { native.MaaRecordControllerCreate = old }()
		ctrl, err := NewRecordController(inner, t.TempDir()+"/record.jsonl")
		require.Error(t, err)
		require.Nil(t, ctrl)
	})

	t.Run("Custom", func(t *testing.T) {
		old := native.MaaCustomControllerCreate
		native.MaaCustomControllerCreate = func(unsafe.Pointer, uintptr) uintptr { return 0 }
		defer func() { native.MaaCustomControllerCreate = old }()
		ctrl, err := NewCustomController(&BlankController{})
		require.Error(t, err)
		require.Nil(t, ctrl)
	})
}

// TestNewCustomController_Nil pins the nil-guard contract.
func TestNewCustomController_Nil(t *testing.T) {
	ctrl, err := NewCustomController(nil)
	require.Error(t, err)
	require.Nil(t, ctrl)
}

// TestNewRecordController_Nil pins the nil-guard contract.
func TestNewRecordController_Nil(t *testing.T) {
	ctrl, err := NewRecordController(nil, "/tmp/unused.jsonl")
	require.Error(t, err)
	require.Nil(t, ctrl)
}

// shellRecorder is a custom controller whose Shell records the command and
// timeout it was invoked with and returns a fixed output.
type shellRecorder struct {
	BlankController
	mu      sync.Mutex
	cmd     string
	timeout int64
}

func (s *shellRecorder) Shell(cmd string, timeout int64) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cmd = cmd
	s.timeout = timeout
	return "shell-output-42", true
}

// TestController_PostShellAndGetShellOutput pins the shell contract on a
// custom controller implementing Shell: PostShell succeeds, the timeout is
// forwarded in milliseconds, and GetShellOutput returns the callback's output.
func TestController_PostShellAndGetShellOutput(t *testing.T) {
	rec := &shellRecorder{}
	ctrl, err := NewCustomController(rec)
	require.NoError(t, err)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	require.True(t, connectJob.Wait().Success())

	shellJob, err := ctrl.PostShell("echo hi", 3*time.Second)
	require.NoError(t, err)
	require.True(t, shellJob.Wait().Success())

	rec.mu.Lock()
	require.Equal(t, "echo hi", rec.cmd)
	require.Equal(t, int64(3000), rec.timeout)
	rec.mu.Unlock()

	output, err := ctrl.GetShellOutput()
	require.NoError(t, err)
	require.Equal(t, "shell-output-42", output)
}

// touchRecorder is a custom controller that routes clicks and swipes through
// touch events (UseMouseDownAndUpInsteadOfClick) and records every touch
// call, so contact and pressure forwarding can be asserted.
type touchRecorder struct {
	BlankController
	mu    sync.Mutex
	downs []touchCall
	moves []touchCall
	ups   []int32
}

type touchCall struct {
	contact  int32
	x, y     int32
	pressure int32
}

func (r *touchRecorder) GetFeature() ControllerFeature {
	return ControllerFeatureUseMouseDownAndUpInsteadOfClick
}

func (r *touchRecorder) TouchDown(contact int32, x int32, y int32, pressure int32) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.downs = append(r.downs, touchCall{contact: contact, x: x, y: y, pressure: pressure})
	return true
}

func (r *touchRecorder) TouchMove(contact int32, x int32, y int32, pressure int32) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.moves = append(r.moves, touchCall{contact: contact, x: x, y: y, pressure: pressure})
	return true
}

func (r *touchRecorder) TouchUp(contact int32) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ups = append(r.ups, contact)
	return true
}

// TestController_PostClickV2_PostSwipeV2_ForwardContactAndPressure pins that
// on a controller using touch-down/up for clicks and swipes, the V2 post APIs
// forward contact and pressure through the native pipeline to TouchDown,
// TouchMove, and TouchUp.
func TestController_PostClickV2_PostSwipeV2_ForwardContactAndPressure(t *testing.T) {
	rec := &touchRecorder{}
	ctrl, err := NewCustomController(rec)
	require.NoError(t, err)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	require.True(t, connectJob.Wait().Success())

	clickJob, err := ctrl.PostClickV2(10, 20, 2, 500)
	require.NoError(t, err)
	require.True(t, clickJob.Wait().Success())

	rec.mu.Lock()
	require.Equal(t, []touchCall{{contact: 2, x: 10, y: 20, pressure: 500}}, rec.downs)
	require.Equal(t, []int32{2}, rec.ups)
	rec.mu.Unlock()

	swipeJob, err := ctrl.PostSwipeV2(1, 2, 3, 4, 50*time.Millisecond, 3, 250)
	require.NoError(t, err)
	require.True(t, swipeJob.Wait().Success())

	rec.mu.Lock()
	require.Equal(t, touchCall{contact: 3, x: 1, y: 2, pressure: 250}, rec.downs[1])
	require.NotEmpty(t, rec.moves)
	require.Equal(t, touchCall{contact: 3, x: 3, y: 4, pressure: 250}, rec.moves[len(rec.moves)-1])
	require.Equal(t, []int32{2, 3}, rec.ups)
	rec.mu.Unlock()
}

// TestController_GetResolution_RawVersusScaled pins that GetResolution
// reports the raw device resolution while CacheImage is scaled to the
// configured screenshot target.
func TestController_GetResolution_RawVersusScaled(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	require.True(t, connectJob.Wait().Success())
	screencapJob, err := ctrl.PostScreencap()
	require.NoError(t, err)
	require.True(t, screencapJob.Wait().Success())

	width, height, err := ctrl.GetResolution()
	require.NoError(t, err)
	require.Equal(t, int32(1280), width)
	require.Equal(t, int32(720), height)

	require.NoError(t, ctrl.SetScreenshot(WithScreenshotTargetShortSide(360)))
	scaledJob, err := ctrl.PostScreencap()
	require.NoError(t, err)
	require.True(t, scaledJob.Wait().Success())

	img, err := ctrl.CacheImage()
	require.NoError(t, err)
	require.Equal(t, image.Rect(0, 0, 640, 360), img.Bounds())

	width, height, err = ctrl.GetResolution()
	require.NoError(t, err)
	require.Equal(t, int32(1280), width)
	require.Equal(t, int32(720), height)
}
