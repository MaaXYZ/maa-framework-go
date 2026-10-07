// Package buffer provides Go wrappers for the MaaFramework buffer C API
// (MaaBuffer.h): image, image-list, string, string-list, and rect buffers
// exchanged with framework functions by handle.
//
// Each buffer is owned by the wrapper that created it and must be destroyed
// exactly once through Destroy. The ByHandle constructors instead borrow a
// handle owned elsewhere (typically a framework callback argument): the
// wrapper is a shared view, and the owner stays responsible for destroying.
// List buffers return copies from their element accessors, keep appended
// elements valid independently of the source value, and surface an
// out-of-range index as a zero value ("" or nil) and false from Remove,
// matching the native side. Image buffers hold raw
// BGR pixels in OpenCV's CV_8UC3 layout: Get decodes them into a fresh
// opaque image.RGBA and Set encodes an image.Image back, while the native
// PNG encoding API is intentionally unbound because Go handles image files
// natively. String buffers copy text through NUL-terminated C strings: Set
// truncates at the first NUL, SetWithSize copies an exact byte count and
// preserves embedded NULs, and Get truncates embedded NULs on the Go read
// side.
package buffer

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"reflect"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
)

// ImageBuffer wraps a native image buffer holding raw BGR pixels. Create one
// with NewImageBuffer, or borrow an existing handle with NewImageBufferByHandle.
type ImageBuffer struct {
	handle uintptr
}

const (
	cvType8UC3   int32 = 16
	bytesPerBGR        = 3
	bytesPerRGBA       = 4
	opaqueAlpha  byte  = 0xff
)

// NewImageBuffer creates a new image buffer.
// It returns an error when the underlying native buffer cannot be created.
func NewImageBuffer() (*ImageBuffer, error) {
	handle := native.MaaImageBufferCreate()
	if handle == 0 {
		return nil, errors.New("failed to create image buffer")
	}
	return &ImageBuffer{
		handle: handle,
	}, nil
}

// NewImageBufferByHandle wraps an existing native image buffer handle without
// taking ownership. The caller keeps owning the handle and stays responsible
// for destroying it exactly once; the wrapper is a shared view of the same
// buffer.
func NewImageBufferByHandle(handle uintptr) *ImageBuffer {
	return &ImageBuffer{
		handle: handle,
	}
}

// Destroy releases the underlying native buffer. The owner must call it
// exactly once and not use the wrapper afterwards; a borrowed wrapper must
// not destroy the shared handle.
func (i *ImageBuffer) Destroy() {
	native.MaaImageBufferDestroy(i.handle)
}

// Handle returns the underlying native buffer handle for passing to framework
// functions that take an image buffer.
func (i *ImageBuffer) Handle() uintptr {
	return i.handle
}

// IsEmpty reports whether the buffer holds no image.
func (i *ImageBuffer) IsEmpty() bool {
	return native.MaaImageBufferIsEmpty(i.handle)
}

// Clear releases the stored image and reports whether the native call
// succeeded.
func (i *ImageBuffer) Clear() bool {
	return native.MaaImageBufferClear(i.handle)
}

// Get retrieves the image from raw data stored in the buffer, interpreted as
// the CV_8UC3 BGR layout Set produces. It returns a fresh opaque *image.RGBA
// copy each call (alpha fixed to 255) and never shares memory with the
// buffer; it returns a nil interface when the buffer has no raw image data.
func (i *ImageBuffer) Get() image.Image {
	img := i.GetInto(nil)
	if img == nil {
		return nil
	}
	return img
}

// GetInto retrieves the image from raw data stored in the buffer and writes
// into dst when possible, interpreted as the CV_8UC3 BGR layout Set produces.
// The pixels are copied into dst, which is reallocated when dst is nil or
// sized for a different image; it returns nil and leaves dst unchanged when
// the buffer has no raw image data.
func (i *ImageBuffer) GetInto(dst *image.RGBA) *image.RGBA {
	rawData := i.getRawData()
	if rawData == nil {
		return nil
	}
	width := int(i.getWidth())
	height := int(i.getHeight())

	dst = ensureRGBA(dst, width, height)
	raw := unsafe.Slice((*byte)(rawData), width*height*bytesPerBGR)
	decodeBGRToRGBA(raw, dst, width, height)
	return dst
}

