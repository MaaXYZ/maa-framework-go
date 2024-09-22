package buffer

import (
	"github.com/stretchr/testify/require"
	"image"
	"image/color"
	"testing"
)

func createImageBuffer(t *testing.T) *ImageBuffer {
	imageBuffer := NewImageBuffer()
	require.NotNil(t, imageBuffer)
	return imageBuffer
}

func TestNewImageBuffer(t *testing.T) {
	imageBuffer := createImageBuffer(t)
	imageBuffer.Destroy()
}

func TestImageBuffer_Handle(t *testing.T) {
	imageBuffer := createImageBuffer(t)
	defer imageBuffer.Destroy()
	handle := imageBuffer.Handle()
	require.NotNil(t, handle)
}

func TestImageBuffer_IsEmpty(t *testing.T) {
	imageBuffer := createImageBuffer(t)
	defer imageBuffer.Destroy()
	got := imageBuffer.IsEmpty()
	require.True(t, got)
}

func TestImageBuffer_Clear(t *testing.T) {
	imageBuffer := createImageBuffer(t)
	defer imageBuffer.Destroy()
	got := imageBuffer.Clear()
	require.True(t, got)
}

func TestImageBuffer_SetRawData(t *testing.T) {
	imageBuffer := createImageBuffer(t)
	defer imageBuffer.Destroy()

	width, height := 2, 2
	img1 := image.NewNRGBA(image.Rect(0, 0, width, height))
	img1.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	img1.SetNRGBA(1, 0, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
	img1.SetNRGBA(0, 1, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
	img1.SetNRGBA(1, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	got := imageBuffer.SetRawData(img1)
	require.True(t, got)

	img2 := imageBuffer.GetByRawData()
	require.NotNil(t, img2)
	require.Equal(t, img1, img2)
}
