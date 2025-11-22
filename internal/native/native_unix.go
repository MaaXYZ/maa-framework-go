//go:build darwin || linux

package native

import "github.com/ebitengine/purego"

func handleLibDir(libDir string) {
	// do nothing
}

func openLibrary(name string) (uintptr, error) {
	return purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
}

func unloadLibrary(handle uintptr) error {
	return purego.Dlclose(handle)
}
