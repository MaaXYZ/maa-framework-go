package maa

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

type rectBuffer struct {
	handle *C.MaaRect
}

func newRectBuffer() *rectBuffer {
	handle := C.MaaRectCreate()
	if handle == nil {
		return nil
	}
	return &rectBuffer{
		handle: handle,
	}
}

func newRectBufferByHandle(handle unsafe.Pointer) *rectBuffer {
	return &rectBuffer{
		handle: (*C.MaaRect)(handle),
	}
}

func (r *rectBuffer) Destroy() {
	C.MaaRectDestroy(r.handle)
}

func (r *rectBuffer) Handle() unsafe.Pointer {
	return unsafe.Pointer(r.handle)
}

func (r *rectBuffer) Get() Rect {
	return Rect{r.GetX(), r.GetY(), r.GetW(), r.GetH()}
}

func (r *rectBuffer) GetX() int32 {
	return int32(C.MaaRectGetX(r.handle))
}

func (r *rectBuffer) GetY() int32 {
	return int32(C.MaaRectGetY(r.handle))
}

func (r *rectBuffer) GetW() int32 {
	return int32(C.MaaRectGetW(r.handle))
}

func (r *rectBuffer) GetH() int32 {
	return int32(C.MaaRectGetH(r.handle))
}

func (r *rectBuffer) Set(rect Rect) bool {
	return C.MaaRectSet(
		r.handle,
		C.int32_t(rect.X),
		C.int32_t(rect.Y),
		C.int32_t(rect.W),
		C.int32_t(rect.H),
	) != 0
}

type stringBuffer struct {
	handle *C.MaaStringBuffer
}

func newStringBuffer() *stringBuffer {
	handle := C.MaaStringBufferCreate()
	if handle == nil {
		return nil
	}
	return &stringBuffer{
		handle: handle,
	}
}

func newStringBufferByHandle(handle unsafe.Pointer) *stringBuffer {
	return &stringBuffer{
		handle: (*C.MaaStringBuffer)(handle),
	}
}

func (s *stringBuffer) Destroy() {
	C.MaaStringBufferDestroy(s.handle)
}

func (s *stringBuffer) Handle() unsafe.Pointer {
	return unsafe.Pointer(s.handle)
}

func (s *stringBuffer) IsEmpty() bool {
	return C.MaaStringBufferIsEmpty(s.handle) != 0
}

func (s *stringBuffer) Clear() bool {
	return C.MaaStringBufferClear(s.handle) != 0
}

func (s *stringBuffer) Get() string {
	return C.GoString(C.MaaStringBufferGet(s.handle))
}

func (s *stringBuffer) Size() uint64 {
	return uint64(C.MaaStringBufferSize(s.handle))
}

func (s *stringBuffer) Set(str string) bool {
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr))
	return C.MaaStringBufferSet(
		s.handle,
		cStr,
	) != 0
}

func (s *stringBuffer) SetWithSize(str string, size uint64) bool {
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr))
	return C.MaaStringBufferSetEx(
		s.handle,
		cStr,
		C.uint64_t(size),
	) != 0
}

type stringListBuffer struct {
	handle *C.MaaStringListBuffer
}

func newStringListBuffer() *stringListBuffer {
	handle := C.MaaStringListBufferCreate()
	if handle == nil {
		return nil
	}
	return &stringListBuffer{
		handle: handle,
	}
}

func newStringListBufferByHandle(handle unsafe.Pointer) *stringListBuffer {
	return &stringListBuffer{
		handle: (*C.MaaStringListBuffer)(handle),
	}
}

func (sl *stringListBuffer) Destroy() {
	C.MaaStringListBufferDestroy(sl.handle)
}

func (sl *stringListBuffer) Handle() unsafe.Pointer {
	return unsafe.Pointer(sl.handle)
}

func (sl *stringListBuffer) IsEmpty() bool {
	return C.MaaStringListBufferIsEmpty(sl.handle) != 0
}

func (sl *stringListBuffer) Clear() bool {
	return C.MaaStringListBufferClear(sl.handle) != 0
}

func (sl *stringListBuffer) Size() uint64 {
	return uint64(C.MaaStringListBufferSize(sl.handle))
}

func (sl *stringListBuffer) Get(index uint64) string {
	handle := C.MaaStringListBufferAt(sl.handle, C.uint64_t(index))
	str := &stringBuffer{
		handle: handle,
	}
	return str.Get()
}

func (sl *stringListBuffer) GetAll() []string {
	size := sl.Size()
	strings := make([]string, size)
	for i := uint64(0); i < size; i++ {
		strings[i] = sl.Get(i)
	}
	return strings
}

