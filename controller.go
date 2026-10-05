package maa

import (
	"errors"
	"fmt"
	"image"
	"time"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v4/controller/adb"
	"github.com/MaaXYZ/maa-framework-go/v4/controller/macos"
	"github.com/MaaXYZ/maa-framework-go/v4/controller/win32"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/store"
)

func initControllerStore(handle uintptr) {
	store.CtrlStore.Lock()
	store.CtrlStore.Set(handle, store.CtrlStoreValue{
		SinkIDToEventCallbackID:     make(map[int64]uint64),
		CustomControllerCallbacksID: 0,
	})
	store.CtrlStore.Unlock()
}

// Controller is an owned or borrowed native controller. Handle lifetime is
// guarded across calls; callers must coordinate configuration and execution.
type Controller struct {
	handle uintptr
	state  *handleState
	owned  bool
}

func newOwnedController(handle uintptr) *Controller {
	var state *handleState
	state = newHandleState(handle, func(handle uintptr) {
		store.CtrlStore.Lock()
		value := store.CtrlStore.Get(handle)
		for _, cbID := range value.SinkIDToEventCallbackID {
			unregisterEventCallback(cbID)
		}
		store.CtrlStore.Del(handle)
		store.CtrlStore.Unlock()
		native.MaaControllerDestroy(handle)
		unregisterCustomControllerCallbacks(value.CustomControllerCallbacksID)
		controllerStates.CompareAndDelete(handle, state)
	})
	state.jobStatus = func(handle uintptr, id int64) Status {
		return Status(native.MaaControllerStatus(handle, id))
	}
	state.idleProbe = func(handle uintptr) bool {
		// Stop can invalidate action IDs before their worker returns. Observe a
		// new inactive action to prove completion without waiting under mu.
		if state.closeBarrierID == 0 {
			state.closeBarrierID = native.MaaControllerPostInactive(handle)
			return false
		}
		status := Status(native.MaaControllerStatus(handle, state.closeBarrierID))
		if status.Invalid() {
			state.closeBarrierID = 0
		}
		return status.Done()
	}
	controllerStates.Store(handle, state)
	return &Controller{handle: handle, state: state, owned: true}
}

// NewAdbController creates a new ADB controller.
// config must be a valid JSON object; use "{}" for defaults. It is normally
// obtained from AdbDevice.Config (see FindAdbDevices). Its "command.<Name>"
// arrays override the built-in adb command templates.
// agentPath is the directory holding the agent binaries; the minitouch and
// maatouch input methods require their binaries to exist there.
func NewAdbController(
	adbPath, address string,
	screencapMethod adb.ScreencapMethod,
	inputMethod adb.InputMethod,
	config, agentPath string,
) (*Controller, error) {
	handle := native.MaaAdbControllerCreate(
		adbPath,
		address,
		uint64(screencapMethod),
		uint64(inputMethod),
		config,
		agentPath,
	)
	if handle == 0 {
		return nil, errors.New("failed to create ADB controller")
	}

	initControllerStore(handle)

	return newOwnedController(handle), nil
}

// NewPlayCoverController creates a new PlayCover controller.
// address is the PlayTools service address in "host:port" format.
// uuid is the bundle identifier of the target application
// (e.g. "com.hypergryph.arknights").
// start_app, input_text, click_key, key_down, key_up, and scroll are not supported.
func NewPlayCoverController(
	address, uuid string,
) (*Controller, error) {
	handle := native.MaaPlayCoverControllerCreate(address, uuid)
	if handle == 0 {
		return nil, errors.New("failed to create PlayCover controller")
	}

	initControllerStore(handle)

	return newOwnedController(handle), nil
}

// NewWin32Controller creates a win32 controller instance.
func NewWin32Controller(
	hWnd unsafe.Pointer,
	screencapMethod win32.ScreencapMethod,
	mouseMethod win32.InputMethod,
	keyboardMethod win32.InputMethod,
) (*Controller, error) {
	handle := native.MaaWin32ControllerCreate(
		hWnd,
		uint64(screencapMethod),
		uint64(mouseMethod),
		uint64(keyboardMethod),
	)
	if handle == 0 {
		return nil, errors.New("failed to create Win32 controller")
	}

	initControllerStore(handle)

	return newOwnedController(handle), nil
}

// NewLinuxController creates a Linux controller from a JSON configuration.
// configJson must specify screencap_method and input_method, plus the fields
// required by the selected methods (such as wlr_socket_path for wlroots,
// pw_node_id for PipeWire, or eis_socket_path for libei). The PipeWire portal
// screencap method requires both pw_socket_fd and pw_node_id. See the
// MaaFramework MaaLinuxControllerCreate documentation for the complete
// configuration format.
// This controller is only available on Linux.
func NewLinuxController(configJson string) (*Controller, error) {
	handle := native.MaaLinuxControllerCreate(configJson)
	if handle == 0 {
		return nil, errors.New("failed to create Linux controller")
	}

	initControllerStore(handle)

	return newOwnedController(handle), nil
}

