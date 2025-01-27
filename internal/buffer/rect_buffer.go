package buffer

import (
	"github.com/MaaXYZ/maa-framework-go/v2/internal/maa"
)

type Rect struct {
	X, Y, W, H int32
}

func (r Rect) ToInts() [4]int32 {
	return [4]int32{r.X, r.Y, r.W, r.H}
}

type RectBuffer struct {
	handle uintptr
}

func NewRectBuffer() *RectBuffer {
	handle := maa.MaaRectCreate()
	if handle == 0 {
		return nil
	}
	return &RectBuffer{
		handle: handle,
	}
}

func NewRectBufferByHandle(handle uintptr) *RectBuffer {
	return &RectBuffer{
		handle: handle,
	}
}

func (r *RectBuffer) Destroy() {
	maa.MaaRectDestroy(r.handle)
}

func (r *RectBuffer) Handle() uintptr {
	return r.handle
}

func (r *RectBuffer) Get() Rect {
	return Rect{r.GetX(), r.GetY(), r.GetW(), r.GetH()}
}

func (r *RectBuffer) GetX() int32 {
	return maa.MaaRectGetX(r.handle)
}

func (r *RectBuffer) GetY() int32 {
	return maa.MaaRectGetY(r.handle)
}

func (r *RectBuffer) GetW() int32 {
	return maa.MaaRectGetW(r.handle)
}

func (r *RectBuffer) GetH() int32 {
	return maa.MaaRectGetH(r.handle)
}

func (r *RectBuffer) Set(rect Rect) bool {
	return maa.MaaRectSet(r.handle, rect.X, rect.Y, rect.W, rect.H)
}
