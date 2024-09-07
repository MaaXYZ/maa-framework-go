package maa

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

type rectBuffer struct {
	handle *C.MaaRect
}

func newRectBuffer() *rectBuffer {
	handle := C.MaaRectCreate()
	return &rectBuffer{
		handle: handle,
	}
}

func newRectBufferByHandle(handle unsafe.Pointer) *rectBuffer {
	return &rectBuffer{
		handle: (*C.MaaRect)(handle),
	}
}

func (r *rectBuffer) Destroy() {
	C.MaaRectDestroy(r.handle)
}

func (r *rectBuffer) Handle() unsafe.Pointer {
	return unsafe.Pointer(r.handle)
}

func (r *rectBuffer) Get() Rect {
	return Rect{r.GetX(), r.GetY(), r.GetW(), r.GetH()}
}

func (r *rectBuffer) GetX() int32 {
	return int32(C.MaaRectGetX(r.handle))
}

func (r *rectBuffer) GetY() int32 {
	return int32(C.MaaRectGetY(r.handle))
}

func (r *rectBuffer) GetW() int32 {
	return int32(C.MaaRectGetW(r.handle))
}

func (r *rectBuffer) GetH() int32 {
	return int32(C.MaaRectGetH(r.handle))
}

func (r *rectBuffer) Set(rect Rect) bool {
	return C.MaaRectSet(
		r.handle,
		C.int32_t(rect.X),
		C.int32_t(rect.Y),
		C.int32_t(rect.W),
		C.int32_t(rect.H),
	) != 0
}