// NewMacOSController creates a macOS controller for native macOS applications.
// windowID is the CGWindowID of the target window (0 for desktop).
// screencapMethod is the macOS screencap method to use.
// inputMethod is the macOS input method to use.
//
// Note: Requires Screen Recording permission for screencap.
// Input simulation requires Accessibility permission.
// Some features are not supported: start_app, stop_app, scroll.
// Only single touch is supported (contact must be 0).
func NewMacOSController(
	windowID uint32,
	screencapMethod macos.ScreencapMethod,
	inputMethod macos.InputMethod,
) (*Controller, error) {
	handle := native.MaaMacOSControllerCreate(
		windowID,
		native.MaaMacOSScreencapMethod(screencapMethod),
		native.MaaMacOSInputMethod(inputMethod),
	)
	if handle == 0 {
		return nil, errors.New("failed to create macOS controller")
	}

	initControllerStore(handle)

	return newOwnedController(handle), nil
}

// NewAndroidNativeController creates an Android native controller backed by MaaAndroidNativeControlUnit.
// configJson is the JSON config for the control unit. Required fields:
//   - library_path: path to the Android native control unit library
//   - screen_resolution.width / screen_resolution.height: raw screenshot and touch resolution
//
// Optional fields:
//   - display_id: target display id, defaults to 0
//   - force_stop: whether to force stop before start_app, defaults to false
//
// Note: This controller is only available on Android.
// The configured screen_resolution must match the control unit's raw screenshot/touch coordinate space.
// Multi-touch is supported: contact is the finger id (0 for the first finger).
// The external library's TouchArgs must carry contact.
func NewAndroidNativeController(configJson string) (*Controller, error) {
	handle := native.MaaAndroidNativeControllerCreate(configJson)
	if handle == 0 {
		return nil, errors.New("failed to create Android native controller")
	}

	initControllerStore(handle)

	return newOwnedController(handle), nil
}

// NewReplayController creates a replay controller that replays recorded operations.
// recordingPath is the path to the recording JSONL file written by a record controller.
// Screenshot image paths in the file are resolved relative to this file's parent directory.
func NewReplayController(recordingPath string) (*Controller, error) {
	handle := native.MaaReplayControllerCreate(recordingPath)
	if handle == 0 {
		return nil, errors.New("failed to create replay controller")
	}

	initControllerStore(handle)

	return newOwnedController(handle), nil
}

// NewRecordController creates a record controller that wraps an existing controller and records all operations.
// inner is the inner controller to forward all operations to; it must not be nil.
// The record controller does NOT take ownership of the inner controller.
// recordingPath is the path to the recording JSONL file to write.
// Screenshot images will be saved to a "{stem}-Screenshot" folder in the same directory as this file.
// The recorded file can be replayed using NewReplayController.
func NewRecordController(inner *Controller, recordingPath string) (*Controller, error) {
	if inner == nil {
		return nil, errors.New("inner controller is nil")
	}
	innerHandle, done, err := inner.state.begin()
	if err != nil {
		return nil, err
	}
	defer done()
	handle := native.MaaRecordControllerCreate(innerHandle, recordingPath)
	if handle == 0 {
		return nil, errors.New("failed to create record controller")
	}

	initControllerStore(handle)

	return newOwnedController(handle), nil
}

// GamepadType defines the type of virtual gamepad.
type GamepadType = native.MaaGamepadType

// Gamepad type constants.
const (
	GamepadTypeXbox360    GamepadType = native.MaaGamepadType_Xbox360
	GamepadTypeDualShock4 GamepadType = native.MaaGamepadType_DualShock4
)

// NewGamepadController creates a virtual gamepad controller for Windows.
//
// hWnd: Window handle for screencap (optional, can be nil if screencap not needed).
// gamepadType: Type of virtual gamepad (Xbox360 or DualShock4).
// screencapMethod: Win32 screencap method to use. Ignored if hWnd is nil.
//
// Note: Requires ViGEm Bus Driver to be installed on the system.
// click and swipe are not directly supported; input_text, start_app,
// stop_app, and scroll are not supported.
// For gamepad button and touch constants, import "github.com/MaaXYZ/maa-framework-go/v4/controller/gamepad".
func NewGamepadController(
	hWnd unsafe.Pointer,
	gamepadType GamepadType,
	screencapMethod win32.ScreencapMethod,
) (*Controller, error) {
	handle := native.MaaGamepadControllerCreate(hWnd, gamepadType, uint64(screencapMethod))
	if handle == 0 {
		return nil, errors.New("failed to create Gamepad controller")
	}

	initControllerStore(handle)

	return newOwnedController(handle), nil
}

