package maa

import "github.com/MaaXYZ/maa-framework-go/v3/internal/native"

// LoadPlugin loads a plugin specified by path.
// The path may be a full filesystem path or just a plugin name.
// When only a name is provided, the function searches system directories and the current working directory for a matching plugin.
// If the path refers to a directory, plugins inside that directory are searched recursively.
func LoadPlugin(path string) bool {
	return native.MaaGlobalLoadPlugin(path)
}
