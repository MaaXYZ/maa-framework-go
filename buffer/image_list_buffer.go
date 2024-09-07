package buffer

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import (
	"image"
	"unsafe"
)

type ImageListBuffer struct {
	handle *C.MaaImageListBuffer
}

func NewImageListBuffer() *ImageListBuffer {
	handle := C.MaaImageListBufferCreate()
	return &ImageListBuffer{
		handle: handle,
	}
}

func NewImageListBufferByHandle(handle unsafe.Pointer) *ImageListBuffer {
	return &ImageListBuffer{
		handle: (*C.MaaImageListBuffer)(handle),
	}
}

func (il *ImageListBuffer) Destroy() {
	C.MaaImageListBufferDestroy(il.handle)
}

func (il *ImageListBuffer) Handle() unsafe.Pointer {
	return unsafe.Pointer(il.handle)
}

func (il *ImageListBuffer) IsEmpty() bool {
	return C.MaaImageListBufferIsEmpty(il.handle) != 0
}

func (il *ImageListBuffer) Clear() bool {
	return C.MaaImageListBufferClear(il.handle) != 0
}

func (il *ImageListBuffer) Size() uint64 {
	return uint64(C.MaaImageListBufferSize(il.handle))
}

func (il *ImageListBuffer) Get(index uint64) (image.Image, error) {
	handle := C.MaaImageListBufferAt(il.handle, C.uint64_t(index))
	img := &ImageBuffer{
		handle: handle,
	}
	return img.GetByRawData()
}

func (il *ImageListBuffer) GetAll() ([]image.Image, error) {
	size := il.Size()
	images := make([]image.Image, size)
	for i := uint64(0); i < size; i++ {
		img, err := il.Get(i)
		if err != nil {
			return nil, err
		}
		images[i] = img
	}
	return images, nil
}

func (il *ImageListBuffer) Append(value ImageBuffer) bool {
	return C.MaaImageListBufferAppend(
		il.handle,
		(*C.MaaImageBuffer)(value.Handle()),
	) != 0
}

func (il *ImageListBuffer) Remove(index uint64) bool {
	return C.MaaImageListBufferRemove(
		il.handle,
		C.uint64_t(index),
	) != 0
}
