package maa

import "github.com/MaaXYZ/maa-framework-go/v2/internal/maa"

func LoadPlugin(path string) bool {
	return maa.MaaGlobalLoadPlugin(path)
}
