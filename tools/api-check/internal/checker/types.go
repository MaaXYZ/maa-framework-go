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
	sectionPipeline         = "Pipeline V2 Coverage"
	sectionConstants        = "Constant Coverage"
	sectionEvents           = "Event Coverage"
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
var sectionOrder = []string{sectionNativeAPI, sectionController, sectionControllerMethod, sectionConstants, sectionEvents, sectionPipeline}

// Config selects the native headers and optional pipeline v2 schema to check.
// Relative paths are resolved against the detected repository root; absolute
// paths are used as-is.
type Config struct {
	HeaderDir string   `yaml:"header_dir"`
	Blacklist []string `yaml:"blacklist"`
	// NativeExclusions records intentional native symbol differences with
	// reasons. Keys are exact symbol names and reasons must be nonempty;
	// entries without a current difference are stale.
	NativeExclusions map[string]string `yaml:"native_exclusions"`
	// PipelineSchema enables pipeline v2 type and field-name coverage when set.
	// Use the schema supplied by the same MaaFramework release as HeaderDir.
	PipelineSchema string `yaml:"pipeline_schema"`
	// PipelineExclusions maps exact reported paths to intentional differences.
	// Each reason must be nonempty; exclusions without a difference are stale.
	PipelineExclusions map[string]string `yaml:"pipeline_exclusions"`
	// ConstantExclusions records native constants that the Go binding
	// deliberately does not mirror. Keys are exact C constant names such as
	// MaaInferenceExecutionProvider_CoreML, and reasons must be nonempty;
	// entries without a current difference are stale.
	ConstantExclusions map[string]string `yaml:"constant_exclusions"`
}

type stringSliceFlag []string

// String joins the accumulated values with commas for flag usage output.
func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ",")
}

// Set appends one trimmed value per flag occurrence; empty values are ignored.
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
	symbol  string
}

type methodSig struct {
	params  []string
	returns []string
}
