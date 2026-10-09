package buffer

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/require"
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

	err := imageBuffer.Set(img1)
	require.NoError(t, err)

	appended := imageListBuffer.Append(imageBuffer)
	require.True(t, appended)

	got2 := imageListBuffer.IsEmpty()
	require.False(t, got2)

	img2 := imageListBuffer.Get(0)
	require.NotNil(t, img2)
	requireImagesEqual(t, img1, img2)
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

	err := imageBuffer.Set(img)
	require.NoError(t, err)

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

	err := imageBuffer.Set(img1)
	require.NoError(t, err)

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

	err := imageBuffer.Set(img1)
	require.NoError(t, err)

	appended := imageListBuffer.Append(imageBuffer)
	require.True(t, appended)

	list := imageListBuffer.GetAll()
	require.Len(t, list, 1)
}

// solidTestImage returns a new 2x2 NRGBA test image with a distinct pixel in
// each corner.
func solidTestImage() *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
	img.SetNRGBA(0, 1, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
	img.SetNRGBA(1, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	return img
}

// TestImageListBuffer_AppendNilReturnsFalse pins the null-value contract:
// upstream rejects a null value with false instead of crashing.
func TestImageListBuffer_AppendNilReturnsFalse(t *testing.T) {
	imageListBuffer := createImageListBuffer(t)
	defer imageListBuffer.Destroy()

	require.False(t, imageListBuffer.Append(nil))
	require.True(t, imageListBuffer.IsEmpty())
	require.Equal(t, uint64(0), imageListBuffer.Size())
}

// TestImageListBuffer_AppendDeepCopyOwnership pins the Append ownership
// contract: the stored element stays valid independently of the source, so
// destroying the source buffer afterwards must not affect it.
func TestImageListBuffer_AppendDeepCopyOwnership(t *testing.T) {
	imageListBuffer := createImageListBuffer(t)
	defer imageListBuffer.Destroy()

	img := solidTestImage()
	imageBuffer := createImageBuffer(t)
	require.NoError(t, imageBuffer.Set(img))
	require.True(t, imageListBuffer.Append(imageBuffer))

	imageBuffer.Destroy()

	requireImagesEqual(t, img, imageListBuffer.Get(0))
}

// TestImageListBuffer_OutOfRange pins the out-of-range degradation: upstream
// logs and returns null/false, which the wrapper surfaces as a nil image from
// Get and false from Remove.
func TestImageListBuffer_OutOfRange(t *testing.T) {
	imageListBuffer := createImageListBuffer(t)
	defer imageListBuffer.Destroy()

	if got := imageListBuffer.Get(0); got != nil {
		t.Fatalf("Get(0) on an empty list = %T; want a nil image.Image interface", got)
	}
	require.False(t, imageListBuffer.Remove(0))

	imageBuffer := createImageBuffer(t)
	defer imageBuffer.Destroy()
	require.NoError(t, imageBuffer.Set(solidTestImage()))
	require.True(t, imageListBuffer.Append(imageBuffer))

	if got := imageListBuffer.Get(1); got != nil {
		t.Fatalf("Get(1) past the last element = %T; want a nil image.Image interface", got)
	}
	require.False(t, imageListBuffer.Remove(1))
	require.NotNil(t, imageListBuffer.Get(0))
}

// TestImageListBuffer_GetAllContentAndOrder pins GetAll element content,
// ordering, and the non-nil empty slice on a fresh list.
func TestImageListBuffer_GetAllContentAndOrder(t *testing.T) {
	imageListBuffer := createImageListBuffer(t)
	defer imageListBuffer.Destroy()

	empty := imageListBuffer.GetAll()
	require.NotNil(t, empty)
	require.Empty(t, empty)

	first := solidTestImage()
	second := solidTestImage()
	second.SetNRGBA(0, 0, color.NRGBA{R: 1, G: 2, B: 3, A: 255})

	for _, img := range []*image.NRGBA{first, second} {
		imageBuffer := createImageBuffer(t)
		t.Cleanup(imageBuffer.Destroy)
		require.NoError(t, imageBuffer.Set(img))
		require.True(t, imageListBuffer.Append(imageBuffer))
	}

	all := imageListBuffer.GetAll()
	require.Len(t, all, 2)
	requireImagesEqual(t, first, all[0])
	requireImagesEqual(t, second, all[1])
}