// NewCustomController creates a custom controller instance.
func NewCustomController(
	ctrl CustomController,
) (*Controller, error) {
	if ctrl == nil {
		return nil, errors.New("custom controller is nil")
	}

	ctrlID := registerCustomControllerCallbacks(ctrl)
	handle := native.MaaCustomControllerCreate(
		unsafe.Pointer(customControllerCallbacksHandle),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		uintptr(ctrlID),
	)
	if handle == 0 {
		unregisterCustomControllerCallbacks(ctrlID)
		return nil, errors.New("failed to create Custom controller")
	}

	initControllerStore(handle)

	store.CtrlStore.Update(handle, func(v *store.CtrlStoreValue) {
		v.CustomControllerCallbacksID = ctrlID
	})

	owner := newOwnedController(handle)
	bindCustomControllerCallbacks(ctrlID, owner.state)
	return owner, nil
}

// NOTE: MaaDbgController (MaaDbgControllerCreate) is intentionally NOT implemented in the Go binding.
// MaaDbgControllerCreate remains a current, non-deprecated API upstream; the
// Go binding offers these alternatives instead:
//   - BlankController (blank_controller.go): no-op stub that always succeeds
//   - NewReplayController: replay recorded operations from a JSONL file
// Do NOT add a Go binding for MaaDbgControllerCreate or NewDbgController here.
// The api-check CI tool also blacklists MaaDbgControllerCreate for the same reason.

// Destroy closes the controller once. It returns ErrBound while a tasker uses
// it or an AgentClient retains it, ErrInUse while a call or job is active,
// ErrBorrowed when called on a getter or callback view, and ErrInCallback when
// called from inside one of the controller's own callbacks.
// After a stop invalidates action IDs, Destroy may post an inactive action and
// return ErrInUse until it completes; retry afterward. Native destruction may
// call custom KeyUp/TouchUp methods before Destroy returns successfully.
func (c *Controller) Destroy() error {
	if c == nil || !c.owned {
		return ErrBorrowed
	}
	return c.state.close()
}

// setOption sets options for controller instance.
func (c *Controller) setOption(key native.MaaCtrlOption, value unsafe.Pointer, valSize uintptr) error {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return useErr
	}
	defer done()

	if native.MaaControllerSetOption(c.handle, key, value, uint64(valSize)) {
		return nil
	}
	return fmt.Errorf("failed to set controller option: %v", key)
}

type screenshotOptionKind int

const (
	screenshotOptionUnset screenshotOptionKind = iota
	screenshotOptionLongSide
	screenshotOptionShortSide
	screenshotOptionRawSize
	screenshotOptionResizeMethod
	screenshotOptionExpand
)

type screenshotOptionConfig struct {
	kind            screenshotOptionKind
	targetLongSide  int32
	targetShortSide int32
	useRawSize      bool
	resizeMethod    int32
	targetExpand    [2]int32
}

// String returns a human-readable name for the screenshot option kind.
func (kind screenshotOptionKind) String() string {
	switch kind {
	case screenshotOptionLongSide:
		return "target long side"
	case screenshotOptionShortSide:
		return "target short side"
	case screenshotOptionRawSize:
		return "use raw size"
	case screenshotOptionResizeMethod:
		return "resize method"
	case screenshotOptionExpand:
		return "target expand"
	default:
		return fmt.Sprintf("unknown (%d)", kind)
	}
}

func (cfg screenshotOptionConfig) validate() error {
	switch cfg.kind {
	case screenshotOptionUnset, screenshotOptionRawSize:
		return nil
	case screenshotOptionLongSide:
		if cfg.targetLongSide <= 0 {
			return fmt.Errorf("invalid screenshot target long side: %d (must be positive)", cfg.targetLongSide)
		}
	case screenshotOptionShortSide:
		if cfg.targetShortSide <= 0 {
			return fmt.Errorf("invalid screenshot target short side: %d (must be positive)", cfg.targetShortSide)
		}
	case screenshotOptionExpand:
		if cfg.targetExpand[0] <= 0 || cfg.targetExpand[1] <= 0 {
			return fmt.Errorf("invalid screenshot target expand: %dx%d (both dimensions must be positive)", cfg.targetExpand[0], cfg.targetExpand[1])
		}
	case screenshotOptionResizeMethod:
		if cfg.resizeMethod < int32(ScreenshotResizeMethodNearestNeighbor) || cfg.resizeMethod > int32(ScreenshotResizeMethodLanczos4) {
			return fmt.Errorf("invalid screenshot resize method: %d (must be between 0 and 4)", cfg.resizeMethod)
		}
	default:
		return fmt.Errorf("unknown screenshot option kind: %d", cfg.kind)
	}
	return nil
}

