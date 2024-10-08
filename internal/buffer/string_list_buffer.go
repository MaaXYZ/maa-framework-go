package buffer

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import "unsafe"

type StringListBuffer struct {
	handle *C.MaaStringListBuffer
}

func NewStringListBuffer() *StringListBuffer {
	handle := C.MaaStringListBufferCreate()
	if handle == nil {
		return nil
	}
	return &StringListBuffer{
		handle: handle,
	}
}

func NewStringListBufferByHandle(handle unsafe.Pointer) *StringListBuffer {
	return &StringListBuffer{
		handle: (*C.MaaStringListBuffer)(handle),
	}
}

func (sl *StringListBuffer) Destroy() {
	C.MaaStringListBufferDestroy(sl.handle)
}

func (sl *StringListBuffer) Handle() unsafe.Pointer {
	return unsafe.Pointer(sl.handle)
}

func (sl *StringListBuffer) IsEmpty() bool {
	return C.MaaStringListBufferIsEmpty(sl.handle) != 0
}

func (sl *StringListBuffer) Clear() bool {
	return C.MaaStringListBufferClear(sl.handle) != 0
}

func (sl *StringListBuffer) Size() uint64 {
	return uint64(C.MaaStringListBufferSize(sl.handle))
}

func (sl *StringListBuffer) Get(index uint64) string {
	handle := C.MaaStringListBufferAt(sl.handle, C.uint64_t(index))
	str := &StringBuffer{
		handle: handle,
	}
	return str.Get()
}

func (sl *StringListBuffer) GetAll() []string {
	size := sl.Size()
	strings := make([]string, size)
	for i := uint64(0); i < size; i++ {
		strings[i] = sl.Get(i)
	}
	return strings
}

func (sl *StringListBuffer) Append(value *StringBuffer) bool {
	return C.MaaStringListBufferAppend(
		sl.handle,
		(*C.MaaStringBuffer)(value.Handle()),
	) != 0
}

func (sl *StringListBuffer) Remove(index uint64) bool {
	return C.MaaStringListBufferRemove(
		sl.handle,
		C.uint64_t(index),
	) != 0
}
