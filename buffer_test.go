package maa

import (
	"github.com/stretchr/testify/require"
	"image"
	"image/color"
	"testing"
)

func createRectBuffer(t *testing.T) *rectBuffer {
	rectBuf := newRectBuffer()
	require.NotNil(t, rectBuf)
	return rectBuf
}

func TestNewRectBuffer(t *testing.T) {
	rectBuf := createRectBuffer(t)
	rectBuf.Destroy()
}

func TestRectBuffer_Handle(t *testing.T) {
	rectBuf := createRectBuffer(t)
	defer rectBuf.Destroy()
	handle := rectBuf.Handle()
	require.NotNil(t, handle)
}

func TestRectBuffer_Set(t *testing.T) {
	rectBuf := createRectBuffer(t)
	defer rectBuf.Destroy()

	rect1 := Rect{100, 200, 300, 400}
	got := rectBuf.Set(rect1)
	require.True(t, got)

	x := rectBuf.GetX()
	require.Equal(t, rect1.X, x)
	y := rectBuf.GetY()
	require.Equal(t, rect1.Y, y)
	w := rectBuf.GetW()
	require.Equal(t, rect1.W, w)
	h := rectBuf.GetH()
	require.Equal(t, rect1.H, h)
	rect2 := rectBuf.Get()
	require.Equal(t, rect1, rect2)
}

func createStringBuffer(t *testing.T) *stringBuffer {
	strBuf := newStringBuffer()
	require.NotNil(t, strBuf)
	return strBuf
}

func TestNewStringBuffer(t *testing.T) {
	strBuf := createStringBuffer(t)
	strBuf.Destroy()
}

func TestStringBuffer_Handle(t *testing.T) {
	strBuf := createStringBuffer(t)
	defer strBuf.Destroy()
	handle := strBuf.Handle()
	require.NotNil(t, handle)
}

func TestStringBuffer_IsEmpty(t *testing.T) {
	strBuf := createStringBuffer(t)
	defer strBuf.Destroy()
	got := strBuf.IsEmpty()
	require.True(t, got)
}

func TestStringBuffer_Clear(t *testing.T) {
	strBuf := createStringBuffer(t)
	defer strBuf.Destroy()
	got := strBuf.Clear()
	require.True(t, got)
}

func TestStringBuffer_Set(t *testing.T) {
	strBuf := createStringBuffer(t)
	defer strBuf.Destroy()
	str1 := "test"
	got := strBuf.Set(str1)
	require.True(t, got)

	str2 := strBuf.Get()
	require.Equal(t, str1, str2)
}

func TestStringBuffer_Size(t *testing.T) {
	strBuf := createStringBuffer(t)
	defer strBuf.Destroy()
	str := "test"
	got := strBuf.Set(str)
	require.True(t, got)

	size := strBuf.Size()
	require.Equal(t, uint64(len(str)), size)
}

func TestStringBuffer_SetWithSize(t *testing.T) {
	strBuf := createStringBuffer(t)
	defer strBuf.Destroy()
	str1 := "test"
	got := strBuf.SetWithSize(str1, uint64(len(str1)))
	require.True(t, got)

	str2 := strBuf.Get()
	require.Equal(t, str1, str2)
}

func createStringListBuffer(t *testing.T) *stringListBuffer {
	strListBuf := newStringListBuffer()
	require.NotNil(t, strListBuf)
	return strListBuf
}

func TestNewStringListBuffer(t *testing.T) {
	strListBuf := createStringListBuffer(t)
	strListBuf.Destroy()
}

func TestStringListBuffer_Handle(t *testing.T) {
	strListBuf := createStringListBuffer(t)
	defer strListBuf.Destroy()
	handle := strListBuf.Handle()
	require.NotNil(t, handle)
}

func TestStringListBuffer_IsEmpty(t *testing.T) {
	strListBuf := createStringListBuffer(t)
	defer strListBuf.Destroy()
	got := strListBuf.IsEmpty()
	require.True(t, got)
}

