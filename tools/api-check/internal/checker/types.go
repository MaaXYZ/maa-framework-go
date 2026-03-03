package checker

import "strings"

const (
	repoRootModulePath = "github.com/MaaXYZ/maa-framework-go/v4"
	autoConfigFileName = "config.yaml"
)

const (
	sectionNativeAPI        = "Native API Coverage"
	sectionController       = "CustomController Consistency"
	sectionControllerMethod = "Controller Method Coverage"
	defaultHeaderDirRel     = "deps/include"
	customControllerRel     = "custom_controller.go"
	controllerHeaderRel     = "MaaFramework/Instance/MaaCustomController.h"
	maaDefHeaderRel         = "MaaFramework/MaaDef.h"
	adbControllerRel        = "controller/adb/adb.go"
	win32ControllerRel      = "controller/win32/win32.go"
	apiCheckConfigPathRel   = "tools/api-check/config.yaml"
)

var nativeFilesByModule = map[string][]string{
	"framework": {
		"internal/native/framework.go",
	},
	"toolkit": {
		"internal/native/toolkit.go",
	},
	"agent_server": {
		"internal/native/agent_server.go",
	},
	"agent_client": {
		"internal/native/agent_client.go",
	},
}

var moduleOrder = []string{"framework", "toolkit", "agent_server", "agent_client"}
var sectionOrder = []string{sectionNativeAPI, sectionController, sectionControllerMethod}

type Config struct {
	HeaderDir string   `yaml:"header_dir"`
	Blacklist []string `yaml:"blacklist"`
}

type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSliceFlag) Set(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	*s = append(*s, trimmed)
	return nil
}

type issue struct {
	section string
	message string
}

type methodSig struct {
	params  []string
	returns []string
}
