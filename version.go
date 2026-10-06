package maa

import "github.com/MaaXYZ/maa-framework-go/v4/internal/native"

// Version returns the version of the loaded MaaFramework library.
func Version() string {
	return native.MaaVersion()
}
