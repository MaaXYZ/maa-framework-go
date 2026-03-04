package buffer

import (
	"image"
	"image/color"
	"runtime"
	"sync"
	"testing"
	"unsafe"
)

var benchmarkImageSink image.Image

func isLittleEndian() bool {
	var v uint16 = 0x0102
	b := *(*[2]byte)(unsafe.Pointer(&v))
	return b[0] == 0x02
}

// decodeBGRToNRGBASetNRGBA matches the original per-pixel SetNRGBA path in ImageBuffer.Get.
func decodeBGRToNRGBASetNRGBA(raw []byte, width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := (y*width + x) * 3
			r := raw[offset+2]
			g := raw[offset+1]
			b := raw[offset]
			img.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

// decodeBGRToNRGBADirectPix matches the current direct Pix write path in ImageBuffer.Get/GetInto.
func decodeBGRToNRGBADirectPix(raw []byte, width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	decodeBGRToNRGBADirectPixInto(raw, img, width, height)
	return img
}

// decodeBGRToNRGBADirectPixInto writes conversion results into a caller-provided destination buffer.
// This is the core helper used by "reuse dst" benchmarks.
func decodeBGRToNRGBADirectPixInto(raw []byte, img *image.NRGBA, width, height int) {
	dst := img.Pix
	for src, dstIdx := 0, 0; src < len(raw); src, dstIdx = src+3, dstIdx+4 {
		dst[dstIdx] = raw[src+2]
		dst[dstIdx+1] = raw[src+1]
		dst[dstIdx+2] = raw[src]
		dst[dstIdx+3] = 255
	}
}

// decodeBGRToNRGBAUint32 is an experimental unsafe path using uint32 stores on little-endian layout.
func decodeBGRToNRGBAUint32(raw []byte, width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	if len(img.Pix) == 0 {
		return img
	}

	dst := unsafe.Slice((*uint32)(unsafe.Pointer(&img.Pix[0])), len(img.Pix)/4)
	for src, i := 0, 0; src < len(raw); src, i = src+3, i+1 {
		// Image.NRGBA stores bytes as RGBA. On little-endian, one uint32 can write all 4 bytes.
		dst[i] = 0xff000000 | uint32(raw[src])<<16 | uint32(raw[src+1])<<8 | uint32(raw[src+2])
	}
	return img
}

// decodeBGRToNRGBADirectPixParallel is an experimental multi-goroutine converter.
// It may reduce conversion latency but can increase allocations/scheduling overhead.
func decodeBGRToNRGBADirectPixParallel(raw []byte, width, height, workers int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	if workers <= 1 {
		decodeBGRToNRGBADirectPixInto(raw, img, width, height)
		return img
	}

	if workers > height {
		workers = height
	}

	var wg sync.WaitGroup
	wg.Add(workers)
	for worker := 0; worker < workers; worker++ {
		startY := height * worker / workers
		endY := height * (worker + 1) / workers
		go func(start, end int) {
			defer wg.Done()
			for y := start; y < end; y++ {
				srcIdx := y * width * 3
				dstIdx := y * img.Stride
				for x := 0; x < width; x++ {
					img.Pix[dstIdx] = raw[srcIdx+2]
					img.Pix[dstIdx+1] = raw[srcIdx+1]
					img.Pix[dstIdx+2] = raw[srcIdx]
					img.Pix[dstIdx+3] = 255
					srcIdx += 3
					dstIdx += 4
				}
			}
		}(startY, endY)
	}
	wg.Wait()
	return img
}

// makeSourceImage creates a deterministic in-memory test image for stable benchmark inputs.
func makeSourceImage(width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		row := img.Pix[y*img.Stride : y*img.Stride+width*4]
		for x := 0; x < width; x++ {
			off := x * 4
			row[off] = byte((x*17 + y*3) & 0xff)
			row[off+1] = byte((x*7 + y*29) & 0xff)
			row[off+2] = byte((x*31 + y*11) & 0xff)
			row[off+3] = 255
		}
	}
	return img
}