// ScreenshotResizeMethod is the interpolation method used when resizing screenshots.
// Values correspond to cv::InterpolationFlags.
type ScreenshotResizeMethod int32

const (
	// ScreenshotResizeMethodNearestNeighbor selects nearest-neighbor interpolation (cv::INTER_NEAREST).
	ScreenshotResizeMethodNearestNeighbor ScreenshotResizeMethod = 0
	// ScreenshotResizeMethodLinear selects bilinear interpolation (cv::INTER_LINEAR).
	ScreenshotResizeMethodLinear ScreenshotResizeMethod = 1
	// ScreenshotResizeMethodCubic selects bicubic interpolation (cv::INTER_CUBIC).
	ScreenshotResizeMethodCubic ScreenshotResizeMethod = 2
	// ScreenshotResizeMethodArea selects pixel-area interpolation (cv::INTER_AREA).
	// This is the default method.
	ScreenshotResizeMethodArea ScreenshotResizeMethod = 3
	// ScreenshotResizeMethodLanczos4 selects 8x8 Lanczos interpolation (cv::INTER_LANCZOS4).
	ScreenshotResizeMethodLanczos4 ScreenshotResizeMethod = 4
)

// ScreenshotOption configures how the screenshot is resized.
// Options for different settings can be combined; see Controller.SetScreenshot
// for validation and conflict rules.
type ScreenshotOption func(*screenshotOptionConfig)

// WithScreenshotTargetLongSide sets screenshot target long side.
// The short side is scaled proportionally. Setting this replaces short-side and expand targets.
// The target must be positive. It does not disable raw-size mode.
//
// eg: 1280
func WithScreenshotTargetLongSide(targetLongSide int32) ScreenshotOption {
	return func(cfg *screenshotOptionConfig) {
		cfg.kind = screenshotOptionLongSide
		cfg.targetLongSide = targetLongSide
	}
}

// WithScreenshotTargetShortSide sets screenshot target short side.
// The long side is scaled proportionally. Setting this replaces long-side and expand targets.
// The target must be positive. It does not disable raw-size mode.
//
// eg: 720
func WithScreenshotTargetShortSide(targetShortSide int32) ScreenshotOption {
	return func(cfg *screenshotOptionConfig) {
		cfg.kind = screenshotOptionShortSide
		cfg.targetShortSide = targetShortSide
	}
}

// WithScreenshotTargetExpand scales screenshots uniformly to cover the reference width and height.
// The scale is max(width/rawWidth, height/rawHeight), preserving the source aspect ratio
// without cropping or stretching. Both output dimensions are at least the reference size.
// Width and height must be positive. Setting this replaces long-side and short-side targets.
// It does not disable raw-size mode; the target is retained for later use while
// raw-size mode is active.
func WithScreenshotTargetExpand(width, height int32) ScreenshotOption {
	return func(cfg *screenshotOptionConfig) {
		cfg.kind = screenshotOptionExpand
		cfg.targetExpand = [2]int32{width, height}
	}
}

// WithScreenshotUseRawSize sets whether the screenshot uses the raw size without scaling.
// Enabling raw-size mode retains the target and interpolation method; disabling
// it resumes scaling with those settings. Combine false with a target option to
// set a new target and resume scaling in one call.
func WithScreenshotUseRawSize(enabled bool) ScreenshotOption {
	return func(cfg *screenshotOptionConfig) {
		cfg.kind = screenshotOptionRawSize
		cfg.useRawSize = enabled
	}
}

// WithScreenshotResizeMethod sets the interpolation method used when resizing screenshots.
// Defaults to ScreenshotResizeMethodArea (cv::INTER_AREA).
// Only values from ScreenshotResizeMethodNearestNeighbor through
// ScreenshotResizeMethodLanczos4 are supported (0 through 4).
func WithScreenshotResizeMethod(method ScreenshotResizeMethod) ScreenshotOption {
	return func(cfg *screenshotOptionConfig) {
		cfg.kind = screenshotOptionResizeMethod
		cfg.resizeMethod = int32(method)
	}
}