func (sl *stringListBuffer) Append(value *stringBuffer) bool {
	return C.MaaStringListBufferAppend(
		sl.handle,
		(*C.MaaStringBuffer)(value.Handle()),
	) != 0
}

func (sl *stringListBuffer) Remove(index uint64) bool {
	return C.MaaStringListBufferRemove(
		sl.handle,
		C.uint64_t(index),
	) != 0
}

type imageBuffer struct {
	handle *C.MaaImageBuffer
}

func newImageBuffer() *imageBuffer {
	handle := C.MaaImageBufferCreate()
	if handle == nil {
		return nil
	}
	return &imageBuffer{
		handle: handle,
	}
}

func newImageBufferByHandle(handle unsafe.Pointer) *imageBuffer {
	return &imageBuffer{
		handle: (*C.MaaImageBuffer)(handle),
	}
}

func (i *imageBuffer) Destroy() {
	C.MaaImageBufferDestroy(i.handle)
}

func (i *imageBuffer) Handle() unsafe.Pointer {
	return unsafe.Pointer(i.handle)
}

func (i *imageBuffer) IsEmpty() bool {
	return C.MaaImageBufferIsEmpty(i.handle) != 0
}

func (i *imageBuffer) Clear() bool {
	return C.MaaImageBufferClear(i.handle) != 0
}

// Get retrieves the image from raw data stored in the buffer.
func (i *imageBuffer) Get() image.Image {
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

// Set converts an image.Image to raw data and sets it in the buffer.
func (i *imageBuffer) Set(img image.Image) bool {
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
func (i *imageBuffer) getRawData() unsafe.Pointer {
	return unsafe.Pointer(C.MaaImageBufferGetRawData(i.handle))
}

// getWidth retrieves the width of the image stored in the buffer.
// It returns the width as an int32.
func (i *imageBuffer) getWidth() int32 {
	return int32(C.MaaImageBufferWidth(i.handle))
}

// getHeight retrieves the height of the image stored in the buffer.
// It returns the height as an int32.
func (i *imageBuffer) getHeight() int32 {
	return int32(C.MaaImageBufferHeight(i.handle))
}

// getType retrieves the type of the image stored in the buffer.
// This corresponds to the cv::Mat.type() in OpenCV.
// It returns the type as an int32.
func (i *imageBuffer) getType() int32 {
	return int32(C.MaaImageBufferType(i.handle))
}

// setRawData sets the raw image data in the buffer.
// It takes a pointer to the raw image data, the width, height, and type of the image.
// It returns true if the operation was successful, otherwise false.
func (i *imageBuffer) setRawData(data unsafe.Pointer, width, height, imageType int32) bool {
	return C.MaaImageBufferSetRawData(
		i.handle,
		C.MaaImageRawData(data),
		C.int32_t(width),
		C.int32_t(height),
		C.int32_t(imageType),
	) != 0
}

type imageListBuffer struct {
	handle *C.MaaImageListBuffer
}

func newImageListBuffer() *imageListBuffer {
	handle := C.MaaImageListBufferCreate()
	if handle == nil {
		return nil
	}
	return &imageListBuffer{
		handle: handle,
	}
}

func newImageListBufferByHandle(handle unsafe.Pointer) *imageListBuffer {
	return &imageListBuffer{
		handle: (*C.MaaImageListBuffer)(handle),
	}
}

func (il *imageListBuffer) Destroy() {
	C.MaaImageListBufferDestroy(il.handle)
}

func (il *imageListBuffer) Handle() unsafe.Pointer {
	return unsafe.Pointer(il.handle)
}

func (il *imageListBuffer) IsEmpty() bool {
	return C.MaaImageListBufferIsEmpty(il.handle) != 0
}

func (il *imageListBuffer) Clear() bool {
	return C.MaaImageListBufferClear(il.handle) != 0
}

func (il *imageListBuffer) Size() uint64 {
	return uint64(C.MaaImageListBufferSize(il.handle))
}

func (il *imageListBuffer) Get(index uint64) image.Image {
	handle := C.MaaImageListBufferAt(il.handle, C.uint64_t(index))
	img := &imageBuffer{
		handle: handle,
	}
	return img.Get()
}

func (il *imageListBuffer) GetAll() []image.Image {
	size := il.Size()
	images := make([]image.Image, size)
	for i := uint64(0); i < size; i++ {
		img := il.Get(i)
		images[i] = img
	}
	return images
}

func (il *imageListBuffer) Append(value *imageBuffer) bool {
	return C.MaaImageListBufferAppend(
		il.handle,
		(*C.MaaImageBuffer)(value.Handle()),
	) != 0
}

func (il *imageListBuffer) Remove(index uint64) bool {
	return C.MaaImageListBufferRemove(
		il.handle,
		C.uint64_t(index),
	) != 0
}
