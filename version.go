package maa

import "github.com/MaaXYZ/maa-framework-go/v2/internal/maa"

// Version returns the version of the maa framework.
func Version() string {
	return maa.MaaVersion()
}
