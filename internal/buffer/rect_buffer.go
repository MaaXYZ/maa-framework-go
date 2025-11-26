package buffer

import (
	"github.com/MaaXYZ/maa-framework-go/v3/internal/native"
)

// Rect represents a 2D rectangle area
type Rect [4]int

func (r Rect) X() int {
	return r[0]
}

func (r Rect) Y() int {
	return r[1]
}

func (r Rect) Width() int {
	return r[2]
}

func (r Rect) Height() int {
	return r[3]
}

type RectBuffer struct {
	handle uintptr
}

func NewRectBuffer() *RectBuffer {
	handle := native.MaaRectCreate()
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
	native.MaaRectDestroy(r.handle)
}

func (r *RectBuffer) Handle() uintptr {
	return r.handle
}

func (r *RectBuffer) Get() Rect {
	return Rect{int(r.GetX()), int(r.GetY()), int(r.GetW()), int(r.GetH())}
}

func (r *RectBuffer) GetX() int32 {
	return native.MaaRectGetX(r.handle)
}

func (r *RectBuffer) GetY() int32 {
	return native.MaaRectGetY(r.handle)
}

func (r *RectBuffer) GetW() int32 {
	return native.MaaRectGetW(r.handle)
}

func (r *RectBuffer) GetH() int32 {
	return native.MaaRectGetH(r.handle)
}

func (r *RectBuffer) Set(rect Rect) bool {
	return native.MaaRectSet(r.handle, int32(rect.X()), int32(rect.Y()), int32(rect.Width()), int32(rect.Height()))
}
