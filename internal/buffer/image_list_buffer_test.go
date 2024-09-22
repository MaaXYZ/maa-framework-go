package buffer

import (
	"github.com/stretchr/testify/require"
	"image"
	"image/color"
	"testing"
)

func createImageListBuffer(t *testing.T) *ImageListBuffer {
	imageListBuffer := NewImageListBuffer()
	require.NotNil(t, imageListBuffer)
	return imageListBuffer
}

func TestNewImageListBuffer(t *testing.T) {
	imageListBuffer := createImageListBuffer(t)
	imageListBuffer.Destroy()
}

func TestImageListBuffer_Handle(t *testing.T) {
	imageListBuffer := createImageListBuffer(t)
	defer imageListBuffer.Destroy()
	handle := imageListBuffer.Handle()
	require.NotNil(t, handle)
}

func TestImageListBuffer_IsEmpty(t *testing.T) {
	imageListBuffer := createImageListBuffer(t)
	defer imageListBuffer.Destroy()
	got := imageListBuffer.IsEmpty()
	require.True(t, got)
}

func TestImageListBuffer_Clear(t *testing.T) {
	imageListBuffer := createImageListBuffer(t)
	defer imageListBuffer.Destroy()
	got := imageListBuffer.Clear()
	require.True(t, got)
}

func TestImageListBuffer_Append(t *testing.T) {
	imageListBuffer := createImageListBuffer(t)
	defer imageListBuffer.Destroy()

	imageBuffer := createImageBuffer(t)
	defer imageBuffer.Destroy()

	width, height := 2, 2
	img1 := image.NewNRGBA(image.Rect(0, 0, width, height))
	img1.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	img1.SetNRGBA(1, 0, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
	img1.SetNRGBA(0, 1, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
	img1.SetNRGBA(1, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	got := imageBuffer.Set(img1)
	require.True(t, got)

	appended := imageListBuffer.Append(imageBuffer)
	require.True(t, appended)

	got2 := imageListBuffer.IsEmpty()
	require.False(t, got2)

	img2 := imageListBuffer.Get(0)
	require.NotNil(t, img2)
	require.Equal(t, img1, img2)
}

func TestImageListBuffer_Remove(t *testing.T) {
	imageListBuffer := createImageListBuffer(t)
	defer imageListBuffer.Destroy()

	imageBuffer := createImageBuffer(t)
	defer imageBuffer.Destroy()

	width, height := 2, 2
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
	img.SetNRGBA(0, 1, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
	img.SetNRGBA(1, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	got := imageBuffer.Set(img)
	require.True(t, got)

	appended := imageListBuffer.Append(imageBuffer)
	require.True(t, appended)

	removed := imageListBuffer.Remove(0)
	require.True(t, removed)

	got2 := imageListBuffer.IsEmpty()
	require.True(t, got2)
}

func TestImageListBuffer_Size(t *testing.T) {
	imageListBuffer := createImageListBuffer(t)
	defer imageListBuffer.Destroy()

	imageBuffer := createImageBuffer(t)
	defer imageBuffer.Destroy()

	width, height := 2, 2
	img1 := image.NewNRGBA(image.Rect(0, 0, width, height))
	img1.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	img1.SetNRGBA(1, 0, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
	img1.SetNRGBA(0, 1, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
	img1.SetNRGBA(1, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	got := imageBuffer.Set(img1)
	require.True(t, got)

	appended := imageListBuffer.Append(imageBuffer)
	require.True(t, appended)

	size := imageListBuffer.Size()
	require.Equal(t, uint64(1), size)
}

func TestImageListBuffer_GetAll(t *testing.T) {
	imageListBuffer := createImageListBuffer(t)
	defer imageListBuffer.Destroy()

	imageBuffer := createImageBuffer(t)
	defer imageBuffer.Destroy()

	width, height := 2, 2
	img1 := image.NewNRGBA(image.Rect(0, 0, width, height))
	img1.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	img1.SetNRGBA(1, 0, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
	img1.SetNRGBA(0, 1, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
	img1.SetNRGBA(1, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	got := imageBuffer.Set(img1)
	require.True(t, got)

	appended := imageListBuffer.Append(imageBuffer)
	require.True(t, appended)

	list := imageListBuffer.GetAll()
	require.Len(t, list, 1)
}
