package buffer

import (
	"image"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
)

// ImageListBuffer wraps a native list of image buffers. Create one with
// NewImageListBuffer, or borrow an existing handle with
// NewImageListBufferByHandle.
type ImageListBuffer struct {
	handle uintptr
}

// NewImageListBuffer creates a new empty image list buffer.
// It returns nil when the underlying native buffer cannot be created.
func NewImageListBuffer() *ImageListBuffer {
	handle := native.MaaImageListBufferCreate()
	if handle == 0 {
		return nil
	}
	return &ImageListBuffer{
		handle: handle,
	}
}

// NewImageListBufferByHandle wraps an existing native image list buffer
// handle without taking ownership. The caller keeps owning the handle and
// stays responsible for destroying it exactly once; the wrapper is a shared
// view of the same list.
func NewImageListBufferByHandle(handle uintptr) *ImageListBuffer {
	return &ImageListBuffer{
		handle: handle,
	}
}

// Destroy releases the underlying native list buffer. The owner must call it
// exactly once and not use the wrapper afterwards; a borrowed wrapper must
// not destroy the shared handle.
func (il *ImageListBuffer) Destroy() {
	native.MaaImageListBufferDestroy(il.handle)
}

// Handle returns the underlying native list buffer handle for passing to
// framework functions that take an image list buffer.
func (il *ImageListBuffer) Handle() uintptr {
	return il.handle
}

// IsEmpty reports whether the list holds no images.
func (il *ImageListBuffer) IsEmpty() bool {
	return native.MaaImageListBufferIsEmpty(il.handle)
}

// Clear removes all images and reports whether the native call succeeded.
func (il *ImageListBuffer) Clear() bool {
	return native.MaaImageListBufferClear(il.handle)
}

// Size returns the number of images in the list.
func (il *ImageListBuffer) Size() uint64 {
	return native.MaaImageListBufferSize(il.handle)
}

// Get returns a copy of the image at index. It returns a nil interface when
// index is out of range or the element holds no image.
func (il *ImageListBuffer) Get(index uint64) image.Image {
	handle := native.MaaImageListBufferAt(il.handle, index)
	img := &ImageBuffer{
		handle: handle,
	}
	return img.Get()
}

// GetAll returns copies of all images in list order. It returns a non-nil
// empty slice when the list is empty.
func (il *ImageListBuffer) GetAll() []image.Image {
	size := il.Size()
	images := make([]image.Image, size)
	for i := uint64(0); i < size; i++ {
		img := il.Get(i)
		images[i] = img
	}
	return images
}

// Append stores a deep copy of value and reports whether the native call
// succeeded; the caller keeps owning value. It returns false when value is
// nil.
func (il *ImageListBuffer) Append(value *ImageBuffer) bool {
	// Upstream returns false for a null value instead of crashing.
	if value == nil {
		return false
	}
	return native.MaaImageListBufferAppend(il.handle, value.handle)
}

// Remove deletes the image at index and reports whether index was in range.
func (il *ImageListBuffer) Remove(index uint64) bool {
	return native.MaaImageListBufferRemove(il.handle, index)
}
