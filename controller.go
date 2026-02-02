package maa

import (
	"errors"
	"fmt"
	"image"
	"time"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v4/controller/adb"
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

type Controller struct {
	handle uintptr
}

// NewAdbController creates a new ADB controller.
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

	return &Controller{
		handle: handle,
	}, nil
}

// NewPlayCoverController creates a new PlayCover controller.
func NewPlayCoverController(
	address, uuid string,
) (*Controller, error) {
	handle := native.MaaPlayCoverControllerCreate(address, uuid)
	if handle == 0 {
		return nil, errors.New("failed to create PlayCover controller")
	}

	initControllerStore(handle)

	return &Controller{
		handle: handle,
	}, nil
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

	return &Controller{
		handle: handle,
	}, nil
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
// For gamepad button and touch constants, import "github.com/MaaXYZ/maa-framework-go/v3/controller/gamepad".
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

	return &Controller{
		handle: handle,
	}, nil
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
		return nil, errors.New("failed to create Custom controller")
	}

	initControllerStore(handle)

	store.CtrlStore.Update(handle, func(v *store.CtrlStoreValue) {
		v.CustomControllerCallbacksID = ctrlID
	})

	return &Controller{
		handle: handle,
	}, nil
}

// NOTE: MaaDbgController is intentionally NOT implemented in Go binding.
// Use CarouselImageController or BlankController from dbg_controller.go for debugging purposes.
// Do not add NewDbgController here.

// Destroy frees the controller instance.
func (c *Controller) Destroy() {
	store.CtrlStore.Lock()
	value := store.CtrlStore.Get(c.handle)
	unregisterCustomControllerCallbacks(value.CustomControllerCallbacksID)
	for _, cbID := range value.SinkIDToEventCallbackID {
		unregisterEventCallback(cbID)
	}
	store.CtrlStore.Del(c.handle)
	store.CtrlStore.Unlock()

	native.MaaControllerDestroy(c.handle)
}

// setOption sets options for controller instance.
func (c *Controller) setOption(key native.MaaCtrlOption, value unsafe.Pointer, valSize uintptr) error {
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
)

type screenshotOptionConfig struct {
	kind            screenshotOptionKind
	targetLongSide  int32
	targetShortSide int32
	useRawSize      bool
}

// ScreenshotOption configures how the screenshot is resized.
// If multiple options are provided, only the last one is applied.
type ScreenshotOption func(*screenshotOptionConfig)

// WithScreenshotTargetLongSide sets screenshot target long side.
// Only one of long and short side can be set, and the other is automatically scaled according to the aspect ratio.
//
// eg: 1280
func WithScreenshotTargetLongSide(targetLongSide int32) ScreenshotOption {
	return func(cfg *screenshotOptionConfig) {
		cfg.kind = screenshotOptionLongSide
		cfg.targetLongSide = targetLongSide
	}
}

// WithScreenshotTargetShortSide sets screenshot target short side.
// Only one of long and short side can be set, and the other is automatically scaled according to the aspect ratio.
//
// eg: 720
func WithScreenshotTargetShortSide(targetShortSide int32) ScreenshotOption {
	return func(cfg *screenshotOptionConfig) {
		cfg.kind = screenshotOptionShortSide
		cfg.targetShortSide = targetShortSide
	}
}

// WithScreenshotUseRawSize sets whether the screenshot uses the raw size without scaling.
func WithScreenshotUseRawSize(enabled bool) ScreenshotOption {
	return func(cfg *screenshotOptionConfig) {
		cfg.kind = screenshotOptionRawSize
		cfg.useRawSize = enabled
	}
}

