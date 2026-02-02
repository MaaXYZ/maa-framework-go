package buffer

import (
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/rect"
	"github.com/stretchr/testify/require"
)

func createRectBuffer(t *testing.T) *RectBuffer {
	rectBuffer := NewRectBuffer()
	require.NotNil(t, rectBuffer)
	return rectBuffer
}

func TestNewRectBuffer(t *testing.T) {
	rectBuffer := createRectBuffer(t)
	rectBuffer.Destroy()
}

func TestRectBuffer_Handle(t *testing.T) {
	rectBuffer := createRectBuffer(t)
	defer rectBuffer.Destroy()
	handle := rectBuffer.Handle()
	require.NotNil(t, handle)
}

func TestRectBuffer_Set(t *testing.T) {
	rectBuffer := createRectBuffer(t)
	defer rectBuffer.Destroy()

	rect1 := rect.Rect{100, 200, 300, 400}
	got := rectBuffer.Set(rect1)
	require.True(t, got)

	x := rectBuffer.GetX()
	require.Equal(t, rect1.X(), int(x))
	y := rectBuffer.GetY()
	require.Equal(t, rect1.Y(), int(y))
	w := rectBuffer.GetW()
	require.Equal(t, rect1.Width(), int(w))
	h := rectBuffer.GetH()
	require.Equal(t, rect1.Height(), int(h))
	rect2 := rectBuffer.Get()
	require.Equal(t, rect1, rect2)
}
