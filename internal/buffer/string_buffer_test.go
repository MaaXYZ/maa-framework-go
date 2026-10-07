package buffer

import (
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

func createStringBuffer(t *testing.T) *StringBuffer {
	stringBuffer := NewStringBuffer()
	require.NotNil(t, stringBuffer)
	return stringBuffer
}

func TestNewStringBuffer(t *testing.T) {
	stringBuffer := createStringBuffer(t)
	stringBuffer.Destroy()
}

func TestStringBuffer_Handle(t *testing.T) {
	stringBuffer := createStringBuffer(t)
	defer stringBuffer.Destroy()
	handle := stringBuffer.Handle()
	require.NotNil(t, handle)
}

func TestStringBuffer_IsEmpty(t *testing.T) {
	stringBuffer := createStringBuffer(t)
	defer stringBuffer.Destroy()
	got := stringBuffer.IsEmpty()
	require.True(t, got)
}

func TestStringBuffer_Clear(t *testing.T) {
	stringBuffer := createStringBuffer(t)
	defer stringBuffer.Destroy()
	got := stringBuffer.Clear()
	require.True(t, got)
}

func TestStringBuffer_Set(t *testing.T) {
	stringBuffer := createStringBuffer(t)
	defer stringBuffer.Destroy()
	str1 := "test"
	got := stringBuffer.Set(str1)
	require.True(t, got)

	str2 := stringBuffer.Get()
	require.Equal(t, str1, str2)
}

func TestStringBuffer_Size(t *testing.T) {
	stringBuffer := createStringBuffer(t)
	defer stringBuffer.Destroy()
	str := "test"
	got := stringBuffer.Set(str)
	require.True(t, got)

	size := stringBuffer.Size()
	require.Equal(t, uint64(len(str)), size)
}

func TestStringBuffer_SetWithSize(t *testing.T) {
	stringBuffer := createStringBuffer(t)
	defer stringBuffer.Destroy()
	str1 := "test"
	got := stringBuffer.SetWithSize(str1, uint64(len(str1)))
	require.True(t, got)

	str2 := stringBuffer.Get()
	require.Equal(t, str1, str2)
}

// TestStringBuffer_EmptyGetAndSize pins the empty-buffer contract: Get reads
// an empty string and Size reports zero until something is written.
func TestStringBuffer_EmptyGetAndSize(t *testing.T) {
	stringBuffer := createStringBuffer(t)
	defer stringBuffer.Destroy()

	require.Empty(t, stringBuffer.Get())
	require.Equal(t, uint64(0), stringBuffer.Size())
	require.True(t, stringBuffer.IsEmpty())

	require.True(t, stringBuffer.Set("x"))
	require.False(t, stringBuffer.IsEmpty())
}

// TestStringBuffer_NULSemantics pins the embedded-NUL contract against the
// native conversion semantics: Set goes through std::string(const char*) and
// truncates at the first NUL, while SetWithSize copies exactly size bytes and
// keeps embedded NULs. Get always truncates at the first NUL on the Go read
// side, so the preserved bytes are observed through Size.
func TestStringBuffer_NULSemantics(t *testing.T) {
	stringBuffer := createStringBuffer(t)
	defer stringBuffer.Destroy()

	require.True(t, stringBuffer.Set("a\x00b"))
	require.Equal(t, "a", stringBuffer.Get())
	require.Equal(t, uint64(1), stringBuffer.Size())

	require.True(t, stringBuffer.SetWithSize("a\x00b", 3))
	require.Equal(t, uint64(3), stringBuffer.Size())
	require.Equal(t, "a", stringBuffer.Get())

	require.True(t, stringBuffer.SetWithSize("abcdef", 3))
	require.Equal(t, uint64(3), stringBuffer.Size())
	require.Equal(t, "abc", stringBuffer.Get())
}

// TestStringBuffer_SetWithSize_RejectsOversize pins the wrapper guard that
// keeps the native side from reading past the string allocation: size must
// not exceed len(str), and a rejected call leaves the buffer unchanged.
func TestStringBuffer_SetWithSize_RejectsOversize(t *testing.T) {
	stringBuffer := createStringBuffer(t)
	defer stringBuffer.Destroy()

	require.False(t, stringBuffer.SetWithSize("abc", 1000))
	require.True(t, stringBuffer.IsEmpty())

	require.False(t, stringBuffer.SetWithSize("abc", 4))
	require.True(t, stringBuffer.IsEmpty())

	require.True(t, stringBuffer.SetWithSize("abc", 3))
	require.False(t, stringBuffer.IsEmpty())

	require.True(t, stringBuffer.Set("ok"))
	require.Equal(t, "ok", stringBuffer.Get())
}

// TestStringBuffer_SetWithSize_TrailingNUL exercises the purego zero-copy
// path taken when the string already ends in NUL; it must behave like the
// copying path.
func TestStringBuffer_SetWithSize_TrailingNUL(t *testing.T) {
	stringBuffer := createStringBuffer(t)
	defer stringBuffer.Destroy()

	str := "abc\x00"
	require.True(t, stringBuffer.SetWithSize(str, uint64(len(str))))
	require.Equal(t, uint64(4), stringBuffer.Size())
	require.Equal(t, "abc", stringBuffer.Get())

	require.False(t, stringBuffer.SetWithSize(str, uint64(len(str)+1)))
	require.Equal(t, uint64(4), stringBuffer.Size())
}

// TestStringBuffer_ByHandle pins the borrowed-handle wrapper contract: the
// ByHandle wrapper shares the owner's native buffer, writes through it are
// visible to the owner, and the owner's Destroy is the only destroy call.
func TestStringBuffer_ByHandle(t *testing.T) {
	stringBuffer := createStringBuffer(t)

	borrowed := NewStringBufferByHandle(stringBuffer.Handle())
	require.NotNil(t, borrowed)
	require.Equal(t, stringBuffer.Handle(), borrowed.Handle())
	require.True(t, borrowed.Set("via borrowed"))
	require.Equal(t, "via borrowed", stringBuffer.Get())

	destroys := 0
	oldDestroy := native.MaaStringBufferDestroy
	defer func() { native.MaaStringBufferDestroy = oldDestroy }()
	native.MaaStringBufferDestroy = func(handle uintptr) { destroys++ }

	stringBuffer.Destroy()
	require.Equal(t, 1, destroys)

	// Run the real destroy so the counted call does not leak the object.
	native.MaaStringBufferDestroy = oldDestroy
	native.MaaStringBufferDestroy(stringBuffer.Handle())
}
