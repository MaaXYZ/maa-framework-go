package buffer

import (
	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
)

// StringBuffer wraps a native string buffer holding a byte string. Create one
// with NewStringBuffer, or borrow an existing handle with
// NewStringBufferByHandle.
type StringBuffer struct {
	handle uintptr
}

// NewStringBuffer creates a new empty string buffer.
// It returns nil when the underlying native buffer cannot be created.
func NewStringBuffer() *StringBuffer {
	handle := native.MaaStringBufferCreate()
	if handle == 0 {
		return nil
	}
	return &StringBuffer{
		handle: handle,
	}
}

// NewStringBufferByHandle wraps an existing native string buffer handle
// without taking ownership. The caller keeps owning the handle and stays
// responsible for destroying it exactly once; the wrapper is a shared view of
// the same buffer.
func NewStringBufferByHandle(handle uintptr) *StringBuffer {
	return &StringBuffer{
		handle: handle,
	}
}

// Destroy releases the underlying native string buffer. The owner must call
// it exactly once and not use the wrapper afterwards; a borrowed wrapper must
// not destroy the shared handle.
func (s *StringBuffer) Destroy() {
	native.MaaStringBufferDestroy(s.handle)
}

// Handle returns the underlying native string buffer handle for passing to
// framework functions that take a string buffer.
func (s *StringBuffer) Handle() uintptr {
	return s.handle
}

// IsEmpty reports whether the buffer holds no content.
func (s *StringBuffer) IsEmpty() bool {
	return native.MaaStringBufferIsEmpty(s.handle)
}

// Clear discards the content and reports whether the native call succeeded.
func (s *StringBuffer) Clear() bool {
	return native.MaaStringBufferClear(s.handle)
}

// Get returns the buffer content as a Go string, truncating at the first NUL.
// Embedded NUL bytes stay in the buffer; read them back through Size and a
// native consumer.
func (s *StringBuffer) Get() string {
	return native.MaaStringBufferGet(s.handle)
}

// Size returns the content length in bytes, including any embedded NULs.
func (s *StringBuffer) Size() uint64 {
	return native.MaaStringBufferSize(s.handle)
}

// Set stores str through a NUL-terminated C string, so content is truncated
// at the first NUL. It reports whether the native call succeeded.
func (s *StringBuffer) Set(str string) bool {
	return native.MaaStringBufferSet(s.handle, str)
}

// SetWithSize stores exactly size bytes of str, preserving embedded NULs. It
// returns false when size exceeds len(str), which would make the native side
// read past the string, and otherwise reports whether the native call
// succeeded.
func (s *StringBuffer) SetWithSize(str string, size uint64) bool {
	// The native side copies exactly size bytes from the pointer the binding
	// layer allocates for len(str)+1 bytes; a larger size reads past it.
	if size > uint64(len(str)) {
		return false
	}
	return native.MaaStringBufferSetEx(s.handle, str, size)
}
