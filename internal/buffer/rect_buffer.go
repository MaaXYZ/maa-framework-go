package buffer

import (
	"errors"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/rect"
)

type RectBuffer struct {
	handle uintptr
}

// NewRectBuffer creates a new rect buffer.
// It returns an error when the underlying native buffer cannot be created.
func NewRectBuffer() (*RectBuffer, error) {
	handle := native.MaaRectCreate()
	if handle == 0 {
		return nil, errors.New("failed to create rect buffer")
	}
	return &RectBuffer{
		handle: handle,
	}, nil
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

func (r *RectBuffer) Get() rect.Rect {
	return rect.Rect{int(r.GetX()), int(r.GetY()), int(r.GetW()), int(r.GetH())}
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

// Set writes rect into the buffer.
// It returns an error when the underlying native write fails.
func (r *RectBuffer) Set(rect rect.Rect) error {
	if !native.MaaRectSet(r.handle, int32(rect.X()), int32(rect.Y()), int32(rect.Width()), int32(rect.Height())) {
		return errors.New("failed to set rect")
	}
	return nil
}
