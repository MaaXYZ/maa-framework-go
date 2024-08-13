package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import (
	"image"
	"unsafe"
)

type ImageListBuffer interface {
	Destroy()
	Handle() unsafe.Pointer
	IsEmpty() bool
	Clear() bool
	Size() uint64
	Get(index uint64) (image.Image, error)
	GetAll() ([]image.Image, error)
	Append(value ImageBuffer) bool
	Remove(index uint64) bool
}

type imageListBuffer struct {
	handle C.MaaImageListBufferHandle
}

func NewImageListBuffer() ImageListBuffer {
	handle := C.MaaCreateImageListBuffer()
	return &imageListBuffer{handle: handle}
}

func (il *imageListBuffer) Destroy() {
	C.MaaDestroyImageListBuffer(il.handle)
}

func (il *imageListBuffer) Handle() unsafe.Pointer {
	return unsafe.Pointer(il.handle)
}

func (il *imageListBuffer) IsEmpty() bool {
	return C.MaaIsImageListEmpty(il.handle) != 0
}

func (il *imageListBuffer) Clear() bool {
	return C.MaaClearImageList(il.handle) != 0
}

func (il *imageListBuffer) Size() uint64 {
	return uint64(C.MaaGetImageListSize(il.handle))
}

func (il *imageListBuffer) Get(index uint64) (image.Image, error) {
	handle := C.MaaGetImageListAt(il.handle, C.uint64_t(index))
	img := &imageBuffer{handle: handle}
	return img.GetByRawData()
}

func (il *imageListBuffer) GetAll() ([]image.Image, error) {
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

func (il *imageListBuffer) Append(value ImageBuffer) bool {
	return C.MaaImageListAppend(il.handle, C.MaaImageBufferHandle(value.Handle())) != 0
}

func (il *imageListBuffer) Remove(index uint64) bool {
	return C.MaaImageListRemove(il.handle, C.uint64_t(index)) != 0
}
