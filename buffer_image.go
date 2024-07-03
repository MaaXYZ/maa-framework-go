package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import "unsafe"

type ImageBuffer interface {
	Destroy()
	Handle() unsafe.Pointer
	IsEmpty() bool
	Clear() bool
	GetRawData() unsafe.Pointer
	GetWidth() int32
	GetHeight() int32
	GetType() int32
	SetRawData(data unsafe.Pointer, width, height, imageType int32) bool
	GetEncoded() unsafe.Pointer
	GetEncodedSize() int32
	SetEncoded(data unsafe.Pointer, size uint64) bool
}

type imageBuffer struct {
	handle C.MaaImageBufferHandle
}

func NewImageBuffer() ImageBuffer {
	handle := C.MaaCreateImageBuffer()
	return &imageBuffer{handle: handle}
}

func (i *imageBuffer) Destroy() {
	C.MaaDestroyImageBuffer(i.handle)
}

func (i *imageBuffer) Handle() unsafe.Pointer {
	return unsafe.Pointer(i.handle)
}

func (i *imageBuffer) IsEmpty() bool {
	return C.MaaIsImageEmpty(i.handle) != 0
}

func (i *imageBuffer) Clear() bool {
	return C.MaaClearImage(i.handle) != 0
}

func (i *imageBuffer) GetRawData() unsafe.Pointer {
	return unsafe.Pointer(C.MaaGetImageRawData(i.handle))
}

func (i *imageBuffer) GetWidth() int32 {
	return int32(C.MaaGetImageWidth(i.handle))
}

func (i *imageBuffer) GetHeight() int32 {
	return int32(C.MaaGetImageHeight(i.handle))
}

// GetType return cv::Mat.type()
func (i *imageBuffer) GetType() int32 {
	return int32(C.MaaGetImageType(i.handle))
}

func (i *imageBuffer) SetRawData(data unsafe.Pointer, width, height, imageType int32) bool {
	return C.MaaSetImageRawData(i.handle, C.MaaImageRawData(data), C.int32_t(width), C.int32_t(height), C.int32_t(imageType)) != 0
}

func (i *imageBuffer) GetEncoded() unsafe.Pointer {
	return unsafe.Pointer(C.MaaGetImageEncoded(i.handle))
}

func (i *imageBuffer) GetEncodedSize() int32 {
	return int32(C.MaaGetImageEncodedSize(i.handle))
}

func (i *imageBuffer) SetEncoded(data unsafe.Pointer, size uint64) bool {
	return C.MaaSetImageEncoded(i.handle, C.MaaImageEncodedData(data), C.uint64_t(size)) != 0
}
