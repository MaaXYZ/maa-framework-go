package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import "unsafe"

type RectBuffer interface {
	Destroy()
	Handle() unsafe.Pointer
	GetX() int32
	GetY() int32
	GetW() int32
	GetH() int32
	Set(x, y, w, h int32) bool
	setX(value int32) bool
	SetY(value int32) bool
	SetW(value int32) bool
	SetH(value int32) bool
}

type rectBuffer struct {
	handle C.MaaRectHandle
}

func NewRectBuffer() RectBuffer {
	handle := C.MaaCreateRectBuffer()
	return &rectBuffer{handle: handle}
}

func (r *rectBuffer) Destroy() {
	C.MaaDestroyRectBuffer(r.handle)
}

func (r *rectBuffer) Handle() unsafe.Pointer {
	return unsafe.Pointer(r.handle)
}

func (r *rectBuffer) GetX() int32 {
	return int32(C.MaaGetRectX(r.handle))
}

func (r *rectBuffer) GetY() int32 {
	return int32(C.MaaGetRectY(r.handle))
}

func (r *rectBuffer) GetW() int32 {
	return int32(C.MaaGetRectW(r.handle))
}

func (r *rectBuffer) GetH() int32 {
	return int32(C.MaaGetRectH(r.handle))
}

func (r *rectBuffer) Set(x, y, w, h int32) bool {
	return C.MaaSetRect(r.handle, C.int32_t(x), C.int32_t(y), C.int32_t(w), C.int32_t(h)) != 0
}

func (r *rectBuffer) setX(value int32) bool {
	return C.MaaSetRectX(r.handle, C.int32_t(value)) != 0
}

func (r *rectBuffer) SetY(value int32) bool {
	return C.MaaSetRectY(r.handle, C.int32_t(value)) != 0
}

func (r *rectBuffer) SetW(value int32) bool {
	return C.MaaSetRectW(r.handle, C.int32_t(value)) != 0
}

func (r *rectBuffer) SetH(value int32) bool {
	return C.MaaSetRectH(r.handle, C.int32_t(value)) != 0
}
