package buffer

import (
	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
)

// StringListBuffer wraps a native list of string buffers. Create one with
// NewStringListBuffer, or borrow an existing handle with
// NewStringListBufferByHandle.
type StringListBuffer struct {
	handle uintptr
}

// NewStringListBuffer creates a new empty string list buffer.
// It returns nil when the underlying native buffer cannot be created.
func NewStringListBuffer() *StringListBuffer {
	handle := native.MaaStringListBufferCreate()
	if handle == 0 {
		return nil
	}
	return &StringListBuffer{
		handle: handle,
	}
}

// NewStringListBufferByHandle wraps a borrowed native string list buffer handle.
func NewStringListBufferByHandle(handle uintptr) *StringListBuffer {
	return &StringListBuffer{
		handle: handle,
	}
}

// Destroy releases the underlying native string list buffer.
func (sl *StringListBuffer) Destroy() {
	native.MaaStringListBufferDestroy(sl.handle)
}

// Handle returns the underlying native list buffer handle for passing to
// framework functions that take a string list buffer.
func (sl *StringListBuffer) Handle() uintptr {
	return sl.handle
}

// IsEmpty reports whether the list holds no strings.
func (sl *StringListBuffer) IsEmpty() bool {
	return native.MaaStringListBufferIsEmpty(sl.handle)
}

// Clear removes all strings and reports whether the native call succeeded.
func (sl *StringListBuffer) Clear() bool {
	return native.MaaStringListBufferClear(sl.handle)
}

// Size returns the number of strings in the list.
func (sl *StringListBuffer) Size() uint64 {
	return native.MaaStringListBufferSize(sl.handle)
}

// Get returns the string at index. It returns an empty string when index is
// out of range, indistinguishable from an empty element.
func (sl *StringListBuffer) Get(index uint64) string {
	handle := native.MaaStringListBufferAt(sl.handle, index)
	str := &StringBuffer{handle: handle}
	return str.Get()
}

// GetAll returns all strings in list order. It returns a non-nil empty slice
// when the list is empty.
func (sl *StringListBuffer) GetAll() []string {
	size := sl.Size()
	strings := make([]string, size)
	for i := uint64(0); i < size; i++ {
		strings[i] = sl.Get(i)
	}
	return strings
}

// Append stores a deep copy of value and reports whether the native call
// succeeded; the caller keeps owning value. It returns false when value is
// nil.
func (sl *StringListBuffer) Append(value *StringBuffer) bool {
	// Upstream returns false for a null value instead of crashing.
	if value == nil {
		return false
	}
	return native.MaaStringListBufferAppend(sl.handle, value.handle)
}

// Remove deletes the string at index and reports whether index was in range.
func (sl *StringListBuffer) Remove(index uint64) bool {
	return native.MaaStringListBufferRemove(sl.handle, index)
}