// benchmarkDecodeBGRToNRGBA is a micro-benchmark that isolates BGR->NRGBA conversion cost only.
// It excludes native calls and focuses on conversion strategy tradeoffs.
func benchmarkDecodeBGRToNRGBA(b *testing.B, width, height int) {
	raw := make([]byte, width*height*3)
	for i := range raw {
		raw[i] = byte((i*37 + 19) & 0xff)
	}

	b.Run("SetNRGBA", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchmarkImageSink = decodeBGRToNRGBASetNRGBA(raw, width, height)
		}
	})

	b.Run("DirectPix", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchmarkImageSink = decodeBGRToNRGBADirectPix(raw, width, height)
		}
	})

	b.Run("Uint32", func(b *testing.B) {
		if !isLittleEndian() {
			b.Skip("requires little-endian byte order")
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchmarkImageSink = decodeBGRToNRGBAUint32(raw, width, height)
		}
	})

	workers := runtime.GOMAXPROCS(0)
	b.Run("DirectPixParallel", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchmarkImageSink = decodeBGRToNRGBADirectPixParallel(raw, width, height, workers)
		}
	})

	b.Run("DirectPixReuseDst", func(b *testing.B) {
		dst := image.NewNRGBA(image.Rect(0, 0, width, height))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			decodeBGRToNRGBADirectPixInto(raw, dst, width, height)
			benchmarkImageSink = dst
		}
	})
}

// benchmarkImageBufferGet measures end-to-end ImageBuffer.Get (current implementation).
func benchmarkImageBufferGet(b *testing.B, width, height int) {
	imageBuffer := NewImageBuffer()
	if imageBuffer == nil {
		b.Skip("failed to create image buffer")
	}
	defer imageBuffer.Destroy()

	src := makeSourceImage(width, height)
	if ok := imageBuffer.Set(src); !ok {
		b.Fatal("failed to set source image")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got := imageBuffer.Get()
		if got == nil {
			b.Fatal("Get returned nil")
		}
		benchmarkImageSink = got
	}
}

// benchmarkImageBufferGetLegacy measures end-to-end behavior equivalent to the historical Get path.
// Use this as "before" baseline when comparing current optimizations.
func benchmarkImageBufferGetLegacy(b *testing.B, width, height int) {
	imageBuffer := NewImageBuffer()
	if imageBuffer == nil {
		b.Skip("failed to create image buffer")
	}
	defer imageBuffer.Destroy()

	src := makeSourceImage(width, height)
	if ok := imageBuffer.Set(src); !ok {
		b.Fatal("failed to set source image")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rawData := imageBuffer.getRawData()
		if rawData == nil {
			b.Fatal("getRawData returned nil")
		}
		w := int(imageBuffer.getWidth())
		h := int(imageBuffer.getHeight())
		raw := unsafe.Slice((*byte)(rawData), w*h*3)
		benchmarkImageSink = decodeBGRToNRGBASetNRGBA(raw, w, h)
	}
}

// benchmarkImageBufferGetReuseDst measures end-to-end GetInto with a reused destination buffer.
// This reflects the recommended low-allocation usage pattern in hot loops.
func benchmarkImageBufferGetReuseDst(b *testing.B, width, height int) {
	imageBuffer := NewImageBuffer()
	if imageBuffer == nil {
		b.Skip("failed to create image buffer")
	}
	defer imageBuffer.Destroy()

	src := makeSourceImage(width, height)
	if ok := imageBuffer.Set(src); !ok {
		b.Fatal("failed to set source image")
	}

	dst := image.NewNRGBA(image.Rect(0, 0, width, height))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got := imageBuffer.GetInto(dst)
		if got == nil {
			b.Fatal("GetInto returned nil")
		}
		benchmarkImageSink = got
	}
}

func BenchmarkDecodeBGRToNRGBA_1280x720(b *testing.B) {
	benchmarkDecodeBGRToNRGBA(b, 1280, 720)
}

func BenchmarkDecodeBGRToNRGBA_1920x1080(b *testing.B) {
	benchmarkDecodeBGRToNRGBA(b, 1920, 1080)
}

func BenchmarkImageBufferGet_1280x720(b *testing.B) {
	benchmarkImageBufferGet(b, 1280, 720)
}

func BenchmarkImageBufferGet_1920x1080(b *testing.B) {
	benchmarkImageBufferGet(b, 1920, 1080)
}

func BenchmarkImageBufferGetLegacy_1280x720(b *testing.B) {
	benchmarkImageBufferGetLegacy(b, 1280, 720)
}

func BenchmarkImageBufferGetLegacy_1920x1080(b *testing.B) {
	benchmarkImageBufferGetLegacy(b, 1920, 1080)
}

func BenchmarkImageBufferGetReuseDst_1280x720(b *testing.B) {
	benchmarkImageBufferGetReuseDst(b, 1280, 720)
}

func BenchmarkImageBufferGetReuseDst_1920x1080(b *testing.B) {
	benchmarkImageBufferGetReuseDst(b, 1920, 1080)
}
