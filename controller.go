package maa

import (
	"fmt"
	"image"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v2/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v2/internal/maa"
	"github.com/MaaXYZ/maa-framework-go/v2/internal/store"
)

// Controller is an interface that defines various methods for MAA controller.
type Controller interface {
	Destroy()
	Handle() uintptr

	SetScreenshotTargetLongSide(targetLongSide int32) bool
	SetScreenshotTargetShortSide(targetShortSide int32) bool
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

var (
	controllerStore      = store.New[controllerStoreValue]()
	controllerStoreMutex sync.RWMutex
)

// controller is a concrete implementation of the Controller interface.
type controller struct {
	handle uintptr
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
	AdbScreencapMethodDefault = AdbScreencapMethodAll & (^AdbScreencapMethodRawByNetcat) & (^AdbScreencapMethodMinicapDirect) & (^AdbScreencapMethodMinicapStream)
)

func (m AdbScreencapMethod) String() string {
	switch m {
	case AdbScreencapMethodNone:
		return ""
	case AdbScreencapMethodEncodeToFileAndPull:
		return AdbScreencapMethodEncodeToFileAndPullValue
	case AdbScreencapMethodEncode:
		return AdbScreencapMethodEncodeValue
	case AdbScreencapMethodRawWithGzip:
		return AdbScreencapMethodRawWithGzipValue
	case AdbScreencapMethodRawByNetcat:
		return AdbScreencapMethodRawByNetcatValue
	case AdbScreencapMethodMinicapDirect:
		return AdbScreencapMethodMinicapDirectValue
	case AdbScreencapMethodMinicapStream:
		return AdbScreencapMethodMinicapStreamValue
	case AdbScreencapMethodEmulatorExtras:
		return AdbScreencapMethodEmulatorExtrasValue
	case AdbScreencapMethodAll:
		return AdbScreencapMethodAllValue
	case AdbScreencapMethodDefault:
		return AdbScreencapMethodDefaultValue
	}
	return strconv.Itoa(int(m))
}

func ParseAdbScreencapMethod(methodStr string) (AdbScreencapMethod, error) {
	switch {
	case strings.EqualFold(methodStr, AdbScreencapMethodEncodeToFileAndPull.String()):
		return AdbScreencapMethodEncodeToFileAndPull, nil
	case strings.EqualFold(methodStr, AdbScreencapMethodEncode.String()):
		return AdbScreencapMethodEncode, nil
	case strings.EqualFold(methodStr, AdbScreencapMethodRawWithGzip.String()):
		return AdbScreencapMethodRawWithGzip, nil
	case strings.EqualFold(methodStr, AdbScreencapMethodRawByNetcat.String()):
		return AdbScreencapMethodRawByNetcat, nil
	case strings.EqualFold(methodStr, AdbScreencapMethodMinicapDirect.String()):
		return AdbScreencapMethodMinicapDirect, nil
	case strings.EqualFold(methodStr, AdbScreencapMethodMinicapStream.String()):
		return AdbScreencapMethodMinicapStream, nil
	case strings.EqualFold(methodStr, AdbScreencapMethodEmulatorExtras.String()):
		return AdbScreencapMethodEmulatorExtras, nil
	case strings.EqualFold(methodStr, AdbScreencapMethodAll.String()):
		return AdbScreencapMethodAll, nil
	case strings.EqualFold(methodStr, AdbScreencapMethodDefault.String()):
		return AdbScreencapMethodDefault, nil
	}
	i, err := strconv.Atoi(methodStr)
	if err != nil {
		return AdbScreencapMethodNone, fmt.Errorf("unknown Adb Screencap Method String: '%s', defaulting to AdbScreencapMethodNone", methodStr)
	}
	return AdbScreencapMethod(i), nil
}

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

func (m AdbInputMethod) String() string {
	switch m {
	case AdbInputMethodNone:
		return ""
	case AdbInputMethodAdbShell:
		return AdbInputMethodAdbShellValue
	case AdbInputMethodMinitouchAndAdbKey:
		return AdbInputMethodMinitouchAndAdbKeyValue
	case AdbInputMethodMaatouch:
		return AdbInputMethodMaatouchValue
	case AdbInputMethodEmulatorExtras:
		return AdbInputMethodEmulatorExtrasValue
	case AdbInputMethodAll:
		return AdbInputMethodAllValue
	case AdbInputMethodDefault:
		return AdbInputMethodDefaultValue
	}
	return strconv.Itoa(int(m))
}

func ParseAdbInputMethod(methodStr string) (AdbInputMethod, error) {
	switch {
	case strings.EqualFold(methodStr, AdbInputMethodAdbShell.String()):
		return AdbInputMethodAdbShell, nil
	case strings.EqualFold(methodStr, AdbInputMethodMinitouchAndAdbKey.String()):
		return AdbInputMethodMinitouchAndAdbKey, nil
	case strings.EqualFold(methodStr, AdbInputMethodMaatouch.String()):
		return AdbInputMethodMaatouch, nil
	case strings.EqualFold(methodStr, AdbInputMethodEmulatorExtras.String()):
		return AdbInputMethodEmulatorExtras, nil
	case strings.EqualFold(methodStr, AdbInputMethodAll.String()):
		return AdbInputMethodAll, nil
	case strings.EqualFold(methodStr, AdbInputMethodDefault.String()):
		return AdbInputMethodDefault, nil
	}
	i, err := strconv.Atoi(methodStr)
	if err != nil {
		return AdbInputMethodNone, fmt.Errorf("unknown Adb Input Method String: '%s', defaulting to AdbInputMethodNone", methodStr)
	}
	return AdbInputMethod(i), nil
}

// NewAdbController creates an ADB controller instance.
func NewAdbController(
	adbPath, address string,
	screencapMethod AdbScreencapMethod,
	inputMethod AdbInputMethod,
	config, agentPath string,
	notify Notification,
) Controller {
	id := registerNotificationCallback(notify)
	handle := maa.MaaAdbControllerCreate(
		adbPath,
		address,
		maa.MaaAdbScreencapMethod(screencapMethod),
		maa.MaaAdbInputMethod(inputMethod),
		config,
		agentPath,
		_MaaNotificationCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		uintptr(id),
	)
	if handle == 0 {
		return nil
	}

	controllerStoreMutex.Lock()
	controllerStore.Set(handle, controllerStoreValue{
		NotificationCallbackID: id,
	})
	controllerStoreMutex.Unlock()

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

func (m Win32ScreencapMethod) String() string {
	switch m {
	case Win32ScreencapMethodNone:
		return ""
	case Win32ScreencapMethodGDI:
		return Win32ScreencapMethodGDIValue
	case Win32ScreencapMethodFramePool:
		return Win32ScreencapMethodFramePoolValue
	case Win32ScreencapMethodDXGIDesktopDup:
		return Win32ScreencapMethodDXGIDesktopDupValue
	}
	return strconv.Itoa(int(m))
}

func ParseWin32ScreencapMethod(methodStr string) (Win32ScreencapMethod, error) {
	switch {
	case strings.EqualFold(methodStr, Win32ScreencapMethodGDI.String()):
		return Win32ScreencapMethodGDI, nil
	case strings.EqualFold(methodStr, Win32ScreencapMethodFramePool.String()):
		return Win32ScreencapMethodFramePool, nil
	case strings.EqualFold(methodStr, Win32ScreencapMethodDXGIDesktopDup.String()):
		return Win32ScreencapMethodDXGIDesktopDup, nil
	}
	i, err := strconv.Atoi(methodStr)
	if err != nil {
		return Win32ScreencapMethodNone, fmt.Errorf("unknown Win32 Screencap Method String: '%s', defaulting to Win32ScreencapMethodNone", methodStr)
	}
	return Win32ScreencapMethod(i), nil
}

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

func (m Win32InputMethod) String() string {
	switch m {
	case Win32InputMethodNone:
		return ""
	case Win32InputMethodSeize:
		return Win32InputMethodSeizeValue
	case Win32InputMethodSendMessage:
		return Win32InputMethodSendMessageValue
	}
	return strconv.Itoa(int(m))
}

func ParseWin32InputMethod(methodStr string) (Win32InputMethod, error) {
	switch {
	case strings.EqualFold(methodStr, Win32InputMethodSeize.String()):
		return Win32InputMethodSeize, nil
	case strings.EqualFold(methodStr, Win32InputMethodSendMessage.String()):
		return Win32InputMethodSendMessage, nil
	}
	i, err := strconv.Atoi(methodStr)
	if err != nil {
		return Win32InputMethodNone, fmt.Errorf("unknown Win32 Input Method String: '%s', defaulting to Win32InputMethodNone", methodStr)
	}
	return Win32InputMethod(i), nil
}

// NewWin32Controller creates a win32 controller instance.
func NewWin32Controller(
	hWnd unsafe.Pointer,
	screencapMethod Win32ScreencapMethod,
	inputMethod Win32InputMethod,
	notify Notification,
) Controller {
	id := registerNotificationCallback(notify)
	handle := maa.MaaWin32ControllerCreate(
		hWnd,
		maa.MaaWin32ScreencapMethod(screencapMethod),
		maa.MaaWin32InputMethod(inputMethod),
		_MaaNotificationCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		uintptr(id),
	)
	if handle == 0 {
		return nil
	}

	controllerStoreMutex.Lock()
	controllerStore.Set(handle, controllerStoreValue{
		NotificationCallbackID: id,
	})
	controllerStoreMutex.Unlock()

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

func (t DbgControllerType) String() string {
	switch t {
	case DbgControllerTypeNone:
		return ""
	case DbgControllerTypeCarouselImage:
		return DbgControllerTypeCarouselImageValue
	case DbgControllerTypeReplayRecording:
		return DbgControllerTypeReplayRecordingValue
	}
	return strconv.Itoa(int(t))
}

func ParseDbgControllerType(typeStr string) (DbgControllerType, error) {
	switch {
	case strings.EqualFold(typeStr, DbgControllerTypeCarouselImage.String()):
		return DbgControllerTypeCarouselImage, nil
	case strings.EqualFold(typeStr, DbgControllerTypeReplayRecording.String()):
		return DbgControllerTypeReplayRecording, nil
	}
	i, err := strconv.Atoi(typeStr)
	if err != nil {
		return DbgControllerTypeNone, fmt.Errorf("unknown DBG Controller Type String: '%s', defaulting to DbgControllerTypeNone", typeStr)
	}
	return DbgControllerType(i), nil
}

// NewDbgController creates a DBG controller instance.
func NewDbgController(
	readPath, writePath string,
	dbgCtrlType DbgControllerType,
	config string,
	notify Notification,
) Controller {
	id := registerNotificationCallback(notify)
	handle := maa.MaaDbgControllerCreate(
		readPath,
		writePath,
		maa.MaaDbgControllerType(dbgCtrlType),
		config,
		_MaaNotificationCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		uintptr(id),
	)
	if handle == 0 {
		return nil
	}

	controllerStoreMutex.Lock()
	controllerStore.Set(handle, controllerStoreValue{
		NotificationCallbackID: id,
	})
	controllerStoreMutex.Unlock()

	return &controller{handle: handle}
}

// NewCustomController creates a custom controller instance.
func NewCustomController(
	ctrl CustomController,
	notify Notification,
) Controller {
	ctrlID := registerCustomControllerCallbacks(ctrl)
	notifyID := registerNotificationCallback(notify)
	handle := maa.MaaCustomControllerCreate(
		uintptr(ctrl.Handle()),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		uintptr(ctrlID),
		_MaaNotificationCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		uintptr(notifyID),
	)
	if handle == 0 {
		return nil
	}

	controllerStoreMutex.Lock()
	controllerStore.Set(handle, controllerStoreValue{
		NotificationCallbackID:      notifyID,
		CustomControllerCallbacksID: ctrlID,
	})
	controllerStoreMutex.Unlock()

	return &controller{handle: handle}
}

// Destroy frees the controller instance.
func (c *controller) Destroy() {
	controllerStoreMutex.Lock()
	value := controllerStore.Get(c.handle)
	unregisterNotificationCallback(value.NotificationCallbackID)
	unregisterCustomControllerCallbacks(value.CustomControllerCallbacksID)
	controllerStore.Del(c.handle)
	controllerStoreMutex.Unlock()

	maa.MaaControllerDestroy(c.handle)
}

// Handle returns controller handle.
func (c *controller) Handle() uintptr {
	return c.handle
}

// setOption sets options for controller instance.
func (c *controller) setOption(key maa.MaaCtrlOption, value unsafe.Pointer, valSize uintptr) bool {
	return maa.MaaControllerSetOption(c.handle, key, value, uint64(valSize))
}

// SetScreenshotTargetLongSide sets screenshot target long side.
// Only one of long and short side can be set, and the other is automatically scaled according to the aspect ratio.
//
// eg: 1280
func (c *controller) SetScreenshotTargetLongSide(targetLongSide int32) bool {
	return c.setOption(
		maa.MaaCtrlOption_ScreenshotTargetLongSide,
		unsafe.Pointer(&targetLongSide),
		unsafe.Sizeof(targetLongSide),
	)
}

// SetScreenshotTargetShortSide sets screenshot target short side.
// Only one of long and short side can be set, and the other is automatically scaled according to the aspect ratio.
//
// eg: 720
func (c *controller) SetScreenshotTargetShortSide(targetShortSide int32) bool {
	return c.setOption(
		maa.MaaCtrlOption_ScreenshotTargetShortSide,
		unsafe.Pointer(&targetShortSide),
		unsafe.Sizeof(targetShortSide),
	)
}

// SetScreenshotUseRawSize sets whether the screenshot uses the raw size without scaling.
func (c *controller) SetScreenshotUseRawSize(enabled bool) bool {
	return c.setOption(
		maa.MaaCtrlOption_ScreenshotUseRawSize,
		unsafe.Pointer(&enabled),
		unsafe.Sizeof(enabled),
	)
}

// SetRecording sets whether to dump all screenshots and actions.
func (c *controller) SetRecording(enabled bool) bool {
	return c.setOption(
		maa.MaaCtrlOption_Recording,
		unsafe.Pointer(&enabled),
		unsafe.Sizeof(enabled),
	)
}

// PostConnect posts a connection.
func (c *controller) PostConnect() *Job {
	id := maa.MaaControllerPostConnection(c.handle)
	return NewJob(id, c.status, c.wait)
}

// PostClick posts a click.
func (c *controller) PostClick(x, y int32) *Job {
	id := maa.MaaControllerPostClick(c.handle, x, y)
	return NewJob(id, c.status, c.wait)
}

// PostSwipe posts a swipe.
func (c *controller) PostSwipe(x1, y1, x2, y2 int32, duration time.Duration) *Job {
	id := maa.MaaControllerPostSwipe(c.handle, x1, y1, x2, y2, int32(duration.Milliseconds()))
	return NewJob(id, c.status, c.wait)
}

// PostPressKey posts a press key.
func (c *controller) PostPressKey(keycode int32) *Job {
	id := maa.MaaControllerPostPressKey(c.handle, keycode)
	return NewJob(id, c.status, c.wait)
}

// PostInputText posts an input text.
func (c *controller) PostInputText(text string) *Job {
	id := maa.MaaControllerPostInputText(c.handle, text)
	return NewJob(id, c.status, c.wait)
}

// PostStartApp posts a start app.
func (c *controller) PostStartApp(intent string) *Job {
	id := maa.MaaControllerPostStartApp(c.handle, intent)
	return NewJob(id, c.status, c.wait)
}

// PostStopApp posts a stop app.
func (c *controller) PostStopApp(intent string) *Job {
	id := maa.MaaControllerPostStopApp(c.handle, intent)
	return NewJob(id, c.status, c.wait)
}

// PostTouchDown posts a touch-down.
func (c *controller) PostTouchDown(contact, x, y, pressure int32) *Job {
	id := maa.MaaControllerPostTouchDown(c.handle, contact, x, y, pressure)
	return NewJob(id, c.status, c.wait)
}

// PostTouchMove posts a touch-move.
func (c *controller) PostTouchMove(contact, x, y, pressure int32) *Job {
	id := maa.MaaControllerPostTouchMove(c.handle, contact, x, y, pressure)
	return NewJob(id, c.status, c.wait)
}

// PostTouchUp posts a touch-up.
func (c *controller) PostTouchUp(contact int32) *Job {
	id := maa.MaaControllerPostTouchUp(c.handle, contact)
	return NewJob(id, c.status, c.wait)
}

// PostScreencap posts a screencap.
func (c *controller) PostScreencap() *Job {
	id := maa.MaaControllerPostScreencap(c.handle)
	return NewJob(id, c.status, c.wait)
}

// status gets the status of a request identified by the given id.
func (c *controller) status(id int64) Status {
	return Status(maa.MaaControllerStatus(c.handle, id))
}

func (c *controller) wait(id int64) Status {
	return Status(maa.MaaControllerWait(c.handle, id))
}

// Connected checks if the controller is connected.
func (c *controller) Connected() bool {
	return maa.MaaControllerConnected(c.handle)
}

// CacheImage gets the image buffer of the last screencap request.
func (c *controller) CacheImage() image.Image {
	imgBuffer := buffer.NewImageBuffer()
	defer imgBuffer.Destroy()

	got := maa.MaaControllerCachedImage(c.handle, imgBuffer.Handle())
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
	got := maa.MaaControllerGetUuid(c.handle, uuid.Handle())
	if !got {
		return "", false
	}
	return uuid.Get(), true
}
