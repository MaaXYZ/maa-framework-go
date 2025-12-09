package maa

import (
	"image"
	"time"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v3/controller/adb"
	"github.com/MaaXYZ/maa-framework-go/v3/controller/win32"
	"github.com/MaaXYZ/maa-framework-go/v3/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v3/internal/native"
	"github.com/MaaXYZ/maa-framework-go/v3/internal/store"
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
) *Controller {
	handle := native.MaaAdbControllerCreate(
		adbPath,
		address,
		uint64(screencapMethod),
		uint64(inputMethod),
		config,
		agentPath,
	)
	if handle == 0 {
		return nil
	}

	initControllerStore(handle)

	return &Controller{
		handle: handle,
	}
}

// NewWin32Controller creates a win32 controller instance.
func NewWin32Controller(
	hWnd unsafe.Pointer,
	screencapMethod win32.ScreencapMethod,
	mouseMethod win32.InputMethod,
	keyboardMethod win32.InputMethod,
) *Controller {
	handle := native.MaaWin32ControllerCreate(
		hWnd,
		uint64(screencapMethod),
		uint64(mouseMethod),
		uint64(keyboardMethod),
	)
	if handle == 0 {
		return nil
	}

	initControllerStore(handle)

	return &Controller{
		handle: handle,
	}
}

// NewCustomController creates a custom controller instance.
func NewCustomController(
	ctrl CustomController,
) *Controller {
	ctrlID := registerCustomControllerCallbacks(ctrl)
	handle := native.MaaCustomControllerCreate(
		unsafe.Pointer(customControllerCallbacksHandle),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		uintptr(ctrlID),
	)
	if handle == 0 {
		return nil
	}

	store.CtrlStore.Lock()
	store.CtrlStore.Set(handle, store.CtrlStoreValue{
		SinkIDToEventCallbackID:     make(map[int64]uint64),
		CustomControllerCallbacksID: ctrlID,
	})
	store.CtrlStore.Unlock()

	return &Controller{
		handle: handle,
	}
}

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
func (c *Controller) setOption(key native.MaaCtrlOption, value unsafe.Pointer, valSize uintptr) bool {
	return native.MaaControllerSetOption(c.handle, key, value, uint64(valSize))
}

// SetScreenshotTargetLongSide sets screenshot target long side.
// Only one of long and short side can be set, and the other is automatically scaled according to the aspect ratio.
//
// eg: 1280
func (c *Controller) SetScreenshotTargetLongSide(targetLongSide int32) bool {
	return c.setOption(
		native.MaaCtrlOption_ScreenshotTargetLongSide,
		unsafe.Pointer(&targetLongSide),
		unsafe.Sizeof(targetLongSide),
	)
}

// SetScreenshotTargetShortSide sets screenshot target short side.
// Only one of long and short side can be set, and the other is automatically scaled according to the aspect ratio.
//
// eg: 720
func (c *Controller) SetScreenshotTargetShortSide(targetShortSide int32) bool {
	return c.setOption(
		native.MaaCtrlOption_ScreenshotTargetShortSide,
		unsafe.Pointer(&targetShortSide),
		unsafe.Sizeof(targetShortSide),
	)
}

// SetScreenshotUseRawSize sets whether the screenshot uses the raw size without scaling.
func (c *Controller) SetScreenshotUseRawSize(enabled bool) bool {
	return c.setOption(
		native.MaaCtrlOption_ScreenshotUseRawSize,
		unsafe.Pointer(&enabled),
		unsafe.Sizeof(enabled),
	)
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

// PostSwipe posts a swipe.
func (c *Controller) PostSwipe(x1, y1, x2, y2 int32, duration time.Duration) *Job {
	id := native.MaaControllerPostSwipe(c.handle, x1, y1, x2, y2, int32(duration.Milliseconds()))
	return newJob(id, c.status, c.wait)
}

// PostPressKey posts a click key.
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
func (c *Controller) GetShellOutput() (string, bool) {
	output := buffer.NewStringBuffer()
	defer output.Destroy()

	got := native.MaaControllerGetShellOutput(c.handle, output.Handle())
	if !got {
		return "", false
	}
	return output.Get(), true
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
func (c *Controller) CacheImage() image.Image {
	imgBuffer := buffer.NewImageBuffer()
	defer imgBuffer.Destroy()

	got := native.MaaControllerCachedImage(c.handle, imgBuffer.Handle())
	if !got {
		return nil
	}

	img := imgBuffer.Get()

	return img
}

// GetUUID gets the UUID of the controller.
func (c *Controller) GetUUID() (string, bool) {
	uuid := buffer.NewStringBuffer()
	defer uuid.Destroy()
	got := native.MaaControllerGetUuid(c.handle, uuid.Handle())
	if !got {
		return "", false
	}
	return uuid.Get(), true
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
	return sinkId
}

// RemoveSink removes a event callback sink by sink ID.
func (c *Controller) RemoveSink(sinkId int64) {
	native.MaaControllerRemoveSink(c.handle, sinkId)
}

// ClearSinks clears all event callback sinks.
func (c *Controller) ClearSinks() {
	native.MaaControllerClearSinks(c.handle)
}
