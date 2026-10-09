package buffer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func createStringListBuffer(t *testing.T) *StringListBuffer {
	stringListBuffer := NewStringListBuffer()
	require.NotNil(t, stringListBuffer)
	return stringListBuffer
}

func TestNewStringListBuffer(t *testing.T) {
	stringListBuffer := createStringListBuffer(t)
	stringListBuffer.Destroy()
}

func TestStringListBuffer_Handle(t *testing.T) {
	stringListBuffer := createStringListBuffer(t)
	defer stringListBuffer.Destroy()
	handle := stringListBuffer.Handle()
	require.NotNil(t, handle)
}

func TestStringListBuffer_IsEmpty(t *testing.T) {
	stringListBuffer := createStringListBuffer(t)
	defer stringListBuffer.Destroy()
	got := stringListBuffer.IsEmpty()
	require.True(t, got)
}

func TestStringListBuffer_Clear(t *testing.T) {
	stringListBuffer := createStringListBuffer(t)
	defer stringListBuffer.Destroy()
	got := stringListBuffer.Clear()
	require.True(t, got)
}

func TestStringListBuffer_Append(t *testing.T) {
	stringListBuffer := createStringListBuffer(t)
	defer stringListBuffer.Destroy()

	stringBuffer := createStringBuffer(t)
	str1 := "test"
	got1 := stringBuffer.Set(str1)
	require.True(t, got1)

	got2 := stringListBuffer.Append(stringBuffer)
	require.True(t, got2)

	got3 := stringListBuffer.IsEmpty()
	require.False(t, got3)

	str2 := stringListBuffer.Get(0)
	require.Equal(t, str1, str2)
}

func TestStringListBuffer_Remove(t *testing.T) {
	stringListBuffer := createStringListBuffer(t)
	defer stringListBuffer.Destroy()

	stringBuffer := createStringBuffer(t)
	str1 := "test"
	got1 := stringBuffer.Set(str1)
	require.True(t, got1)

	got2 := stringListBuffer.Append(stringBuffer)
	require.True(t, got2)

	removed := stringListBuffer.Remove(0)
	require.True(t, removed)

	got3 := stringListBuffer.IsEmpty()
	require.True(t, got3)
}

func TestStringListBuffer_Size(t *testing.T) {
	stringListBuffer := createStringListBuffer(t)
	defer stringListBuffer.Destroy()

	stringBuffer := createStringBuffer(t)
	str1 := "test"
	got1 := stringBuffer.Set(str1)
	require.True(t, got1)

	got2 := stringListBuffer.Append(stringBuffer)
	require.True(t, got2)

	got3 := stringListBuffer.IsEmpty()
	require.False(t, got3)

	size := stringListBuffer.Size()
	require.Equal(t, uint64(1), size)
}

func TestStringListBuffer_GetAll(t *testing.T) {
	stringListBuffer := createStringListBuffer(t)
	defer stringListBuffer.Destroy()

	stringBuffer := createStringBuffer(t)
	str1 := "test"
	got1 := stringBuffer.Set(str1)
	require.True(t, got1)

	got2 := stringListBuffer.Append(stringBuffer)
	require.True(t, got2)

	got3 := stringListBuffer.IsEmpty()
	require.False(t, got3)

	list := stringListBuffer.GetAll()
	require.Len(t, list, 1)
}

// TestStringListBuffer_AppendNilReturnsFalse pins the null-value contract:
// upstream rejects a null value with false instead of crashing.
func TestStringListBuffer_AppendNilReturnsFalse(t *testing.T) {
	stringListBuffer := createStringListBuffer(t)
	defer stringListBuffer.Destroy()

	require.False(t, stringListBuffer.Append(nil))
	require.True(t, stringListBuffer.IsEmpty())
	require.Equal(t, uint64(0), stringListBuffer.Size())
}

// TestStringListBuffer_AppendDeepCopyOwnership pins the Append ownership
// contract: the list stores a deep copy, so destroying the source buffer
// afterwards must not affect the stored element.
func TestStringListBuffer_AppendDeepCopyOwnership(t *testing.T) {
	stringListBuffer := createStringListBuffer(t)
	defer stringListBuffer.Destroy()

	stringBuffer := createStringBuffer(t)
	require.True(t, stringBuffer.Set("owned by list"))
	require.True(t, stringListBuffer.Append(stringBuffer))

	stringBuffer.Destroy()

	require.Equal(t, "owned by list", stringListBuffer.Get(0))
}

// TestStringListBuffer_OutOfRange pins the out-of-range degradation: upstream
// logs and returns null/false, which the wrapper surfaces as an empty string
// from Get and false from Remove, indistinguishable from an empty element.
func TestStringListBuffer_OutOfRange(t *testing.T) {
	stringListBuffer := createStringListBuffer(t)
	defer stringListBuffer.Destroy()

	require.Empty(t, stringListBuffer.Get(0))
	require.False(t, stringListBuffer.Remove(0))

	require.True(t, stringListBuffer.Append(createStringBufferWith(t, "only")))
	require.Empty(t, stringListBuffer.Get(1))
	require.False(t, stringListBuffer.Remove(1))
	require.Equal(t, "only", stringListBuffer.Get(0))
}

// TestStringListBuffer_GetAllContentAndOrder pins GetAll element content,
// ordering, and the non-nil empty slice on a fresh list.
func TestStringListBuffer_GetAllContentAndOrder(t *testing.T) {
	stringListBuffer := createStringListBuffer(t)
	defer stringListBuffer.Destroy()

	empty := stringListBuffer.GetAll()
	require.NotNil(t, empty)
	require.Empty(t, empty)

	for _, s := range []string{"a", "b", "c"} {
		require.True(t, stringListBuffer.Append(createStringBufferWith(t, s)))
	}
	require.Equal(t, []string{"a", "b", "c"}, stringListBuffer.GetAll())
}

// createStringBufferWith returns a new string buffer holding s. The buffer is
// destroyed when the test finishes; Append stores a deep copy, so passing it
// to a list is safe either way.
func createStringBufferWith(t *testing.T, s string) *StringBuffer {
	t.Helper()
	stringBuffer := createStringBuffer(t)
	t.Cleanup(stringBuffer.Destroy)
	require.True(t, stringBuffer.Set(s))
	return stringBuffer
}