// SetScreenshot updates the supplied screenshot settings on the controller.
// A target, interpolation method, and WithScreenshotUseRawSize(false) can be
// combined in any order. Long-side, short-side, and expand targets are mutually
// exclusive within one call, as are a target and WithScreenshotUseRawSize(true).
// Target dimensions must be positive, and the interpolation method must be one
// of the five ScreenshotResizeMethod constants (0 through 4). The native API
// treats a target value of 0 as clearing the target; this Go wrapper
// deliberately rejects non-positive values instead.
// Repeated settings are rejected even when their values are identical. Nil
// options are ignored; an empty call leaves the settings unchanged.
//
// Omitted settings are retained. Across separate calls, a new target replaces
// the previous target without disabling raw-size mode. Disabling raw-size mode
// resumes resizing with the retained target and interpolation method. If no
// target has ever been set, the native default target scales the short side
// to 720.
//
// Rejecting raw-size true with a target is a Go API guard against ambiguous
// screenshot-size intent: the target would be inactive in raw-size mode. The
// native API can retain a target while raw-size mode is enabled; set it in a
// separate call to prepare for later resizing, or combine it with
// WithScreenshotUseRawSize(false) to resume resizing with it. Interpolation is
// an independent preference and can accompany either raw-size value.
//
// All options are validated before any native setting is changed. Valid options
// are applied in the order target, interpolation method, then raw-size mode.
// If a native setter fails, earlier changes remain applied and later setters are
// skipped. Callers must serialize configuration with other controller operations.
//
// For example, to resume resizing with a 1280-pixel long side and linear interpolation:
//
//	if err := ctrl.SetScreenshot(
//		WithScreenshotUseRawSize(false),
//		WithScreenshotTargetLongSide(1280),
//		WithScreenshotResizeMethod(ScreenshotResizeMethodLinear),
//	); err != nil {
//		return err
//	}
func (c *Controller) SetScreenshot(opts ...ScreenshotOption) error {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return useErr
	}
	defer done()

	var settings [screenshotOptionExpand + 1]screenshotOptionConfig
	targetKind := screenshotOptionUnset
	for i, opt := range opts {
		if opt == nil {
			continue
		}
		var cfg screenshotOptionConfig
		opt(&cfg)
		if err := cfg.validate(); err != nil {
			return fmt.Errorf("screenshot option %d: %w", i+1, err)
		}
		if cfg.kind == screenshotOptionUnset {
			continue
		}
		if settings[cfg.kind].kind != screenshotOptionUnset {
			return fmt.Errorf("duplicate screenshot option: %s", cfg.kind)
		}
		switch cfg.kind {
		case screenshotOptionLongSide, screenshotOptionShortSide, screenshotOptionExpand:
			if targetKind != screenshotOptionUnset {
				return fmt.Errorf("conflicting screenshot targets: %s and %s", targetKind, cfg.kind)
			}
			targetKind = cfg.kind
		}
		settings[cfg.kind] = cfg
	}
	if settings[screenshotOptionRawSize].useRawSize && targetKind != screenshotOptionUnset {
		return fmt.Errorf("conflicting screenshot options: use raw size true and %s", targetKind)
	}

	for _, kind := range [...]screenshotOptionKind{
		screenshotOptionLongSide,
		screenshotOptionShortSide,
		screenshotOptionExpand,
		screenshotOptionResizeMethod,
		screenshotOptionRawSize,
	} {
		cfg := settings[kind]
		if cfg.kind == screenshotOptionUnset {
			continue
		}
		if err := c.setScreenshotOption(cfg); err != nil {
			return fmt.Errorf("failed to set screenshot %s: %w", kind, err)
		}
	}
	return nil
}

func (c *Controller) setScreenshotOption(cfg screenshotOptionConfig) error {
	switch cfg.kind {
	case screenshotOptionLongSide:
		return c.setOption(
			native.MaaCtrlOption_ScreenshotTargetLongSide,
			unsafe.Pointer(&cfg.targetLongSide),
			unsafe.Sizeof(cfg.targetLongSide),
		)
	case screenshotOptionShortSide:
		return c.setOption(
			native.MaaCtrlOption_ScreenshotTargetShortSide,
			unsafe.Pointer(&cfg.targetShortSide),
			unsafe.Sizeof(cfg.targetShortSide),
		)
	case screenshotOptionRawSize:
		return c.setOption(
			native.MaaCtrlOption_ScreenshotUseRawSize,
			unsafe.Pointer(&cfg.useRawSize),
			unsafe.Sizeof(cfg.useRawSize),
		)
	case screenshotOptionResizeMethod:
		return c.setOption(
			native.MaaCtrlOption_ScreenshotResizeMethod,
			unsafe.Pointer(&cfg.resizeMethod),
			unsafe.Sizeof(cfg.resizeMethod),
		)
	case screenshotOptionExpand:
		return c.setOption(
			native.MaaCtrlOption_ScreenshotTargetExpand,
			unsafe.Pointer(&cfg.targetExpand[0]),
			unsafe.Sizeof(cfg.targetExpand),
		)
	default:
		return fmt.Errorf("unknown screenshot option kind: %v", cfg.kind)
	}
}

