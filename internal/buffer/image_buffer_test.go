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

func requireImagesEqual(t *testing.T, expected, actual image.Image) {
	t.Helper()
	require.NotNil(t, expected)
	require.NotNil(t, actual)
	require.Equal(t, expected.Bounds(), actual.Bounds())

	bounds := expected.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			expR, expG, expB, expA := expected.At(x, y).RGBA()
			actR, actG, actB, actA := actual.At(x, y).RGBA()
			require.Equalf(t, expR, actR, "R mismatch at (%d, %d)", x, y)
			require.Equalf(t, expG, actG, "G mismatch at (%d, %d)", x, y)
			require.Equalf(t, expB, actB, "B mismatch at (%d, %d)", x, y)
			require.Equalf(t, expA, actA, "A mismatch at (%d, %d)", x, y)
		}
	}
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
	requireImagesEqual(t, img1, img2)

	t.Run("supports RGBA input", func(t *testing.T) {
		rgba := image.NewRGBA(image.Rect(0, 0, width, height))
		rgba.SetRGBA(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		rgba.SetRGBA(1, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		rgba.SetRGBA(0, 1, color.RGBA{R: 0, G: 0, B: 255, A: 255})
		rgba.SetRGBA(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})

		require.True(t, imageBuffer.Set(rgba))
		got := imageBuffer.Get()
		require.NotNil(t, got)
		requireImagesEqual(t, img1, got)
	})

	t.Run("supports premultiplied RGBA input", func(t *testing.T) {
		rgba := image.NewRGBA(image.Rect(0, 0, 1, 1))
		rgba.Pix[0] = 128
		rgba.Pix[1] = 0
		rgba.Pix[2] = 0
		rgba.Pix[3] = 128

		require.True(t, imageBuffer.Set(rgba))
		got := imageBuffer.Get()
		require.NotNil(t, got)

		gotRGBA, ok := got.(*image.RGBA)
		require.True(t, ok)
		require.Equal(t, color.RGBA{R: 255, G: 0, B: 0, A: 255}, gotRGBA.RGBAAt(0, 0))
	})

	t.Run("matches legacy conversion for low-alpha RGBA input", func(t *testing.T) {
		rgba := image.NewRGBA(image.Rect(0, 0, 1, 1))
		rgba.Pix[0] = 1
		rgba.Pix[1] = 0
		rgba.Pix[2] = 0
		rgba.Pix[3] = 2

		require.True(t, imageBuffer.Set(rgba))
		got := imageBuffer.Get()
		require.NotNil(t, got)

		gotRGBA, ok := got.(*image.RGBA)
		require.True(t, ok)

		expected := color.NRGBAModel.Convert(color.RGBA{R: 1, G: 0, B: 0, A: 2}).(color.NRGBA)
		expected.A = 255
		require.Equal(t, color.RGBA{R: expected.R, G: expected.G, B: expected.B, A: expected.A}, gotRGBA.RGBAAt(0, 0))
	})

	t.Run("supports RGBA sub-image with larger stride", func(t *testing.T) {
		parent := image.NewRGBA(image.Rect(0, 0, width+2, height+2))
		sub := parent.SubImage(image.Rect(1, 1, 1+width, 1+height)).(*image.RGBA)
		require.Greater(t, sub.Stride, width*4)

		src := image.NewRGBA(image.Rect(0, 0, width, height))
		src.SetRGBA(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		src.SetRGBA(1, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		src.SetRGBA(0, 1, color.RGBA{R: 0, G: 0, B: 255, A: 255})
		src.SetRGBA(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})

		for y := 0; y < height; y++ {
			srcRow := src.Pix[y*src.Stride : y*src.Stride+width*4]
			dstRow := sub.Pix[y*sub.Stride : y*sub.Stride+width*4]
			copy(dstRow, srcRow)
		}

		require.True(t, imageBuffer.Set(sub))
		got := imageBuffer.Get()
		require.NotNil(t, got)
		requireImagesEqual(t, img1, got)
	})

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

		gotRGBA, ok := got.(*image.RGBA)
		require.True(t, ok)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				expected := img1.NRGBAAt(x, y)
				require.Equal(t, color.RGBA{R: expected.R, G: expected.G, B: expected.B, A: expected.A}, gotRGBA.RGBAAt(x, y))
			}
		}
	})

	t.Run("handles zero-sized image without panic", func(t *testing.T) {
		empty := image.NewNRGBA(image.Rect(0, 0, 0, 0))

		require.NotPanics(t, func() {
			require.True(t, imageBuffer.Set(empty))
		})
		require.True(t, imageBuffer.IsEmpty())
		require.Nil(t, imageBuffer.Get())
	})
}

func TestImageBuffer_GetInto(t *testing.T) {
	imageBuffer := createImageBuffer(t)
	defer imageBuffer.Destroy()

	width, height := 2, 2
	img1 := image.NewRGBA(image.Rect(0, 0, width, height))
	img1.SetRGBA(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	img1.SetRGBA(1, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	img1.SetRGBA(0, 1, color.RGBA{R: 0, G: 0, B: 255, A: 255})
	img1.SetRGBA(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	require.True(t, imageBuffer.Set(img1))

	reused := image.NewRGBA(image.Rect(0, 0, width, height))
	got1 := imageBuffer.GetInto(reused)
	require.NotNil(t, got1)
	require.Same(t, reused, got1)
	require.Equal(t, img1, got1)

	mismatch := image.NewRGBA(image.Rect(0, 0, 1, 1))
	got2 := imageBuffer.GetInto(mismatch)
	require.NotNil(t, got2)
	require.NotSame(t, mismatch, got2)
	require.Equal(t, img1, got2)

	t.Run("returns nil and does not mutate dst when buffer has no raw data", func(t *testing.T) {
		require.True(t, imageBuffer.Clear())

		gotNilDst := imageBuffer.GetInto(nil)
		require.Nil(t, gotNilDst)

		dst := image.NewRGBA(image.Rect(0, 0, width, height))
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

		parent := image.NewRGBA(image.Rect(0, 0, 4, 4))
		sub := parent.SubImage(image.Rect(1, 1, 3, 3)).(*image.RGBA)
		require.Greater(t, sub.Stride, width*4)

		got := imageBuffer.GetInto(sub)
		require.NotNil(t, got)
		require.Same(t, sub, got)
		require.Equal(t, width, got.Rect.Dx())
		require.Equal(t, height, got.Rect.Dy())

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				require.Equal(t, img1.RGBAAt(x, y), got.RGBAAt(got.Rect.Min.X+x, got.Rect.Min.Y+y))
			}
		}
	})
}
