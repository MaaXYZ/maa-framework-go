package toolkit

/*
#include <stdlib.h>
#include <MaaToolkit/MaaToolkitAPI.h>
*/
import "C"
import (
	"unsafe"
)

// DesktopWindowFinder is a struct that helps in finding desktop windows.
type DesktopWindowFinder struct {
	handle *C.MaaToolkitDesktopWindowList
}

// NewDesktopWindowFinder creates a new DesktopWindowFinder instance.
func NewDesktopWindowFinder() *DesktopWindowFinder {
	handle := C.MaaToolkitDesktopWindowListCreate()
	return &DesktopWindowFinder{
		handle: handle,
	}
}

// Destroy releases the DesktopWindowFinder.
func (f *DesktopWindowFinder) Destroy() {
	C.MaaToolkitDesktopWindowListDestroy(f.handle)
}

// Find posts a request to find all desktop windows.
func (f *DesktopWindowFinder) Find() bool {
	got := C.MaaToolkitDesktopWindowFindAll(f.handle)
	return got != 0
}

// Size returns the number of windows found.
func (f *DesktopWindowFinder) Size() uint64 {
	size := C.MaaToolkitDesktopWindowListSize(f.handle)
	return uint64(size)
}

// Get returns DesktopWindow by index.
func (f *DesktopWindowFinder) Get(index uint64) *DesktopWindow {
	handle := C.MaaToolkitDesktopWindowListAt(f.handle, C.uint64_t(index))
	if handle == nil {
		return nil
	}
	return &DesktopWindow{
		handle: handle,
	}
}

// DesktopWindow represents a single desktop window with various properties and methods to access its information.
type DesktopWindow struct {
	handle *C.MaaToolkitDesktopWindow
}

// GetHandle returns the window handle.
func (w *DesktopWindow) GetHandle() unsafe.Pointer {
	return C.MaaToolkitDesktopWindowGetHandle(w.handle)
}

// GetClassName returns the window class name.
func (w *DesktopWindow) GetClassName() string {
	name := C.MaaToolkitDesktopWindowGetClassName(w.handle)
	return C.GoString(name)
}

// GetWindowName returns the window window name.
func (w *DesktopWindow) GetWindowName() string {
	name := C.MaaToolkitDesktopWindowGetWindowName(w.handle)
	return C.GoString(name)
}
