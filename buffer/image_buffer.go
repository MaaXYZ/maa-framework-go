package buffer

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>

typedef struct MaaImageBuffer* MaaImageBufferHandle;
*/
import "C"
import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"unsafe"
)

type ImageBuffer struct {
	handle C.MaaImageBufferHandle
}

func NewImageBuffer() *ImageBuffer {
	handle := C.MaaImageBufferCreate()
	return &ImageBuffer{
		handle: handle,
	}
}

func NewImageBufferByHandle(handle unsafe.Pointer) *ImageBuffer {
	return &ImageBuffer{
		handle: C.MaaImageBufferHandle(handle),
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
func (i *ImageBuffer) GetByRawData() (image.Image, error) {
	rawData := i.getRawData()
	if rawData == nil {
		return nil, errors.New("failed to get raw image data")
	}

	width := i.getWidth()
	height := i.getHeight()
	imageType := i.getType()

	var img image.Image

	if imageType != 16 { // CV_8UC3
		return nil, errors.New("unsupported image type, only CV_8UC3 is supported")
	}

	// Create a new NRGBA image
	nrgbaImg := image.NewNRGBA(image.Rect(0, 0, int(width), int(height)))
	raw := C.GoBytes(rawData, C.int(width*height*3))

	// Copy RGB data to NRGBA
	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			offset := (y*int(width) + x) * 3
			r := raw[offset+2]
			g := raw[offset+1]
			b := raw[offset]
			nrgbaImg.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}

	img = nrgbaImg

	return img, nil
}

// SetRawData converts an image.Image to raw data and sets it in the buffer.
func (i *ImageBuffer) SetRawData(img image.Image) error {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	imageType := int32(16) // CV_8UC3

	// Create a new raw data buffer
	rawData := make([]byte, width*height*3)

	// Convert image to NRGBA if it's not already
	nrgbaImg, ok := img.(*image.NRGBA)
	if !ok {
		nrgbaImg = image.NewNRGBA(img.Bounds())
		draw.Draw(nrgbaImg, img.Bounds(), img, image.Point{}, draw.Src)
	}

	// Copy NRGBA data to raw data buffer
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

	// Convert raw data to C pointer
	cRawData := C.CBytes(rawData)
	defer C.free(cRawData)

	// Set raw data in the buffer
	if !i.setRawData(cRawData, int32(width), int32(height), imageType) {
		return errors.New("failed to set raw image data")
	}

	return nil
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

// GetByEncoded retrieves the decoded image from the buffer.
// It returns the decoded image and an error if the operation was unsuccessful.
func (i *ImageBuffer) GetByEncoded() (image.Image, error) {
	encodedData := i.getEncoded()
	if encodedData == nil {
		return nil, errors.New("failed to get encoded image data")
	}
	dataSize := i.getEncodedSize()
	if dataSize == 0 {
		return nil, errors.New("encoded image size is zero")
	}

	data := C.GoBytes(encodedData, C.int32_t(dataSize))
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return img, nil
}

// SetEncoded encodes the given image and sets it in the buffer.
// It takes an image.Image as input and returns an error if the operation was unsuccessful.
func (i *ImageBuffer) SetEncoded(img image.Image) error {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}

	data := buf.Bytes()
	cData := C.CBytes(data)
	defer C.free(cData)

	if !i.setEncoded(cData, uint64(len(data))) {
		return errors.New("failed to set encoded image data")
	}
	return nil
}

// getEncoded retrieves the encoded image data from the buffer.
// It returns a pointer to the encoded image data.
func (i *ImageBuffer) getEncoded() unsafe.Pointer {
	return unsafe.Pointer(C.MaaImageBufferGetEncoded(i.handle))
}

// getEncodedSize retrieves the size of the encoded image data in the buffer.
// It returns the size of the encoded image data as an integer.
func (i *ImageBuffer) getEncodedSize() int32 {
	return int32(C.MaaImageBufferGetEncodedSize(i.handle))
}

// setEncoded sets the encoded image data in the buffer.
// It takes a pointer to the encoded image data and the size of the data.
// It returns true if the operation was successful, otherwise false.
func (i *ImageBuffer) setEncoded(data unsafe.Pointer, size uint64) bool {
	return C.MaaImageBufferSetEncoded(
		i.handle,
		C.MaaImageEncodedData(data),
		C.uint64_t(size),
	) != 0
}
