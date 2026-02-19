package checker

import "strings"

const (
	defaultHeaderDir = "deps/include"
	autoConfigPath   = "tools/api-check/config.yaml"
)

const (
	sectionNativeAPI    = "Native API Coverage"
	sectionController   = "CustomController Consistency"
	controllerHeaderRel = "MaaFramework/Instance/MaaCustomController.h"
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
var sectionOrder = []string{sectionNativeAPI, sectionController}

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
