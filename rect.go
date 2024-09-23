package maa

import "github.com/MaaXYZ/maa-framework-go/internal/buffer"

type Rect struct {
	X, Y, W, H int32
}

func (r Rect) ToInts() [4]int32 {
	return [4]int32{r.X, r.Y, r.W, r.H}
}

func (r Rect) toBufferRect() buffer.Rect {
	return buffer.Rect{
		X: r.X,
		Y: r.Y,
		W: r.W,
		H: r.H,
	}
}

func toMaaRect(rect buffer.Rect) Rect {
	return Rect{
		X: rect.X,
		Y: rect.Y,
		W: rect.W,
		H: rect.H,
	}
}