// SetScreenshot applies screenshot options to controller instance.
// Only the last option is applied when multiple options are provided.
func (c *Controller) SetScreenshot(opts ...ScreenshotOption) error {
	cfg := screenshotOptionConfig{
		kind: screenshotOptionUnset,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	switch cfg.kind {
	case screenshotOptionUnset:
		return nil
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
	default:
		return fmt.Errorf("unknown screenshot option kind: %v", cfg.kind)
	}
}

// PostConnect posts a connection.
func (c *Controller) PostConnect() *Job {
	id := native.MaaControllerPostConnection(c.handle)
	return newJob(id, c.status, c.wait)
}

// PostClick posts a click.
func (c *Controller) PostClick(x, y int32) *Job {
	id := native.MaaControllerPostClick(c.handle, x, y)
	return newJob(id, c.status, c.wait)
}

// PostClickV2 posts a click with contact and pressure.
// For adb controller, contact means finger id (0 for first finger, 1 for second finger, etc).
// For win32 controller, contact means mouse button id (0 for left, 1 for right, 2 for middle).
func (c *Controller) PostClickV2(x, y, contact, pressure int32) *Job {
	id := native.MaaControllerPostClickV2(c.handle, x, y, contact, pressure)
	return newJob(id, c.status, c.wait)
}

// PostSwipe posts a swipe.
func (c *Controller) PostSwipe(x1, y1, x2, y2 int32, duration time.Duration) *Job {
	id := native.MaaControllerPostSwipe(c.handle, x1, y1, x2, y2, int32(duration.Milliseconds()))
	return newJob(id, c.status, c.wait)
}

// PostSwipeV2 posts a swipe with contact and pressure.
// For adb controller, contact means finger id (0 for first finger, 1 for second finger, etc).
// For win32 controller, contact means mouse button id (0 for left, 1 for right, 2 for middle).
func (c *Controller) PostSwipeV2(x1, y1, x2, y2 int32, duration time.Duration, contact, pressure int32) *Job {
	id := native.MaaControllerPostSwipeV2(c.handle, x1, y1, x2, y2, int32(duration.Milliseconds()), contact, pressure)
	return newJob(id, c.status, c.wait)
}

// PostClickKey posts a click key.
func (c *Controller) PostClickKey(keycode int32) *Job {
	id := native.MaaControllerPostClickKey(c.handle, keycode)
	return newJob(id, c.status, c.wait)
}

// PostInputText posts an input text.
func (c *Controller) PostInputText(text string) *Job {
	id := native.MaaControllerPostInputText(c.handle, text)
	return newJob(id, c.status, c.wait)
}

// PostStartApp posts a start app.
func (c *Controller) PostStartApp(intent string) *Job {
	id := native.MaaControllerPostStartApp(c.handle, intent)
	return newJob(id, c.status, c.wait)
}

// PostStopApp posts a stop app.
func (c *Controller) PostStopApp(intent string) *Job {
	id := native.MaaControllerPostStopApp(c.handle, intent)
	return newJob(id, c.status, c.wait)
}

// PostTouchDown posts a touch-down.
func (c *Controller) PostTouchDown(contact, x, y, pressure int32) *Job {
	id := native.MaaControllerPostTouchDown(c.handle, contact, x, y, pressure)
	return newJob(id, c.status, c.wait)
}

// PostTouchMove posts a touch-move.
func (c *Controller) PostTouchMove(contact, x, y, pressure int32) *Job {
	id := native.MaaControllerPostTouchMove(c.handle, contact, x, y, pressure)
	return newJob(id, c.status, c.wait)
}

// PostTouchUp posts a touch-up.
func (c *Controller) PostTouchUp(contact int32) *Job {
	id := native.MaaControllerPostTouchUp(c.handle, contact)
	return newJob(id, c.status, c.wait)
}

func (c *Controller) PostKeyDown(keycode int32) *Job {
	id := native.MaaControllerPostKeyDown(c.handle, keycode)
	return newJob(id, c.status, c.wait)
}

func (c *Controller) PostKeyUp(keycode int32) *Job {
	id := native.MaaControllerPostKeyUp(c.handle, keycode)
	return newJob(id, c.status, c.wait)
}

// PostScreencap posts a screencap.
func (c *Controller) PostScreencap() *Job {
	id := native.MaaControllerPostScreencap(c.handle)
	return newJob(id, c.status, c.wait)
}

// PostScroll posts a scroll.
func (c *Controller) PostScroll(dx, dy int32) *Job {
	id := native.MaaControllerPostScroll(c.handle, dx, dy)
	return newJob(id, c.status, c.wait)
}

// PostShell posts a adb shell command.
// This is only valid for ADB controllers. If the controller is not an ADB controller, the action will fail.
func (c *Controller) PostShell(cmd string, timeout time.Duration) *Job {
	id := native.MaaControllerPostShell(c.handle, cmd, timeout.Milliseconds())
	return newJob(id, c.status, c.wait)
}

// GetShellOutput gets the output of the last shell command.
func (c *Controller) GetShellOutput() (string, error) {
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
	return Status(native.MaaControllerStatus(c.handle, id))
}

func (c *Controller) wait(id int64) Status {
	return Status(native.MaaControllerWait(c.handle, id))
}

// Connected checks if the controller is connected.
func (c *Controller) Connected() bool {
	return native.MaaControllerConnected(c.handle)
}

// CacheImage gets the image buffer of the last screencap request.
func (c *Controller) CacheImage() (image.Image, error) {
	imgBuffer := buffer.NewImageBuffer()
	defer imgBuffer.Destroy()

	got := native.MaaControllerCachedImage(c.handle, imgBuffer.Handle())
	if !got {
		return nil, errors.New("failed to get cached image")
	}

	img := imgBuffer.Get()

	return img, nil
}

// GetUUID gets the UUID of the controller.
func (c *Controller) GetUUID() (string, error) {
	uuid := buffer.NewStringBuffer()
	defer uuid.Destroy()
	got := native.MaaControllerGetUuid(c.handle, uuid.Handle())
	if !got {
		return "", errors.New("failed to get UUID")
	}
	return uuid.Get(), nil
}

// GetResolution gets the raw (unscaled) device resolution.
// Returns the width and height. Returns an error if the resolution is not available.
// Note: This returns the actual device screen resolution before any scaling.
// The screenshot obtained via CacheImage is scaled according to the screenshot target size settings.
func (c *Controller) GetResolution() (width, height int32, err error) {
	got := native.MaaControllerGetResolution(c.handle, &width, &height)
	if !got {
		return 0, 0, fmt.Errorf("failed to get resolution")
	}
	return width, height, nil
}

// AddSink adds a event callback sink and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (c *Controller) AddSink(sink ControllerEventSink) int64 {
	id := registerEventCallback(sink)
	sinkId := native.MaaControllerAddSink(
		c.handle,
		_MaaEventCallbackAgent,
		uintptr(id),
	)

	store.CtrlStore.Update(c.handle, func(v *store.CtrlStoreValue) {
		v.SinkIDToEventCallbackID[sinkId] = id
	})

	return sinkId
}

