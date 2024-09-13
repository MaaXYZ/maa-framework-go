package buffer

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import "unsafe"

type StringBuffer struct {
	handle *C.MaaStringBuffer
}

func NewStringBuffer() *StringBuffer {
	handle := C.MaaStringBufferCreate()
	return &StringBuffer{
		handle: handle,
	}
}

func NewStringBufferByHandle(handle unsafe.Pointer) *StringBuffer {
	return &StringBuffer{
		handle: (*C.MaaStringBuffer)(handle),
	}
}

func (s *StringBuffer) Destroy() {
	C.MaaStringBufferDestroy(s.handle)
}

func (s *StringBuffer) Handle() unsafe.Pointer {
	return unsafe.Pointer(s.handle)
}

func (s *StringBuffer) IsEmpty() bool {
	return C.MaaStringBufferIsEmpty(s.handle) != 0
}

func (s *StringBuffer) Clear() bool {
	return C.MaaStringBufferClear(s.handle) != 0
}

func (s *StringBuffer) Get() string {
	return C.GoString(C.MaaStringBufferGet(s.handle))
}

func (s *StringBuffer) Size() uint64 {
	return uint64(C.MaaStringBufferSize(s.handle))
}

func (s *StringBuffer) Set(str string) bool {
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr))
	return C.MaaStringBufferSet(
		s.handle,
		cStr,
	) != 0
}

func (s *StringBuffer) SetWithSize(str string, size uint64) bool {
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr))
	return C.MaaStringBufferSetEx(
		s.handle,
		cStr,
		C.uint64_t(size),
	) != 0
}
