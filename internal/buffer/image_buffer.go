package buffer

import (
	"image"
	"image/color"
	"image/draw"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/native"
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
	rawData := i.getRawData()
	if rawData == nil {
		return nil
	}
	width := i.getWidth()
	height := i.getHeight()

	img := image.NewNRGBA(image.Rect(0, 0, int(width), int(height)))
	raw := unsafe.Slice((*byte)(rawData), width*height*3)
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

// Set converts an image.Image to raw data and sets it in the buffer.
func (i *ImageBuffer) Set(img image.Image) bool {
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

	return i.setRawData(unsafe.Pointer(&rawData[0]), int32(width), int32(height), imageType)
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

// GetEncoded retrieves the encoded image data (e.g., PNG/JPEG) from the buffer.
// It returns the encoded data as a byte slice.
func (i *ImageBuffer) GetEncoded() []byte {
	ptr := native.MaaImageBufferGetEncoded(i.handle)
	if ptr == nil {
		return nil
	}
	size := native.MaaImageBufferGetEncodedSize(i.handle)
	if size == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(ptr), size)
}

// GetEncodedSize retrieves the size of the encoded image data.
func (i *ImageBuffer) GetEncodedSize() uint64 {
	return native.MaaImageBufferGetEncodedSize(i.handle)
}

// SetEncoded sets the encoded image data (e.g., PNG/JPEG bytes) in the buffer.
// It takes the encoded data as a byte slice.
// It returns true if the operation was successful, otherwise false.
func (i *ImageBuffer) SetEncoded(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	return native.MaaImageBufferSetEncoded(i.handle, unsafe.Pointer(&data[0]), uint64(len(data)))
}
