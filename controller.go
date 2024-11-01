package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>

extern void _MaaNotificationCallbackAgent(const char* message, const char* details_json, void* callback_arg);
*/
import "C"
import (
	"github.com/MaaXYZ/maa-framework-go/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/internal/store"
	"image"
	"time"
	"unsafe"
)

// Controller is an interface that defines various methods for MAA controller.
type Controller interface {
	Destroy()
	Handle() unsafe.Pointer

	SetScreenshotTargetLongSide(targetLongSide int) bool
	SetScreenshotTargetShortSide(targetShortSide int) bool
	SetScreenshotUseRawSize(enabled bool) bool
	SetRecording(enabled bool) bool

	PostConnect() *Job
	PostClick(x, y int32) *Job
	PostSwipe(x1, y1, x2, y2 int32, duration time.Duration) *Job
	PostPressKey(keycode int32) *Job
	PostInputText(text string) *Job
	PostStartApp(intent string) *Job
	PostStopApp(intent string) *Job
	PostTouchDown(contact, x, y, pressure int32) *Job
	PostTouchMove(contact, x, y, pressure int32) *Job
	PostTouchUp(contact int32) *Job
	PostScreencap() *Job

	Connected() bool
	CacheImage() image.Image
	GetUUID() (string, bool)
}

type controllerStoreValue struct {
	NotificationCallbackID      uint64
	CustomControllerCallbacksID uint64
}

var controllerStore = store.New[controllerStoreValue]()

// controller is a concrete implementation of the Controller interface.
type controller struct {
	handle *C.MaaController
}

// AdbScreencapMethod
//
// Use bitwise OR to set the method you need,
// MaaFramework will test their speed and use the fastest one.
type AdbScreencapMethod uint64

// AdbScreencapMethod
const (
	AdbScreencapMethodNone                AdbScreencapMethod = 0
	AdbScreencapMethodEncodeToFileAndPull AdbScreencapMethod = 1
	AdbScreencapMethodEncode              AdbScreencapMethod = 1 << 1
	AdbScreencapMethodRawWithGzip         AdbScreencapMethod = 1 << 2
	AdbScreencapMethodRawByNetcat         AdbScreencapMethod = 1 << 3
	AdbScreencapMethodMinicapDirect       AdbScreencapMethod = 1 << 4
	AdbScreencapMethodMinicapStream       AdbScreencapMethod = 1 << 5
	AdbScreencapMethodEmulatorExtras      AdbScreencapMethod = 1 << 6

	AdbScreencapMethodAll     = ^AdbScreencapMethodNone
	AdbScreencapMethodDefault = AdbScreencapMethodAll & (^AdbScreencapMethodMinicapDirect) & (^AdbScreencapMethodMinicapStream)
)

// AdbInputMethod
//
// Use bitwise OR to set the method you need,
// MaaFramework will select the available ones according to priority.
// The priority is: EmulatorExtras > Maatouch > MinitouchAndAdbKey > AdbShell
type AdbInputMethod uint64

// AdbInputMethod
const (
	AdbInputMethodNone               AdbInputMethod = 0
	AdbInputMethodAdbShell           AdbInputMethod = 1
	AdbInputMethodMinitouchAndAdbKey AdbInputMethod = 1 << 1
	AdbInputMethodMaatouch           AdbInputMethod = 1 << 2
	AdbInputMethodEmulatorExtras     AdbInputMethod = 1 << 3

	AdbInputMethodAll     = ^AdbInputMethodNone
	AdbInputMethodDefault = AdbInputMethodAll & (^AdbInputMethodEmulatorExtras)
)