func ensureRGBA(dst *image.RGBA, width, height int) *image.RGBA {
	if dst == nil || dst.Rect.Dx() != width || dst.Rect.Dy() != height {
		return image.NewRGBA(image.Rect(0, 0, width, height))
	}
	return dst
}

func decodeBGRToRGBA(src []byte, dst *image.RGBA, width, height int) {
	srcRowBytes := width * bytesPerBGR
	dstRowBytes := width * bytesPerRGBA

	if dst.Stride == dstRowBytes {
		decodeBGRRowToRGBA(dst.Pix[:dstRowBytes*height], src)
		return
	}

	for y := 0; y < height; y++ {
		srcStart := y * srcRowBytes
		dstStart := y * dst.Stride
		decodeBGRRowToRGBA(dst.Pix[dstStart:dstStart+dstRowBytes], src[srcStart:srcStart+srcRowBytes])
	}
}

func decodeBGRRowToRGBA(dst, src []byte) {
	for srcIdx, dstIdx := 0, 0; srcIdx < len(src); srcIdx, dstIdx = srcIdx+bytesPerBGR, dstIdx+bytesPerRGBA {
		// Native buffer stores pixels as BGR, convert to RGBA (alpha fixed to 255).
		dst[dstIdx] = src[srcIdx+2]
		dst[dstIdx+1] = src[srcIdx+1]
		dst[dstIdx+2] = src[srcIdx]
		dst[dstIdx+3] = opaqueAlpha
	}
}

