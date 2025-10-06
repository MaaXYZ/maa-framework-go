//go:build darwin || linux

package maa

import "github.com/ebitengine/purego"

func handleLibDir(libDir string) {
	// do nothing
}

func openLibrary(name string) (uintptr, error) {
	return purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
}