func TestStringListBuffer_Clear(t *testing.T) {
	strListBuf := createStringListBuffer(t)
	defer strListBuf.Destroy()
	got := strListBuf.Clear()
	require.True(t, got)
}

func TestStringListBuffer_Append(t *testing.T) {
	strListBuf := createStringListBuffer(t)
	defer strListBuf.Destroy()

	strBuf := createStringBuffer(t)
	str1 := "test"
	got1 := strBuf.Set(str1)
	require.True(t, got1)

	got2 := strListBuf.Append(strBuf)
	require.True(t, got2)

	got3 := strListBuf.IsEmpty()
	require.False(t, got3)

	str2 := strListBuf.Get(0)
	require.Equal(t, str1, str2)
}

func TestStringListBuffer_Remove(t *testing.T) {
	strListBuf := createStringListBuffer(t)
	defer strListBuf.Destroy()

	strBuf := createStringBuffer(t)
	str1 := "test"
	got1 := strBuf.Set(str1)
	require.True(t, got1)

	got2 := strListBuf.Append(strBuf)
	require.True(t, got2)

	removed := strListBuf.Remove(0)
	require.True(t, removed)

	got3 := strListBuf.IsEmpty()
	require.True(t, got3)
}

func TestStringListBuffer_Size(t *testing.T) {
	strListBuf := createStringListBuffer(t)
	defer strListBuf.Destroy()

	strBuf := createStringBuffer(t)
	str1 := "test"
	got1 := strBuf.Set(str1)
	require.True(t, got1)

	got2 := strListBuf.Append(strBuf)
	require.True(t, got2)

	got3 := strListBuf.IsEmpty()
	require.False(t, got3)

	size := strListBuf.Size()
	require.Equal(t, uint64(1), size)
}

func TestStringListBuffer_GetAll(t *testing.T) {
	strListBuf := createStringListBuffer(t)
	defer strListBuf.Destroy()

	strBuf := createStringBuffer(t)
	str1 := "test"
	got1 := strBuf.Set(str1)
	require.True(t, got1)

	got2 := strListBuf.Append(strBuf)
	require.True(t, got2)

	got3 := strListBuf.IsEmpty()
	require.False(t, got3)

	list := strListBuf.GetAll()
	require.Len(t, list, 1)
}

func createImageBuffer(t *testing.T) *imageBuffer {
	imgBuf := newImageBuffer()
	require.NotNil(t, imgBuf)
	return imgBuf
}

func TestNewImageBuffer(t *testing.T) {
	imgBuf := createImageBuffer(t)
	imgBuf.Destroy()
}

func TestImageBuffer_Handle(t *testing.T) {
	imgBuf := createImageBuffer(t)
	defer imgBuf.Destroy()
	handle := imgBuf.Handle()
	require.NotNil(t, handle)
}

func TestImageBuffer_IsEmpty(t *testing.T) {
	imgBuf := createImageBuffer(t)
	defer imgBuf.Destroy()
	got := imgBuf.IsEmpty()
	require.True(t, got)
}

func TestImageBuffer_Clear(t *testing.T) {
	imgBuf := createImageBuffer(t)
	defer imgBuf.Destroy()
	got := imgBuf.Clear()
	require.True(t, got)
}

