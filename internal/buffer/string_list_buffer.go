package buffer

import (
	"github.com/MaaXYZ/maa-framework-go/internal/maa"
)

type StringListBuffer struct {
	handle uintptr
}

func NewStringListBuffer() *StringListBuffer {
	handle := maa.MaaStringListBufferCreate()
	if handle == 0 {
		return nil
	}
	return &StringListBuffer{
		handle: handle,
	}
}

func NewStringListBufferByHandle(handle uintptr) *StringListBuffer {
	return &StringListBuffer{
		handle: handle,
	}
}

func (sl *StringListBuffer) Destroy() {
	maa.MaaStringListBufferDestroy(sl.handle)
}

func (sl *StringListBuffer) Handle() uintptr {
	return sl.handle
}

func (sl *StringListBuffer) IsEmpty() bool {
	return maa.MaaStringListBufferIsEmpty(sl.handle)
}

func (sl *StringListBuffer) Clear() bool {
	return maa.MaaStringListBufferClear(sl.handle)
}

func (sl *StringListBuffer) Size() uint64 {
	return maa.MaaStringListBufferSize(sl.handle)
}

func (sl *StringListBuffer) Get(index uint64) string {
	handle := maa.MaaStringListBufferAt(sl.handle, index)
	str := &StringBuffer{handle: handle}
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
	return maa.MaaStringListBufferAppend(sl.handle, value.handle)
}

func (sl *StringListBuffer) Remove(index uint64) bool {
	return maa.MaaStringListBufferRemove(sl.handle, index)
}
