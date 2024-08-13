package buffer

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import "unsafe"

type StringBuffer struct {
	handle C.MaaStringBufferHandle
}

func NewStringBuffer() *StringBuffer {
	handle := C.MaaCreateStringBuffer()
	return &StringBuffer{
		handle: handle,
	}
}

func NewStringBufferByHandle(handle unsafe.Pointer) *StringBuffer {
	return &StringBuffer{
		handle: C.MaaStringBufferHandle(handle),
	}
}

func (s *StringBuffer) Destroy() {
	C.MaaDestroyStringBuffer(s.handle)
}

func (s *StringBuffer) Handle() unsafe.Pointer {
	return unsafe.Pointer(s.handle)
}

func (s *StringBuffer) IsEmpty() bool {
	return C.MaaIsStringEmpty(s.handle) != 0
}

func (s *StringBuffer) Clear() bool {
	return C.MaaClearString(s.handle) != 0
}

func (s *StringBuffer) Get() string {
	return C.GoString(C.MaaGetString(s.handle))
}

func (s *StringBuffer) Size() uint64 {
	return uint64(C.MaaGetStringSize(s.handle))
}

func (s *StringBuffer) Set(str string) bool {
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr))
	return C.MaaSetString(
		s.handle,
		cStr,
	) != 0
}

func (s *StringBuffer) SetWithSize(str string, size uint64) bool {
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr))
	return C.MaaSetStringEx(
		s.handle,
		cStr,
		C.uint64_t(size),
	) != 0
}