// SetMouseLockFollow enables or disables mouse-lock-follow mode.
// This is designed for TPS/FPS games that lock the mouse to their window in the background.
// Only valid for Win32 controllers using message-based input methods.
func (c *Controller) SetMouseLockFollow(enabled bool) error {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return useErr
	}
	defer done()

	return c.setOption(
		native.MaaCtrlOption_MouseLockFollow,
		unsafe.Pointer(&enabled),
		unsafe.Sizeof(enabled),
	)
}

// SetBackgroundManagedKeys sets the Win32 virtual-key codes managed during
// background input. Call it before connecting a Win32 controller. An empty
// slice clears the managed key list.
func (c *Controller) SetBackgroundManagedKeys(keys []int32) error {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return useErr
	}
	defer done()

	value := unsafe.Pointer(unsafe.SliceData(keys))
	var empty int32
	if len(keys) == 0 {
		// Native code constructs a pointer range even for an empty list.
		value = unsafe.Pointer(&empty)
	}
	return c.setOption(
		native.MaaCtrlOption_BackgroundManagedKeys,
		value,
		uintptr(len(keys))*unsafe.Sizeof(int32(0)),
	)
}

// PostConnect posts a connection.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostConnect() (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostConnection(c.handle)
	if id == 0 {
		return failJob(errors.New("failed to post connect"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostClick posts a click.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostClick(x, y int32) (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostClick(c.handle, x, y)
	if id == 0 {
		return failJob(errors.New("failed to post click"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostClickV2 posts a click with contact and pressure.
// For adb controller, contact means finger id (0 for first finger, 1 for second finger, etc).
// For win32 controller, contact means mouse button id (0 for left, 1 for right, 2 for middle).
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostClickV2(x, y, contact, pressure int32) (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostClickV2(c.handle, x, y, contact, pressure)
	if id == 0 {
		return failJob(errors.New("failed to post click v2"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostSwipe posts a swipe.
// The duration is converted to int32 milliseconds for the native API, so
// durations beyond the int32 millisecond range (about 24.8 days) wrap around,
// matching the C API's own limit.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostSwipe(x1, y1, x2, y2 int32, duration time.Duration) (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostSwipe(c.handle, x1, y1, x2, y2, int32(duration.Milliseconds()))
	if id == 0 {
		return failJob(errors.New("failed to post swipe"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostSwipeV2 posts a swipe with contact and pressure.
// For adb controller, contact means finger id (0 for first finger, 1 for second finger, etc).
// For win32 controller, contact means mouse button id (0 for left, 1 for right, 2 for middle).
// The duration is converted to int32 milliseconds for the native API, so
// durations beyond the int32 millisecond range (about 24.8 days) wrap around,
// matching the C API's own limit.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostSwipeV2(x1, y1, x2, y2 int32, duration time.Duration, contact, pressure int32) (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostSwipeV2(c.handle, x1, y1, x2, y2, int32(duration.Milliseconds()), contact, pressure)
	if id == 0 {
		return failJob(errors.New("failed to post swipe v2"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostClickKey posts a click key.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostClickKey(keycode int32) (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostClickKey(c.handle, keycode)
	if id == 0 {
		return failJob(errors.New("failed to post click key"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostInputText posts an input text.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostInputText(text string) (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostInputText(c.handle, text)
	if id == 0 {
		return failJob(errors.New("failed to post input text"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostStartApp posts a start app.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostStartApp(intent string) (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostStartApp(c.handle, intent)
	if id == 0 {
		return failJob(errors.New("failed to post start app"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostStopApp posts a stop app.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostStopApp(intent string) (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostStopApp(c.handle, intent)
	if id == 0 {
		return failJob(errors.New("failed to post stop app"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostTouchDown posts a touch-down.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostTouchDown(contact, x, y, pressure int32) (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostTouchDown(c.handle, contact, x, y, pressure)
	if id == 0 {
		return failJob(errors.New("failed to post touch down"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostTouchMove posts a touch-move.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostTouchMove(contact, x, y, pressure int32) (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostTouchMove(c.handle, contact, x, y, pressure)
	if id == 0 {
		return failJob(errors.New("failed to post touch move"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostTouchUp posts a touch-up.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostTouchUp(contact int32) (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostTouchUp(c.handle, contact)
	if id == 0 {
		return failJob(errors.New("failed to post touch up"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostRelativeMove posts a relative cursor move.
// dx and dy are the horizontal and vertical move offsets.
// Supported by Win32 controllers whose input method implements relative move
// (the message-based input used by mouse-lock-follow), by Linux controllers
// (uinput/wlroots/libei input), and by custom controllers implementing
// RelativeMove; macOS controllers do not support it.
// If the controller does not support relative move, the posted action will fail.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostRelativeMove(dx, dy int32) (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostRelativeMove(c.handle, dx, dy)
	if id == 0 {
		return failJob(errors.New("failed to post relative move"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostKeyDown posts a key-down.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostKeyDown(keycode int32) (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostKeyDown(c.handle, keycode)
	if id == 0 {
		return failJob(errors.New("failed to post key down"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostKeyUp posts a key-up.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostKeyUp(keycode int32) (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostKeyUp(c.handle, keycode)
	if id == 0 {
		return failJob(errors.New("failed to post key up"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostScreencap posts a screencap.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostScreencap() (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostScreencap(c.handle)
	if id == 0 {
		return failJob(errors.New("failed to post screencap"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostScroll posts a scroll.
// dx and dy are the horizontal and vertical scroll deltas: positive dx scrolls
// right, positive dy scrolls up.
// Scroll is supported by Win32 controllers and by custom controllers that
// implement the Scroll method; for other controller types the action fails.
// On Win32 the amounts are wheel-delta units, so multiples of 120 are recommended.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostScroll(dx, dy int32) (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostScroll(c.handle, dx, dy)
	if id == 0 {
		return failJob(errors.New("failed to post scroll"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostInactive posts an inactive request to restore controller/window state.
// For Win32 controllers this restores window position (removes topmost) and unblocks user input.
// For other controllers this is a no-op that typically succeeds.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostInactive() (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostInactive(c.handle)
	if id == 0 {
		return failJob(errors.New("failed to post inactive"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// PostShell posts a shell command.
// Supported by ADB controllers and by custom controllers that implement the
// Shell method; for other controller types the action fails.
// The timeout is passed to the native API in milliseconds: a negative timeout
// waits indefinitely, and zero returns immediately. No default applies, as the
// Go wrapper always passes an explicit value.
// It returns an error and a terminal-failed job when the request cannot
// be submitted, for example when the underlying object is closed.
func (c *Controller) PostShell(cmd string, timeout time.Duration) (*Job, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return failJob(useErr)
	}
	defer done()

	id := native.MaaControllerPostShell(c.handle, cmd, timeout.Milliseconds())
	if id == 0 {
		return failJob(errors.New("failed to post shell"))
	}
	return newJob(id, c.status, c.wait, c.state), nil
}

// GetShellOutput gets the output of the last shell command.
func (c *Controller) GetShellOutput() (string, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return "", useErr
	}
	defer done()

	output := buffer.NewStringBuffer()
	defer output.Destroy()

	got := native.MaaControllerGetShellOutput(c.handle, output.Handle())
	if !got {
		return "", errors.New("failed to get shell output")
	}
	return output.Get(), nil
}

// status gets the status of a request identified by the given id.
func (c *Controller) status(id int64) Status {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return StatusInvalid
	}
	defer done()

	return Status(native.MaaControllerStatus(c.handle, id))
}

func (c *Controller) wait(id int64) Status {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return StatusInvalid
	}
	defer done()

	return Status(native.MaaControllerWait(c.handle, id))
}

// Connected checks if the controller is connected.
func (c *Controller) Connected() bool {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return false
	}
	defer done()

	return native.MaaControllerConnected(c.handle)
}

// CacheImage gets the image buffer of the last screencap request.
// The cached image is scaled to the configured screenshot target size, so its
// dimensions may differ from the raw device resolution (see GetResolution).
func (c *Controller) CacheImage() (image.Image, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return nil, useErr
	}
	defer done()

	img, err := c.CacheImageInto(nil)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// CacheImageInto gets the image buffer of the last screencap request and writes into dst when possible.
// If dst is nil or size mismatched, a new *image.RGBA is allocated and returned.
func (c *Controller) CacheImageInto(dst *image.RGBA) (*image.RGBA, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return nil, useErr
	}
	defer done()

	imgBuffer, err := buffer.NewImageBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to create image buffer: %w", err)
	}
	defer imgBuffer.Destroy()

	got := native.MaaControllerCachedImage(c.handle, imgBuffer.Handle())
	if !got {
		return nil, errors.New("failed to get cached image")
	}

	img := imgBuffer.GetInto(dst)

	return img, nil
}

// GetUUID gets the UUID of the controller.
func (c *Controller) GetUUID() (string, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return "", useErr
	}
	defer done()

	uuid := buffer.NewStringBuffer()
	defer uuid.Destroy()
	got := native.MaaControllerGetUuid(c.handle, uuid.Handle())
	if !got {
		return "", errors.New("failed to get UUID")
	}
	return uuid.Get(), nil
}

// GetResolution gets the raw (unscaled) device resolution.
// Returns the width and height. Returns an error if the resolution is not
// available yet, for example before connecting or before the first screencap.
// Note: This returns the actual device screen resolution before any scaling.
// The screenshot obtained via CacheImage is scaled according to the screenshot target size settings.
func (c *Controller) GetResolution() (width, height int32, err error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return 0, 0, useErr
	}
	defer done()

	got := native.MaaControllerGetResolution(c.handle, &width, &height)
	if !got {
		return 0, 0, fmt.Errorf("failed to get resolution")
	}
	return width, height, nil
}

// GetInfo gets controller information as a JSON string.
// Returns controller-specific information including type, constructor parameters and current state.
// The returned JSON always contains a "type" key identifying the controller
// kind; for custom controllers it is always "custom".
func (c *Controller) GetInfo() (string, error) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return "", useErr
	}
	defer done()

	buf := buffer.NewStringBuffer()
	defer buf.Destroy()
	got := native.MaaControllerGetInfo(c.handle, buf.Handle())
	if !got {
		return "", errors.New("failed to get controller info")
	}
	return buf.Get(), nil
}

// AddSink adds an event callback sink and returns the sink ID.
// The sink ID can be used to remove the sink later.
// The instance and associated taskers must be idle. Do not call this from a callback.
// It returns 0 if registration fails or the object is closed.
func (c *Controller) AddSink(sink ControllerEventSink) int64 {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return 0
	}
	defer done()
	if c.state.external {
		return 0
	}
	c.state.registrationMu.Lock()
	defer c.state.registrationMu.Unlock()

	id := registerEventCallback(sink, c.state)
	sinkId := native.MaaControllerAddSink(
		c.handle,
		_MaaEventCallbackAgent,
		uintptr(id),
	)

	if sinkId == 0 {
		unregisterEventCallback(id)
		return 0
	}

	store.CtrlStore.Update(c.handle, func(v *store.CtrlStoreValue) {
		v.SinkIDToEventCallbackID[sinkId] = id
	})

	return sinkId
}

// RemoveSink removes an event callback sink by sink ID.
// The instance and associated taskers must be idle. Do not call this from a callback.
func (c *Controller) RemoveSink(sinkId int64) {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return
	}
	defer done()
	if c.state.external {
		return
	}
	c.state.registrationMu.Lock()
	defer c.state.registrationMu.Unlock()

	store.CtrlStore.Update(c.handle, func(v *store.CtrlStoreValue) {
		unregisterEventCallback(v.SinkIDToEventCallbackID[sinkId])
		delete(v.SinkIDToEventCallbackID, sinkId)
	})

	native.MaaControllerRemoveSink(c.handle, sinkId)
}

// ClearSinks clears all event callback sinks.
// The instance and associated taskers must be idle. Do not call this from a callback.
func (c *Controller) ClearSinks() {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return
	}
	defer done()
	if c.state.external {
		return
	}
	c.state.registrationMu.Lock()
	defer c.state.registrationMu.Unlock()

	store.CtrlStore.Update(c.handle, func(v *store.CtrlStoreValue) {
		for _, id := range v.SinkIDToEventCallbackID {
			unregisterEventCallback(id)
		}
		v.SinkIDToEventCallbackID = make(map[int64]uint64)
	})

	native.MaaControllerClearSinks(c.handle)
}

// ControllerEventSink is the sink for controller action events
// (Controller.Action.* messages), delivered with a borrowed *Controller view
// of the controller that posted the event.
// See Controller.OnControllerAction for a single-event convenience wrapper.
type ControllerEventSink interface {
	OnControllerAction(ctrl *Controller, event EventStatus, detail ControllerActionDetail)
}

// ctrlEventSinkAdapter is a lightweight adapter that makes it easy to register
// a single-event handler via a callback function.
type ctrlEventSinkAdapter struct {
	onControllerAction func(EventStatus, ControllerActionDetail)
}

// OnControllerAction forwards the event to the wrapped handler function.
func (a *ctrlEventSinkAdapter) OnControllerAction(
	ctrl *Controller,
	status EventStatus,
	detail ControllerActionDetail,
) {
	if a == nil || a.onControllerAction == nil {
		return
	}
	a.onControllerAction(status, detail)
}

// OnControllerAction registers a callback sink that only handles Controller.Action events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (c *Controller) OnControllerAction(
	fn func(EventStatus, ControllerActionDetail),
) int64 {
	_, done, useErr := c.state.begin()
	if useErr != nil {
		return 0
	}
	defer done()

	sink := &ctrlEventSinkAdapter{onControllerAction: fn}
	return c.AddSink(sink)
}