// isNilImage reports whether img is nil or a nil pointer stored in a
// non-nil interface, which would panic on Bounds.
func isNilImage(img image.Image) bool {
	if img == nil {
		return true
	}
	rv := reflect.ValueOf(img)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

// Set converts an image.Image to raw data and sets it in the buffer. Alpha is
// dropped: non-opaque RGBA pixels are exactly unpremultiplied before the BGR
// conversion, and the stored data has no alpha channel. The buffer is left
// unchanged and an error is returned when img is nil, has a non-positive
// dimension, or the underlying write fails.
func (i *ImageBuffer) Set(img image.Image) error {
	if isNilImage(img) {
		return errors.New("image is nil")
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return fmt.Errorf("image has invalid bounds: %dx%d", width, height)
	}

	rawData := make([]byte, width*height*bytesPerBGR)
	encodeImageToBGR(img, bounds, width, height, rawData)
	if !i.setRawData(unsafe.Pointer(&rawData[0]), int32(width), int32(height), cvType8UC3) {
		return errors.New("failed to set image raw data")
	}
	return nil
}

func encodeImageToBGR(img image.Image, bounds image.Rectangle, width, height int, dst []byte) {
	switch sourceImg := img.(type) {
	case *image.NRGBA:
		encodeNRGBAToBGR(sourceImg, dst, width, height)
	case *image.RGBA:
		encodeRGBAToBGR(sourceImg, dst, width, height)
	default:
		nrgbaImg := image.NewNRGBA(image.Rect(0, 0, width, height))
		draw.Draw(nrgbaImg, nrgbaImg.Bounds(), img, bounds.Min, draw.Src)
		encodeNRGBAToBGR(nrgbaImg, dst, width, height)
	}
}

func encodeNRGBAToBGR(src *image.NRGBA, dst []byte, width, height int) {
	encodeRGBABytesToBGR(src.Pix, src.Stride, dst, width, height)
}

func encodeRGBABytesToBGR(srcPix []byte, srcStride int, dst []byte, width, height int) {
	if width == 0 || height == 0 {
		return
	}

	srcRowBytes := width * bytesPerRGBA
	dstRowBytes := width * bytesPerBGR

	if srcStride == srcRowBytes {
		encodeRGBARowToBGR(srcPix[:srcRowBytes*height], dst)
		return
	}

	for y := 0; y < height; y++ {
		srcStart := y * srcStride
		dstStart := y * dstRowBytes
		encodeRGBARowToBGR(srcPix[srcStart:srcStart+srcRowBytes], dst[dstStart:dstStart+dstRowBytes])
	}
}

func encodeRGBARowToBGR(src, dst []byte) {
	for srcIdx, dstIdx := 0, 0; dstIdx < len(dst); srcIdx, dstIdx = srcIdx+bytesPerRGBA, dstIdx+bytesPerBGR {
		dst[dstIdx] = src[srcIdx+2]
		dst[dstIdx+1] = src[srcIdx+1]
		dst[dstIdx+2] = src[srcIdx]
	}
}

func encodeRGBAToBGR(src *image.RGBA, dst []byte, width, height int) {
	if width == 0 || height == 0 {
		return
	}

	if rgbaIsOpaque(src, width, height) {
		encodeOpaqueRGBAToBGR(src, dst, width, height)
		return
	}

	encodeUnpremultipliedRGBAToBGR(src, dst, width, height)
}

func rgbaIsOpaque(src *image.RGBA, width, height int) bool {
	return rgbaBytesAreOpaque(src.Pix, src.Stride, width, height)
}

func rgbaBytesAreOpaque(srcPix []byte, srcStride, width, height int) bool {
	srcRowBytes := width * bytesPerRGBA
	if srcStride == srcRowBytes {
		return rgbaRowIsOpaque(srcPix[:srcRowBytes*height])
	}

	for y := 0; y < height; y++ {
		srcStart := y * srcStride
		if !rgbaRowIsOpaque(srcPix[srcStart : srcStart+srcRowBytes]) {
			return false
		}
	}
	return true
}

func rgbaRowIsOpaque(row []byte) bool {
	for alphaIdx := 3; alphaIdx < len(row); alphaIdx += bytesPerRGBA {
		if row[alphaIdx] != opaqueAlpha {
			return false
		}
	}
	return true
}

func encodeOpaqueRGBAToBGR(src *image.RGBA, dst []byte, width, height int) {
	encodeRGBABytesToBGR(src.Pix, src.Stride, dst, width, height)
}

func encodeUnpremultipliedRGBAToBGR(src *image.RGBA, dst []byte, width, height int) {
	encodeUnpremultipliedRGBABytesToBGR(src.Pix, src.Stride, dst, width, height)
}

func encodeUnpremultipliedRGBABytesToBGR(srcPix []byte, srcStride int, dst []byte, width, height int) {
	if width == 0 || height == 0 {
		return
	}

	srcRowBytes := width * bytesPerRGBA
	dstRowBytes := width * bytesPerBGR

	if srcStride == srcRowBytes {
		encodeUnpremultipliedRGBARowToBGR(srcPix[:srcRowBytes*height], dst)
		return
	}

	for y := 0; y < height; y++ {
		srcStart := y * srcStride
		dstStart := y * dstRowBytes
		encodeUnpremultipliedRGBARowToBGR(srcPix[srcStart:srcStart+srcRowBytes], dst[dstStart:dstStart+dstRowBytes])
	}
}

func encodeUnpremultipliedRGBARowToBGR(src, dst []byte) {
	for srcIdx, dstIdx := 0, 0; dstIdx < len(dst); srcIdx, dstIdx = srcIdx+bytesPerRGBA, dstIdx+bytesPerBGR {
		r, g, b := unpremultiplyRGBAExact(src[srcIdx], src[srcIdx+1], src[srcIdx+2], src[srcIdx+3])
		dst[dstIdx] = b
		dst[dstIdx+1] = g
		dst[dstIdx+2] = r
	}
}

func unpremultiplyRGBAExact(r, g, b, a byte) (byte, byte, byte) {
	if a == 0 {
		return 0, 0, 0
	}
	if a == 0xff {
		return r, g, b
	}

	alpha := uint32(a)
	return uint8(((uint32(r) * 0xffff) / alpha) >> 8),
		uint8(((uint32(g) * 0xffff) / alpha) >> 8),
		uint8(((uint32(b) * 0xffff) / alpha) >> 8)
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

// Resize resizes the image buffer to the specified width and height. It
// returns true if the operation was successful, otherwise false. An empty
// buffer and a zero width together with a zero height fail; a single zero
// dimension is computed proportionally from the other one.
func (i *ImageBuffer) Resize(width, height int32) bool {
	return native.MaaImageBufferResize(i.handle, width, height)
}

// NOTE: GetEncoded and SetEncoded are intentionally NOT implemented in Go binding.
// Go handles image encoding/decoding natively through the standard library (image/png, image/jpeg, etc.).
// Do not add encoded image methods here.
