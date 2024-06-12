package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import "unsafe"

type StringBuffer interface {
	Destroy()
	Handle() unsafe.Pointer
	IsEmpty() bool
	Clear() bool
	Get() string
	Size() uint64
	Set(str string) bool
	SetWithSize(str string, size uint64) bool
}

type stringBuffer struct {
	handle C.MaaStringBufferHandle
}

func NewString() StringBuffer {
	handle := C.MaaCreateStringBuffer()
	return &stringBuffer{handle: handle}
}

func (s *stringBuffer) Destroy() {
	C.MaaDestroyStringBuffer(s.handle)
}

func (s *stringBuffer) Handle() unsafe.Pointer {
	return unsafe.Pointer(s.handle)
}

func (s *stringBuffer) IsEmpty() bool {
	return C.MaaIsStringEmpty(s.handle) != 0
}

func (s *stringBuffer) Clear() bool {
	return C.MaaClearString(s.handle) != 0
}

func (s *stringBuffer) Get() string {
	return C.GoString(C.MaaGetString(s.handle))
}

func (s *stringBuffer) Size() uint64 {
	return uint64(C.MaaGetStringSize(s.handle))
}

func (s *stringBuffer) Set(str string) bool {
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr))
	return C.MaaSetString(s.handle, cStr) != 0
}

func (s *stringBuffer) SetWithSize(str string, size uint64) bool {
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr))
	return C.MaaSetStringEx(s.handle, cStr, C.uint64_t(size)) != 0
}
