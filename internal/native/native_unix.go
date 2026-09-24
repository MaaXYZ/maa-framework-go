//go:build darwin || linux

package native

import "github.com/ebitengine/purego"

func handleLibDir(_ string) error {
	// do nothing
	return nil
}

// platformOpenLibrary opens a dynamic library using the platform loader.
func platformOpenLibrary(name string) (uintptr, error) {
	return purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
}

// platformUnloadLibrary releases an open dynamic library handle.
func platformUnloadLibrary(handle uintptr) error {
	return purego.Dlclose(handle)
}

// platformLookupSymbol resolves a symbol from an open dynamic library handle.
func platformLookupSymbol(handle uintptr, name string) (uintptr, error) {
	return purego.Dlsym(handle, name)
}
