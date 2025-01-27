package buffer

import (
	"github.com/MaaXYZ/maa-framework-go/v2/internal/maa"
)

type StringBuffer struct {
	handle uintptr
}

func NewStringBuffer() *StringBuffer {
	handle := maa.MaaStringBufferCreate()
	if handle == 0 {
		return nil
	}
	return &StringBuffer{
		handle: handle,
	}
}

func NewStringBufferByHandle(handle uintptr) *StringBuffer {
	return &StringBuffer{
		handle: handle,
	}
}

func (s *StringBuffer) Destroy() {
	maa.MaaStringBufferDestroy(s.handle)
}

func (s *StringBuffer) Handle() uintptr {
	return s.handle
}

func (s *StringBuffer) IsEmpty() bool {
	return maa.MaaStringBufferIsEmpty(s.handle)
}

func (s *StringBuffer) Clear() bool {
	return maa.MaaStringBufferClear(s.handle)
}

func (s *StringBuffer) Get() string {
	return maa.MaaStringBufferGet(s.handle)
}

func (s *StringBuffer) Size() uint64 {
	return maa.MaaStringBufferSize(s.handle)
}

func (s *StringBuffer) Set(str string) bool {
	return maa.MaaStringBufferSet(s.handle, str)
}

func (s *StringBuffer) SetWithSize(str string, size uint64) bool {
	return maa.MaaStringBufferSetEx(s.handle, str, size)
}
