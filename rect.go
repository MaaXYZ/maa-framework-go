package maa

import "github.com/MaaXYZ/maa-framework-go/buffer"

type Rect struct {
	X, Y, W, H int32
}

func toBufferRect(rect Rect) buffer.Rect {
	return buffer.Rect{
		X: rect.X,
		Y: rect.Y,
		W: rect.W,
		H: rect.H,
	}
}
