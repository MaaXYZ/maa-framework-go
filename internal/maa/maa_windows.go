//go:build windows

package maa

import (
	"syscall"
	"unsafe"
)

func handleLibDir(libDir string) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setDllDirProc := kernel32.NewProc("SetDllDirectoryA")
	dirPtr, err := syscall.BytePtrFromString(libDir)
	if err != nil {
		panic(err)
	}

	ret, _, err := setDllDirProc.Call(uintptr(unsafe.Pointer(dirPtr)))
	if ret == 0 {
		panic(err)
	}
}

func openLibrary(name string) (uintptr, error) {
	handle, err := syscall.LoadLibrary(name)
	return uintptr(handle), err
}

func unloadLibrary(handle uintptr) error {
	dllHandle := (syscall.Handle)(handle)
	return syscall.FreeLibrary(dllHandle)
}
