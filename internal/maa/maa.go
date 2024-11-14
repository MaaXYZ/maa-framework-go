//go:build darwin || linux || windows

package maa

import (
	"fmt"
	"runtime"
)

func getMaaFrameworkLibrary() string {
	switch runtime.GOOS {
	case "darwin":
		return "libMaaFramework.dylib"
	case "linux":
		return "libMaaFramework.so"
	case "windows":
		return "MaaFramework.dll"
	default:
		panic(fmt.Errorf("GOOS=%s is not supported", runtime.GOOS))
	}
}

func getMaaToolkitLibrary() string {
	switch runtime.GOOS {
	case "darwin":
		return "libMaaToolkit.dylib"
	case "linux":
		return "libMaaToolkit.so"
	case "windows":
		return "MaaToolkit.dll"
	default:
		panic(fmt.Errorf("GOOS=%s is not supported", runtime.GOOS))
	}
}
