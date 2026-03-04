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

func TestImageBuffer_Set(t *testing.T) {
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

	img2 := imageBuffer.Get()
	require.NotNil(t, img2)
	require.Equal(t, img1, img2)

	t.Run("supports NRGBA sub-image with larger stride", func(t *testing.T) {
		parent := image.NewNRGBA(image.Rect(0, 0, width+2, height+2))
		sub := parent.SubImage(image.Rect(1, 1, 1+width, 1+height)).(*image.NRGBA)
		require.Greater(t, sub.Stride, width*4)

		for y := 0; y < height; y++ {
			srcRow := img1.Pix[y*img1.Stride : y*img1.Stride+width*4]
			dstRow := sub.Pix[y*sub.Stride : y*sub.Stride+width*4]
			copy(dstRow, srcRow)
		}

		require.True(t, imageBuffer.Set(sub))
		got := imageBuffer.Get()
		require.NotNil(t, got)

		gotNRGBA, ok := got.(*image.NRGBA)
		require.True(t, ok)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				require.Equal(t, img1.NRGBAAt(x, y), gotNRGBA.NRGBAAt(x, y))
			}
		}
	})
}

func TestImageBuffer_GetInto(t *testing.T) {
	imageBuffer := createImageBuffer(t)
	defer imageBuffer.Destroy()

	width, height := 2, 2
	img1 := image.NewNRGBA(image.Rect(0, 0, width, height))
	img1.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	img1.SetNRGBA(1, 0, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
	img1.SetNRGBA(0, 1, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
	img1.SetNRGBA(1, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	require.True(t, imageBuffer.Set(img1))

	reused := image.NewNRGBA(image.Rect(0, 0, width, height))
	got1 := imageBuffer.GetInto(reused)
	require.NotNil(t, got1)
	require.Same(t, reused, got1)
	require.Equal(t, img1, got1)

	mismatch := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	got2 := imageBuffer.GetInto(mismatch)
	require.NotNil(t, got2)
	require.NotSame(t, mismatch, got2)
	require.Equal(t, img1, got2)

	t.Run("returns nil and does not mutate dst when buffer has no raw data", func(t *testing.T) {
		require.True(t, imageBuffer.Clear())

		gotNilDst := imageBuffer.GetInto(nil)
		require.Nil(t, gotNilDst)

		dst := image.NewNRGBA(image.Rect(0, 0, width, height))
		copy(dst.Pix, []byte{
			1, 2, 3, 4,
			5, 6, 7, 8,
			9, 10, 11, 12,
			13, 14, 15, 16,
		})
		before := append([]byte(nil), dst.Pix...)
		gotNonNilDst := imageBuffer.GetInto(dst)
		require.Nil(t, gotNonNilDst)
		require.Equal(t, before, dst.Pix)
	})

	t.Run("reuses dst with larger stride and takes slow path", func(t *testing.T) {
		require.True(t, imageBuffer.Set(img1))

		parent := image.NewNRGBA(image.Rect(0, 0, 4, 4))
		sub := parent.SubImage(image.Rect(1, 1, 3, 3)).(*image.NRGBA)
		require.Greater(t, sub.Stride, width*4)

		got := imageBuffer.GetInto(sub)
		require.NotNil(t, got)
		require.Same(t, sub, got)
		require.Equal(t, width, got.Rect.Dx())
		require.Equal(t, height, got.Rect.Dy())

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				require.Equal(t, img1.NRGBAAt(x, y), got.NRGBAAt(got.Rect.Min.X+x, got.Rect.Min.Y+y))
			}
		}
	})
}
