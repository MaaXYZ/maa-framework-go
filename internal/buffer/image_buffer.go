package buffer

import (
	"image"
	"image/draw"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
)

type ImageBuffer struct {
	handle uintptr
}

func NewImageBuffer() *ImageBuffer {
	handle := native.MaaImageBufferCreate()
	if handle == 0 {
		return nil
	}
	return &ImageBuffer{
		handle: handle,
	}
}

func NewImageBufferByHandle(handle uintptr) *ImageBuffer {
	return &ImageBuffer{
		handle: handle,
	}
}

func (i *ImageBuffer) Destroy() {
	native.MaaImageBufferDestroy(i.handle)
}

func (i *ImageBuffer) Handle() uintptr {
	return i.handle
}

func (i *ImageBuffer) IsEmpty() bool {
	return native.MaaImageBufferIsEmpty(i.handle)
}

func (i *ImageBuffer) Clear() bool {
	return native.MaaImageBufferClear(i.handle)
}

// Get retrieves the image from raw data stored in the buffer.
func (i *ImageBuffer) Get() image.Image {
	return i.GetInto(nil)
}

// GetInto retrieves the image from raw data stored in the buffer and writes into dst when possible.
// If dst is nil or size mismatched, a new *image.NRGBA is allocated and returned.
func (i *ImageBuffer) GetInto(dst *image.NRGBA) *image.NRGBA {
	rawData := i.getRawData()
	if rawData == nil {
		return nil
	}
	width := int(i.getWidth())
	height := int(i.getHeight())

	if dst == nil || dst.Rect.Dx() != width || dst.Rect.Dy() != height {
		dst = image.NewNRGBA(image.Rect(0, 0, width, height))
	}

	raw := unsafe.Slice((*byte)(rawData), width*height*3)
	pix := dst.Pix
	if dst.Stride == width*4 {
		for src, dstIdx := 0, 0; src < len(raw); src, dstIdx = src+3, dstIdx+4 {
			// Native buffer stores pixels as BGR, convert to NRGBA (alpha fixed to 255).
			pix[dstIdx] = raw[src+2]
			pix[dstIdx+1] = raw[src+1]
			pix[dstIdx+2] = raw[src]
			pix[dstIdx+3] = 255
		}
		return dst
	}

	for y := 0; y < height; y++ {
		srcIdx := y * width * 3
		dstIdx := y * dst.Stride
		for x := 0; x < width; x++ {
			pix[dstIdx] = raw[srcIdx+2]
			pix[dstIdx+1] = raw[srcIdx+1]
			pix[dstIdx+2] = raw[srcIdx]
			pix[dstIdx+3] = 255
			srcIdx += 3
			dstIdx += 4
		}
	}
	return dst
}

// Set converts an image.Image to raw data and sets it in the buffer.
func (i *ImageBuffer) Set(img image.Image) bool {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	imageType := int32(16) // CV_8UC3

	rawData := make([]byte, width*height*3)

	nrgbaImg, ok := img.(*image.NRGBA)
	if !ok {
		nrgbaImg = image.NewNRGBA(image.Rect(0, 0, width, height))
		draw.Draw(nrgbaImg, nrgbaImg.Bounds(), img, bounds.Min, draw.Src)
	}
	encodeNRGBAToBGR(nrgbaImg, rawData, width, height)

	return i.setRawData(unsafe.Pointer(&rawData[0]), int32(width), int32(height), imageType)
}

func encodeNRGBAToBGR(src *image.NRGBA, dst []byte, width, height int) {
	if width == 0 || height == 0 {
		return
	}

	pix := src.Pix
	if src.Stride == width*4 {
		for srcIdx, dstIdx := 0, 0; dstIdx < len(dst); srcIdx, dstIdx = srcIdx+4, dstIdx+3 {
			dst[dstIdx] = pix[srcIdx+2]
			dst[dstIdx+1] = pix[srcIdx+1]
			dst[dstIdx+2] = pix[srcIdx]
		}
		return
	}

	for y := 0; y < height; y++ {
		srcIdx := y * src.Stride
		dstIdx := y * width * 3
		rowEnd := srcIdx + width*4
		for srcIdx < rowEnd {
			dst[dstIdx] = pix[srcIdx+2]
			dst[dstIdx+1] = pix[srcIdx+1]
			dst[dstIdx+2] = pix[srcIdx]
			srcIdx += 4
			dstIdx += 3
		}
	}
}

// getRawData retrieves the raw image data from the buffer.
// It returns a pointer to the raw image data.
func (i *ImageBuffer) getRawData() unsafe.Pointer {
	return native.MaaImageBufferGetRawData(i.handle)
}

// getWidth retrieves the width of the image stored in the buffer.
// It returns the width as an int32.
func (i *ImageBuffer) getWidth() int32 {
	return native.MaaImageBufferWidth(i.handle)
}

// getHeight retrieves the height of the image stored in the buffer.
// It returns the height as an int32.
func (i *ImageBuffer) getHeight() int32 {
	return native.MaaImageBufferHeight(i.handle)
}

// getType retrieves the type of the image stored in the buffer.
// This corresponds to the cv::Mat.type() in OpenCV.
// It returns the type as an int32.
func (i *ImageBuffer) getType() int32 {
	return native.MaaImageBufferType(i.handle)
}

// setRawData sets the raw image data in the buffer.
// It takes a pointer to the raw image data, the width, height, and type of the image.
// It returns true if the operation was successful, otherwise false.
func (i *ImageBuffer) setRawData(data unsafe.Pointer, width, height, imageType int32) bool {
	return native.MaaImageBufferSetRawData(i.handle, data, width, height, imageType)
}

// Resize resizes the image buffer to the specified width and height.
// It returns true if the operation was successful, otherwise false.
func (i *ImageBuffer) Resize(width, height int32) bool {
	return native.MaaImageBufferResize(i.handle, width, height)
}

// NOTE: GetEncoded and SetEncoded are intentionally NOT implemented in Go binding.
// Go handles image encoding/decoding natively through the standard library (image/png, image/jpeg, etc.).
// Do not add encoded image methods here.
