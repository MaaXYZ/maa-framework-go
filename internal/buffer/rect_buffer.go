package buffer

import (
	"github.com/MaaXYZ/maa-framework-go/v3/internal/native"
)

// Rect represents a 2D rectangle area
// X, Y coordinates represent the top-left corner position
// W, H represent the width and height of the rectangle respectively
type Rect struct {
	X, Y, W, H int32
}

// ToInts converts the rectangle to an array of 4 int32 values
// Returns array in format: [X, Y, Width, Height]
func (r Rect) ToInts() [4]int32 {
	return [4]int32{r.X, r.Y, r.W, r.H}
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
	return Rect{r.GetX(), r.GetY(), r.GetW(), r.GetH()}
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
	return native.MaaRectSet(r.handle, rect.X, rect.Y, rect.W, rect.H)
}