// RemoveSink removes a event callback sink by sink ID.
func (c *Controller) RemoveSink(sinkId int64) {
	store.CtrlStore.Update(c.handle, func(v *store.CtrlStoreValue) {
		unregisterEventCallback(v.SinkIDToEventCallbackID[sinkId])
		delete(v.SinkIDToEventCallbackID, sinkId)
	})

	native.MaaControllerRemoveSink(c.handle, sinkId)
}

// ClearSinks clears all event callback sinks.
func (c *Controller) ClearSinks() {
	store.CtrlStore.Update(c.handle, func(v *store.CtrlStoreValue) {
		for _, id := range v.SinkIDToEventCallbackID {
			unregisterEventCallback(id)
		}
		v.SinkIDToEventCallbackID = make(map[int64]uint64)
	})

	native.MaaControllerClearSinks(c.handle)
}

type ControllerEventSink interface {
	OnControllerAction(ctrl *Controller, event EventStatus, detail ControllerActionDetail)
}

// ctrlEventSinkAdapter is a lightweight adapter that makes it easy to register
// a single-event handler via a callback function.
type ctrlEventSinkAdapter struct {
	onControllerAction func(EventStatus, ControllerActionDetail)
}

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
	sink := &ctrlEventSinkAdapter{onControllerAction: fn}
	return c.AddSink(sink)
}
