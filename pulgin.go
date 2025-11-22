package maa

import "github.com/MaaXYZ/maa-framework-go/v3/internal/native"

func LoadPlugin(path string) bool {
	return native.MaaGlobalLoadPlugin(path)
}