// NewAdbController creates an ADB controller instance.
func NewAdbController(
	adbPath, address string,
	screencapMethod AdbScreencapMethod,
	inputMethod AdbInputMethod,
	config, agentPath string,
	notify Notification,
) Controller {
	cAdbPath := C.CString(adbPath)
	cAddress := C.CString(address)
	cConfig := C.CString(config)
	cAgentPath := C.CString(agentPath)
	defer func() {
		C.free(unsafe.Pointer(cAdbPath))
		C.free(unsafe.Pointer(cAddress))
		C.free(unsafe.Pointer(cConfig))
		C.free(unsafe.Pointer(cAgentPath))
	}()

	id := registerNotificationCallback(notify)
	handle := C.MaaAdbControllerCreate(
		cAdbPath,
		cAddress,
		C.uint64_t(screencapMethod),
		C.uint64_t(inputMethod),
		cConfig,
		cAgentPath,
		C.MaaNotificationCallback(C._MaaNotificationCallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	if handle == nil {
		return nil
	}
	controllerStore.Set(unsafe.Pointer(handle), controllerStoreValue{
		NotificationCallbackID: id,
	})
	return &controller{handle: handle}
}

// Win32ScreencapMethod
//
// No bitwise OR, just set it.
type Win32ScreencapMethod uint64

// Win32ScreencapMethod
const (
	Win32ScreencapMethodNone           Win32ScreencapMethod = 0
	Win32ScreencapMethodGDI            Win32ScreencapMethod = 1
	Win32ScreencapMethodFramePool      Win32ScreencapMethod = 1 << 1
	Win32ScreencapMethodDXGIDesktopDup Win32ScreencapMethod = 1 << 2
)

// Win32InputMethod
//
// No bitwise OR, just set it.
type Win32InputMethod uint64

// Win32InputMethod
const (
	Win32InputMethodNone        Win32InputMethod = 0
	Win32InputMethodSeize       Win32InputMethod = 1
	Win32InputMethodSendMessage Win32InputMethod = 1 << 1
)

// NewWin32Controller creates a win32 controller instance.
func NewWin32Controller(
	hWnd unsafe.Pointer,
	screencapMethod Win32ScreencapMethod,
	inputMethod Win32InputMethod,
	notify Notification,
) Controller {
	id := registerNotificationCallback(notify)
	handle := C.MaaWin32ControllerCreate(
		hWnd,
		C.uint64_t(screencapMethod),
		C.uint64_t(inputMethod),
		C.MaaNotificationCallback(C._MaaNotificationCallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	if handle == nil {
		return nil
	}
	controllerStore.Set(unsafe.Pointer(handle), controllerStoreValue{
		NotificationCallbackID: id,
	})
	return &controller{handle: handle}
}

// DbgControllerType
//
// No bitwise OR, just set it.
type DbgControllerType uint64

// DbgControllerType
const (
	DbgControllerTypeNone            DbgControllerType = 0
	DbgControllerTypeCarouselImage   DbgControllerType = 1
	DbgControllerTypeReplayRecording DbgControllerType = 1 << 1
)

// NewDbgController creates a DBG controller instance.
func NewDbgController(
	readPath, writePath string,
	dbgCtrlType DbgControllerType,
	config string,
	notify Notification,
) Controller {
	cReadPath := C.CString(readPath)
	cWritePath := C.CString(writePath)
	cConfig := C.CString(config)
	defer func() {
		C.free(unsafe.Pointer(cReadPath))
		C.free(unsafe.Pointer(cWritePath))
		C.free(unsafe.Pointer(cConfig))
	}()

	id := registerNotificationCallback(notify)
	handle := C.MaaDbgControllerCreate(
		cReadPath,
		cWritePath,
		C.uint64_t(dbgCtrlType),
		cConfig,
		C.MaaNotificationCallback(C._MaaNotificationCallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	if handle == nil {
		return nil
	}
	controllerStore.Set(unsafe.Pointer(handle), controllerStoreValue{
		NotificationCallbackID: id,
	})
	return &controller{handle: handle}
}

// NewCustomController creates a custom controller instance.
func NewCustomController(
	ctrl CustomController,
	notify Notification,
) Controller {
	ctrlID := registerCustomControllerCallbacks(ctrl)
	notifyID := registerNotificationCallback(notify)
	handle := C.MaaCustomControllerCreate(
		(*C.MaaCustomControllerCallbacks)(ctrl.Handle()),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(ctrlID)),
		C.MaaNotificationCallback(C._MaaNotificationCallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(notifyID)),
	)
	if handle == nil {
		return nil
	}
	controllerStore.Set(unsafe.Pointer(handle), controllerStoreValue{
		NotificationCallbackID:      notifyID,
		CustomControllerCallbacksID: ctrlID,
	})
	return &controller{handle: handle}
}

// Destroy frees the controller instance.
func (c *controller) Destroy() {
	value := controllerStore.Get(c.Handle())
	unregisterNotificationCallback(value.NotificationCallbackID)
	unregisterCustomControllerCallbacks(value.CustomControllerCallbacksID)
	controllerStore.Del(c.Handle())
	C.MaaControllerDestroy(c.handle)
}

// Handle returns controller handle.
func (c *controller) Handle() unsafe.Pointer {
	return unsafe.Pointer(c.handle)
}

type ctrlOption int32

// ctrlOption
const (
	ctrlOptionInvalid ctrlOption = 0

	// ctrlOptionScreenshotTargetLongSide specifies that only the long side can be set, and the short side
	// is automatically scaled according to the aspect ratio.
	ctrlOptionScreenshotTargetLongSide ctrlOption = 1

	// ctrlOptionScreenshotTargetShortSide specifies that only the short side can be set, and the long side
	// is automatically scaled according to the aspect ratio.
	ctrlOptionScreenshotTargetShortSide ctrlOption = 2

	// ctrlOptionScreenshotUseRawSize specifies that the screenshot uses the raw size without scaling.
	// Note that this option may cause incorrect coordinates on user devices with different resolutions if scaling is not performed.
	ctrlOptionScreenshotUseRawSize ctrlOption = 3

	/// ctrlOptionRecording indicates that all screenshots and actions should be dumped.
	// Recording will evaluate to true if either this or MaaGlobalOptionEnum::MaaGlobalOption_Recording is true.
	ctrlOptionRecording ctrlOption = 5
)

// setOption sets options for controller instance.
func (c *controller) setOption(key ctrlOption, value unsafe.Pointer, valSize uintptr) bool {
	return C.MaaControllerSetOption(c.handle, C.int32_t(key), C.MaaOptionValue(value), C.uint64_t(valSize)) != 0
}

// SetScreenshotTargetLongSide sets screenshot target long side.
// Only one of long and short side can be set, and the other is automatically scaled according to the aspect ratio.
//
// eg: 1280
func (c *controller) SetScreenshotTargetLongSide(targetLongSide int) bool {
	targetLongSide32 := int32(targetLongSide)
	return c.setOption(
		ctrlOptionScreenshotTargetLongSide,
		unsafe.Pointer(&targetLongSide32),
		unsafe.Sizeof(targetLongSide32),
	)
}

// SetScreenshotTargetShortSide sets screenshot target short side.
// Only one of long and short side can be set, and the other is automatically scaled according to the aspect ratio.
//
// eg: 720
func (c *controller) SetScreenshotTargetShortSide(targetShortSide int) bool {
	targetShortSide32 := int32(targetShortSide)
	return c.setOption(
		ctrlOptionScreenshotTargetShortSide,
		unsafe.Pointer(&targetShortSide32),
		unsafe.Sizeof(targetShortSide32),
	)
}

// SetScreenshotUseRawSize sets whether the screenshot uses the raw size without scaling.
func (c *controller) SetScreenshotUseRawSize(enabled bool) bool {
	var cEnabled uint8
	if enabled {
		cEnabled = 1
	}
	return c.setOption(
		ctrlOptionScreenshotUseRawSize,
		unsafe.Pointer(&cEnabled),
		unsafe.Sizeof(cEnabled),
	)
}

// SetRecording sets whether to dump all screenshots and actions.
func (c *controller) SetRecording(enabled bool) bool {
	var cEnabled uint8
	if enabled {
		cEnabled = 1
	}

	return c.setOption(
		ctrlOptionRecording,
		unsafe.Pointer(&cEnabled),
		unsafe.Sizeof(cEnabled),
	)
}

// PostConnect posts a connection.
func (c *controller) PostConnect() *Job {
	id := int64(C.MaaControllerPostConnection(c.handle))
	return NewJob(id, c.status, c.wait)
}

// PostClick posts a click.
func (c *controller) PostClick(x, y int32) *Job {
	id := int64(C.MaaControllerPostClick(c.handle, C.int32_t(x), C.int32_t(y)))
	return NewJob(id, c.status, c.wait)
}

// PostSwipe posts a swipe.
func (c *controller) PostSwipe(x1, y1, x2, y2 int32, duration time.Duration) *Job {
	id := int64(C.MaaControllerPostSwipe(
		c.handle,
		C.int32_t(x1),
		C.int32_t(y1),
		C.int32_t(x2),
		C.int32_t(y2),
		C.int32_t(duration.Milliseconds()),
	))
	return NewJob(id, c.status, c.wait)
}

// PostPressKey posts a press key.
func (c *controller) PostPressKey(keycode int32) *Job {
	id := int64(C.MaaControllerPostPressKey(c.handle, C.int32_t(keycode)))
	return NewJob(id, c.status, c.wait)
}

// PostInputText posts an input text.
func (c *controller) PostInputText(text string) *Job {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	id := int64(C.MaaControllerPostInputText(c.handle, cText))
	return NewJob(id, c.status, c.wait)
}

// PostStartApp posts a start app.
func (c *controller) PostStartApp(intent string) *Job {
	cIntent := C.CString(intent)
	defer C.free(unsafe.Pointer(cIntent))
	id := int64(C.MaaControllerPostStartApp(c.handle, cIntent))
	return NewJob(id, c.status, c.wait)
}

// PostStopApp posts a stop app.
func (c *controller) PostStopApp(intent string) *Job {
	cIntent := C.CString(intent)
	defer C.free(unsafe.Pointer(cIntent))
	id := int64(C.MaaControllerPostStopApp(c.handle, cIntent))
	return NewJob(id, c.status, c.wait)
}

// PostTouchDown posts a touch-down.
func (c *controller) PostTouchDown(contact, x, y, pressure int32) *Job {
	id := int64(C.MaaControllerPostTouchDown(c.handle, C.int32_t(contact), C.int32_t(x), C.int32_t(y), C.int32_t(pressure)))
	return NewJob(id, c.status, c.wait)
}

// PostTouchMove posts a touch-move.
func (c *controller) PostTouchMove(contact, x, y, pressure int32) *Job {
	id := int64(C.MaaControllerPostTouchMove(c.handle, C.int32_t(contact), C.int32_t(x), C.int32_t(y), C.int32_t(pressure)))
	return NewJob(id, c.status, c.wait)
}

// PostTouchUp posts a touch-up.
func (c *controller) PostTouchUp(contact int32) *Job {
	id := int64(C.MaaControllerPostTouchUp(c.handle, C.int32_t(contact)))
	return NewJob(id, c.status, c.wait)
}

// PostScreencap posts a screencap.
func (c *controller) PostScreencap() *Job {
	id := int64(C.MaaControllerPostScreencap(c.handle))
	return NewJob(id, c.status, c.wait)
}

// status gets the status of a request identified by the given id.
func (c *controller) status(id int64) Status {
	return Status(C.MaaControllerStatus(c.handle, C.int64_t(id)))
}

func (c *controller) wait(id int64) Status {
	return Status(C.MaaControllerWait(c.handle, C.int64_t(id)))
}

// Connected checks if the controller is connected.
func (c *controller) Connected() bool {
	return C.MaaControllerConnected(c.handle) != 0
}

// CacheImage gets the image buffer of the last screencap request.
func (c *controller) CacheImage() image.Image {
	imgBuffer := buffer.NewImageBuffer()
	defer imgBuffer.Destroy()

	got := C.MaaControllerCachedImage(
		c.handle,
		(*C.MaaImageBuffer)(imgBuffer.Handle()),
	) != 0
	if !got {
		return nil
	}

	img := imgBuffer.Get()

	return img
}

// GetUUID gets the UUID of the controller.
func (c *controller) GetUUID() (string, bool) {
	uuid := buffer.NewStringBuffer()
	defer uuid.Destroy()
	got := C.MaaControllerGetUuid(
		c.handle,
		(*C.MaaStringBuffer)(uuid.Handle()),
	) != 0
	if !got {
		return "", false
	}
	return uuid.Get(), true
}
