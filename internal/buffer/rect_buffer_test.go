package buffer

import (
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/rect"
	"github.com/stretchr/testify/require"
)

func createRectBuffer(t *testing.T) *RectBuffer {
	rectBuffer, err := NewRectBuffer()
	require.NoError(t, err)
	require.NotNil(t, rectBuffer)
	return rectBuffer
}

func TestNewRectBuffer(t *testing.T) {
	rectBuffer := createRectBuffer(t)
	rectBuffer.Destroy()
}

func TestNewRectBuffer_CreateFailure(t *testing.T) {
	oldCreate := native.MaaRectCreate
	defer func() { native.MaaRectCreate = oldCreate }()
	native.MaaRectCreate = func() uintptr { return 0 }

	rectBuffer, err := NewRectBuffer()
	require.Nil(t, rectBuffer)
	require.Error(t, err)
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
	err := rectBuffer.Set(rect1)
	require.NoError(t, err)

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

func TestRectBuffer_SetFailure(t *testing.T) {
	rectBuffer := createRectBuffer(t)
	defer rectBuffer.Destroy()

	oldSet := native.MaaRectSet
	defer func() { native.MaaRectSet = oldSet }()
	native.MaaRectSet = func(handle uintptr, x, y, w, h int32) bool { return false }

	err := rectBuffer.Set(rect.Rect{100, 200, 300, 400})
	require.Error(t, err)
}

// TestRectBuffer_ZeroValueAfterCreate pins that a freshly created buffer
// reads back the zero rect, matching the native zero-initialized MaaRect.
func TestRectBuffer_ZeroValueAfterCreate(t *testing.T) {
	rectBuffer := createRectBuffer(t)
	defer rectBuffer.Destroy()

	require.Equal(t, int32(0), rectBuffer.GetX())
	require.Equal(t, int32(0), rectBuffer.GetY())
	require.Equal(t, int32(0), rectBuffer.GetW())
	require.Equal(t, int32(0), rectBuffer.GetH())
	require.Equal(t, rect.Rect{0, 0, 0, 0}, rectBuffer.Get())
}

// TestRectBuffer_ByHandle pins the borrowed-handle wrapper contract: the
// ByHandle wrapper shares the owner's native buffer, writes through it are
// visible to the owner, and the owner's Destroy is the only destroy call.
func TestRectBuffer_ByHandle(t *testing.T) {
	rectBuffer := createRectBuffer(t)

	borrowed := NewRectBufferByHandle(rectBuffer.Handle())
	require.NotNil(t, borrowed)
	require.Equal(t, rectBuffer.Handle(), borrowed.Handle())
	require.NoError(t, borrowed.Set(rect.Rect{1, 2, 3, 4}))
	require.Equal(t, rect.Rect{1, 2, 3, 4}, rectBuffer.Get())

	destroys := 0
	oldDestroy := native.MaaRectDestroy
	defer func() { native.MaaRectDestroy = oldDestroy }()
	native.MaaRectDestroy = func(handle uintptr) { destroys++ }

	rectBuffer.Destroy()
	require.Equal(t, 1, destroys)
}
