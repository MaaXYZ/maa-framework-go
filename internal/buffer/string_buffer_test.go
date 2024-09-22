package buffer

import (
	"github.com/stretchr/testify/require"
	"testing"
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