func TestImageBuffer_Set(t *testing.T) {
	imgBuf := createImageBuffer(t)
	defer imgBuf.Destroy()

	width, height := 2, 2
	img1 := image.NewNRGBA(image.Rect(0, 0, width, height))
	img1.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	img1.SetNRGBA(1, 0, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
	img1.SetNRGBA(0, 1, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
	img1.SetNRGBA(1, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	got := imgBuf.Set(img1)
	require.True(t, got)

	img2 := imgBuf.Get()
	require.NotNil(t, img2)
	require.Equal(t, img1, img2)
}

func createImageListBuffer(t *testing.T) *imageListBuffer {
	imgListBuf := newImageListBuffer()
	require.NotNil(t, imgListBuf)
	return imgListBuf
}

func TestNewImageListBuffer(t *testing.T) {
	imgListBuf := createImageListBuffer(t)
	imgListBuf.Destroy()
}

func TestImageListBuffer_Handle(t *testing.T) {
	imgListBuf := createImageListBuffer(t)
	defer imgListBuf.Destroy()
	handle := imgListBuf.Handle()
	require.NotNil(t, handle)
}

func TestImageListBuffer_IsEmpty(t *testing.T) {
	imgListBuf := createImageListBuffer(t)
	defer imgListBuf.Destroy()
	got := imgListBuf.IsEmpty()
	require.True(t, got)
}

func TestImageListBuffer_Clear(t *testing.T) {
	imgListBuf := createImageListBuffer(t)
	defer imgListBuf.Destroy()
	got := imgListBuf.Clear()
	require.True(t, got)
}

func TestImageListBuffer_Append(t *testing.T) {
	imgListBuf := createImageListBuffer(t)
	defer imgListBuf.Destroy()

	imgBuf := createImageBuffer(t)
	defer imgBuf.Destroy()

	width, height := 2, 2
	img1 := image.NewNRGBA(image.Rect(0, 0, width, height))
	img1.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	img1.SetNRGBA(1, 0, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
	img1.SetNRGBA(0, 1, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
	img1.SetNRGBA(1, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	got := imgBuf.Set(img1)
	require.True(t, got)

	appended := imgListBuf.Append(imgBuf)
	require.True(t, appended)

	got2 := imgListBuf.IsEmpty()
	require.False(t, got2)

	img2 := imgListBuf.Get(0)
	require.NotNil(t, img2)
	require.Equal(t, img1, img2)
}

func TestImageListBuffer_Remove(t *testing.T) {
	imgListBuf := createImageListBuffer(t)
	defer imgListBuf.Destroy()

	imgBuf := createImageBuffer(t)
	defer imgBuf.Destroy()

	width, height := 2, 2
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
	img.SetNRGBA(0, 1, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
	img.SetNRGBA(1, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	got := imgBuf.Set(img)
	require.True(t, got)

	appended := imgListBuf.Append(imgBuf)
	require.True(t, appended)

	removed := imgListBuf.Remove(0)
	require.True(t, removed)

	got2 := imgListBuf.IsEmpty()
	require.True(t, got2)
}

func TestImageListBuffer_Size(t *testing.T) {
	imgListBuf := createImageListBuffer(t)
	defer imgListBuf.Destroy()

	imgBuf := createImageBuffer(t)
	defer imgBuf.Destroy()

	width, height := 2, 2
	img1 := image.NewNRGBA(image.Rect(0, 0, width, height))
	img1.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	img1.SetNRGBA(1, 0, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
	img1.SetNRGBA(0, 1, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
	img1.SetNRGBA(1, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	got := imgBuf.Set(img1)
	require.True(t, got)

	appended := imgListBuf.Append(imgBuf)
	require.True(t, appended)

	size := imgListBuf.Size()
	require.Equal(t, uint64(1), size)
}

func TestImageListBuffer_GetAll(t *testing.T) {
	imgListBuf := createImageListBuffer(t)
	defer imgListBuf.Destroy()

	imgBuf := createImageBuffer(t)
	defer imgBuf.Destroy()

	width, height := 2, 2
	img1 := image.NewNRGBA(image.Rect(0, 0, width, height))
	img1.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	img1.SetNRGBA(1, 0, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
	img1.SetNRGBA(0, 1, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
	img1.SetNRGBA(1, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	got := imgBuf.Set(img1)
	require.True(t, got)

	appended := imgListBuf.Append(imgBuf)
	require.True(t, appended)

	list := imgListBuf.GetAll()
	require.Len(t, list, 1)
}
