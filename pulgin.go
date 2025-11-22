package maa

import "github.com/MaaXYZ/maa-framework-go/v2/internal/native"

func LoadPlugin(path string) bool {
	return native.MaaGlobalLoadPlugin(path)
}
