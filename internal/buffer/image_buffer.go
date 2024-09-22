package buffer

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import (
	"image"
	"image/color"
	"image/draw"
	"unsafe"
)

type ImageBuffer struct {
	handle *C.MaaImageBuffer
}

func NewImageBuffer() *ImageBuffer {
	handle := C.MaaImageBufferCreate()
	if handle == nil {
		return nil
	}
	return &ImageBuffer{
		handle: handle,
	}
}

func NewImageBufferByHandle(handle unsafe.Pointer) *ImageBuffer {
	return &ImageBuffer{
		handle: (*C.MaaImageBuffer)(handle),
	}
}

func (i *ImageBuffer) Destroy() {
	C.MaaImageBufferDestroy(i.handle)
}

func (i *ImageBuffer) Handle() unsafe.Pointer {
	return unsafe.Pointer(i.handle)
}

func (i *ImageBuffer) IsEmpty() bool {
	return C.MaaImageBufferIsEmpty(i.handle) != 0
}

func (i *ImageBuffer) Clear() bool {
	return C.MaaImageBufferClear(i.handle) != 0
}

// GetByRawData retrieves the image from raw data stored in the buffer.
func (i *ImageBuffer) GetByRawData() image.Image {
	rawData := i.getRawData()
	if rawData == nil {
		return nil
	}
	width := i.getWidth()
	height := i.getHeight()

	img := image.NewNRGBA(image.Rect(0, 0, int(width), int(height)))
	raw := C.GoBytes(rawData, C.int(width*height*3))
	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			offset := (y*int(width) + x) * 3
			r := raw[offset+2]
			g := raw[offset+1]
			b := raw[offset]
			img.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

// SetRawData converts an image.Image to raw data and sets it in the buffer.
func (i *ImageBuffer) SetRawData(img image.Image) bool {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	imageType := int32(16) // CV_8UC3

	rawData := make([]byte, width*height*3)

	nrgbaImg, ok := img.(*image.NRGBA)
	if !ok {
		nrgbaImg = image.NewNRGBA(img.Bounds())
		draw.Draw(nrgbaImg, img.Bounds(), img, image.Point{}, draw.Src)
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := (y*width + x) * 4
			r := nrgbaImg.Pix[offset]
			g := nrgbaImg.Pix[offset+1]
			b := nrgbaImg.Pix[offset+2]
			rawData[(y*width+x)*3] = b
			rawData[(y*width+x)*3+1] = g
			rawData[(y*width+x)*3+2] = r
		}
	}

	cRawData := C.CBytes(rawData)
	defer C.free(cRawData)

	return i.setRawData(cRawData, int32(width), int32(height), imageType)
}

// getRawData retrieves the raw image data from the buffer.
// It returns a pointer to the raw image data.
func (i *ImageBuffer) getRawData() unsafe.Pointer {
	return unsafe.Pointer(C.MaaImageBufferGetRawData(i.handle))
}

// getWidth retrieves the width of the image stored in the buffer.
// It returns the width as an int32.
func (i *ImageBuffer) getWidth() int32 {
	return int32(C.MaaImageBufferWidth(i.handle))
}

// getHeight retrieves the height of the image stored in the buffer.
// It returns the height as an int32.
func (i *ImageBuffer) getHeight() int32 {
	return int32(C.MaaImageBufferHeight(i.handle))
}

// getType retrieves the type of the image stored in the buffer.
// This corresponds to the cv::Mat.type() in OpenCV.
// It returns the type as an int32.
func (i *ImageBuffer) getType() int32 {
	return int32(C.MaaImageBufferType(i.handle))
}

// setRawData sets the raw image data in the buffer.
// It takes a pointer to the raw image data, the width, height, and type of the image.
// It returns true if the operation was successful, otherwise false.
func (i *ImageBuffer) setRawData(data unsafe.Pointer, width, height, imageType int32) bool {
	return C.MaaImageBufferSetRawData(
		i.handle,
		C.MaaImageRawData(data),
		C.int32_t(width),
		C.int32_t(height),
		C.int32_t(imageType),
	) != 0
}
