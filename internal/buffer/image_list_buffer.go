package buffer

import (
	"image"

	"github.com/MaaXYZ/maa-framework-go/v2/internal/maa"
)

type ImageListBuffer struct {
	handle uintptr
}

func NewImageListBuffer() *ImageListBuffer {
	handle := maa.MaaImageListBufferCreate()
	if handle == 0 {
		return nil
	}
	return &ImageListBuffer{
		handle: handle,
	}
}

func NewImageListBufferByHandle(handle uintptr) *ImageListBuffer {
	return &ImageListBuffer{
		handle: handle,
	}
}

func (il *ImageListBuffer) Destroy() {
	maa.MaaImageListBufferDestroy(il.handle)
}

func (il *ImageListBuffer) Handle() uintptr {
	return il.handle
}

func (il *ImageListBuffer) IsEmpty() bool {
	return maa.MaaImageListBufferIsEmpty(il.handle)
}

func (il *ImageListBuffer) Clear() bool {
	return maa.MaaImageListBufferClear(il.handle)
}

func (il *ImageListBuffer) Size() uint64 {
	return maa.MaaImageListBufferSize(il.handle)
}

func (il *ImageListBuffer) Get(index uint64) image.Image {
	handle := maa.MaaImageListBufferAt(il.handle, index)
	img := &ImageBuffer{
		handle: handle,
	}
	return img.Get()
}

func (il *ImageListBuffer) GetAll() []image.Image {
	size := il.Size()
	images := make([]image.Image, size)
	for i := uint64(0); i < size; i++ {
		img := il.Get(i)
		images[i] = img
	}
	return images
}

func (il *ImageListBuffer) Append(value *ImageBuffer) bool {
	return maa.MaaImageListBufferAppend(il.handle, value.handle)
}

func (il *ImageListBuffer) Remove(index uint64) bool {
	return maa.MaaImageListBufferRemove(il.handle, index)
}
