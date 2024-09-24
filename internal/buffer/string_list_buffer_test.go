package buffer

import (
	"github.com/stretchr/testify/require"
	"testing"
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
