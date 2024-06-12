package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"

type StringListBuffer interface {
	Destroy()
	IsEmpty() bool
	Clear() bool
	Size() uint64
	Get(index uint64) string
	Append(value StringBuffer) bool
	Remove(index uint64) bool
}

type stringList struct {
	handle C.MaaStringListBufferHandle
}

func NewStringList() StringListBuffer {
	handle := C.MaaCreateStringListBuffer()
	return &stringList{handle: handle}
}

func (sl *stringList) Destroy() {
	C.MaaDestroyStringListBuffer(sl.handle)
}

func (sl *stringList) IsEmpty() bool {
	return C.MaaIsStringListEmpty(sl.handle) != 0
}

func (sl *stringList) Clear() bool {
	return C.MaaClearStringList(sl.handle) != 0
}

func (sl *stringList) Size() uint64 {
	return uint64(C.MaaGetStringListSize(sl.handle))
}

func (sl *stringList) Get(index uint64) string {
	handle := C.MaaGetStringListAt(sl.handle, C.uint64_t(index))
	str := &stringBuffer{handle: handle}
	return str.Get()
}

func (sl *stringList) Append(value StringBuffer) bool {
	return C.MaaStringListAppend(sl.handle, C.MaaStringBufferHandle(value.Handle())) != 0
}

func (sl *stringList) Remove(index uint64) bool {
	return C.MaaStringListRemove(sl.handle, C.uint64_t(index)) != 0
}
