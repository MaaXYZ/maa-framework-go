package buffer

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import (
	"unsafe"
)

type Rect struct {
	X, Y, W, H int32
}

type RectBuffer struct {
	handle *C.MaaRect
}

func NewRectBuffer() *RectBuffer {
	handle := C.MaaRectCreate()
	return &RectBuffer{
		handle: handle,
	}
}

func NewRectBufferByHandle(handle unsafe.Pointer) *RectBuffer {
	return &RectBuffer{
		handle: (*C.MaaRect)(handle),
	}
}

func (r *RectBuffer) Destroy() {
	C.MaaRectDestroy(r.handle)
}

func (r *RectBuffer) Handle() unsafe.Pointer {
	return unsafe.Pointer(r.handle)
}

func (r *RectBuffer) Get() Rect {
	return Rect{r.GetX(), r.GetY(), r.GetW(), r.GetH()}
}

func (r *RectBuffer) GetX() int32 {
	return int32(C.MaaRectGetX(r.handle))
}

func (r *RectBuffer) GetY() int32 {
	return int32(C.MaaRectGetY(r.handle))
}

func (r *RectBuffer) GetW() int32 {
	return int32(C.MaaRectGetW(r.handle))
}

func (r *RectBuffer) GetH() int32 {
	return int32(C.MaaRectGetH(r.handle))
}

func (r *RectBuffer) Set(rect Rect) bool {
	return C.MaaRectSet(
		r.handle,
		C.int32_t(rect.X),
		C.int32_t(rect.Y),
		C.int32_t(rect.W),
		C.int32_t(rect.H),
	) != 0
}
