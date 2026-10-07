package buffer

import (
	"errors"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/rect"
)

// RectBuffer wraps a native rect buffer holding x, y, width, and height as
// int32 values. Create one with NewRectBuffer, or borrow an existing handle
// with NewRectBufferByHandle.
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

// NewRectBufferByHandle wraps an existing native rect buffer handle without
// taking ownership. The caller keeps owning the handle and stays responsible
// for destroying it exactly once; the wrapper is a shared view of the same
// buffer.
func NewRectBufferByHandle(handle uintptr) *RectBuffer {
	return &RectBuffer{
		handle: handle,
	}
}

// Destroy releases the underlying native rect buffer. The owner must call it
// exactly once and not use the wrapper afterwards; a borrowed wrapper must
// not destroy the shared handle.
func (r *RectBuffer) Destroy() {
	native.MaaRectDestroy(r.handle)
}

// Handle returns the underlying native rect buffer handle for passing to
// framework functions that take a rect buffer.
func (r *RectBuffer) Handle() uintptr {
	return r.handle
}

// Get returns the rect currently stored in the buffer.
func (r *RectBuffer) Get() rect.Rect {
	return rect.Rect{int(r.GetX()), int(r.GetY()), int(r.GetW()), int(r.GetH())}
}

// GetX returns the x component of the stored rect.
func (r *RectBuffer) GetX() int32 {
	return native.MaaRectGetX(r.handle)
}

// GetY returns the y component of the stored rect.
func (r *RectBuffer) GetY() int32 {
	return native.MaaRectGetY(r.handle)
}

// GetW returns the width component of the stored rect.
func (r *RectBuffer) GetW() int32 {
	return native.MaaRectGetW(r.handle)
}

// GetH returns the height component of the stored rect.
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
