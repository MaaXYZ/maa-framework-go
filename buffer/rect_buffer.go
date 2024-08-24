package buffer

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import "unsafe"

type Rect struct {
	X, Y, W, H int32
}

type RectBuffer struct {
	handle C.MaaRectHandle
}

func NewRectBuffer() *RectBuffer {
	handle := C.MaaCreateRectBuffer()
	return &RectBuffer{
		handle: handle,
	}
}

func NewRectBufferByHandle(handle unsafe.Pointer) *RectBuffer {
	return &RectBuffer{
		handle: C.MaaRectHandle(handle),
	}
}

func (r *RectBuffer) Destroy() {
	C.MaaDestroyRectBuffer(r.handle)
}

func (r *RectBuffer) Handle() unsafe.Pointer {
	return unsafe.Pointer(r.handle)
}

func (r *RectBuffer) Get() Rect {
	return Rect{r.GetX(), r.GetY(), r.GetW(), r.GetH()}
}

func (r *RectBuffer) GetX() int32 {
	return int32(C.MaaGetRectX(r.handle))
}

func (r *RectBuffer) GetY() int32 {
	return int32(C.MaaGetRectY(r.handle))
}

func (r *RectBuffer) GetW() int32 {
	return int32(C.MaaGetRectW(r.handle))
}

func (r *RectBuffer) GetH() int32 {
	return int32(C.MaaGetRectH(r.handle))
}

func (r *RectBuffer) Set(rect Rect) bool {
	return C.MaaSetRect(
		r.handle,
		C.int32_t(rect.X),
		C.int32_t(rect.Y),
		C.int32_t(rect.W),
		C.int32_t(rect.H),
	) != 0
}

func (r *RectBuffer) SetX(value int32) bool {
	return C.MaaSetRectX(
		r.handle,
		C.int32_t(value),
	) != 0
}

func (r *RectBuffer) SetY(value int32) bool {
	return C.MaaSetRectY(
		r.handle,
		C.int32_t(value),
	) != 0
}

func (r *RectBuffer) SetW(value int32) bool {
	return C.MaaSetRectW(
		r.handle,
		C.int32_t(value),
	) != 0
}

func (r *RectBuffer) SetH(value int32) bool {
	return C.MaaSetRectH(
		r.handle,
		C.int32_t(value),
	) != 0
}
