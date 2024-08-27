package buffer

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import "unsafe"

type StringListBuffer struct {
	handle C.MaaStringListBufferHandle
}

func NewStringListBuffer() *StringListBuffer {
	handle := C.MaaCreateStringListBuffer()
	return &StringListBuffer{
		handle: handle,
	}
}

func NewStringListBufferByHandle(handle unsafe.Pointer) *StringListBuffer {
	return &StringListBuffer{
		handle: C.MaaStringListBufferHandle(handle),
	}
}

func (sl *StringListBuffer) Destroy() {
	C.MaaDestroyStringListBuffer(sl.handle)
}

func (sl *StringListBuffer) IsEmpty() bool {
	return C.MaaIsStringListEmpty(sl.handle) != 0
}

func (sl *StringListBuffer) Clear() bool {
	return C.MaaClearStringList(sl.handle) != 0
}

func (sl *StringListBuffer) Size() uint64 {
	return uint64(C.MaaGetStringListSize(sl.handle))
}

func (sl *StringListBuffer) Get(index uint64) string {
	handle := C.MaaGetStringListAt(sl.handle, C.uint64_t(index))
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

func (sl *StringListBuffer) Append(value StringBuffer) bool {
	return C.MaaStringListAppend(
		sl.handle,
		C.MaaStringBufferHandle(value.Handle()),
	) != 0
}

func (sl *StringListBuffer) Remove(index uint64) bool {
	return C.MaaStringListRemove(
		sl.handle,
		C.uint64_t(index),
	) != 0
}