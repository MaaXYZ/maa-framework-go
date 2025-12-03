package maa

import (
	"errors"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/native"
)

var (
	inited bool

	ErrAlreadyInitialized = errors.New("maa framework already initialized")
	ErrNotInitialized     = errors.New("maa framework not initialized")
)

// InitConfig contains configuration options for initializing the MAA framework.
// It specifies various settings that control the framework's behavior,
// logging, debugging, and resource locations.
type InitConfig struct {
	// LibDir specifies the directory path where MAA dynamic libraries are located.
	// If empty, the framework will attempt to locate libraries in default paths.
	LibDir string

	// LogDir specifies the directory where log files will be written.
	// Defaults to "./debug" if not specified.
	LogDir string

	// SaveDraw controls whether to save recognition results to LogDir/vision.
	// When enabled, RecoDetail will be able to retrieve draws for debugging purposes.
	SaveDraw bool

	// StdoutLevel sets the logging verbosity level for standard output.
	// Controls which log messages are displayed on the console.
	StdoutLevel LoggingLevel

	// DebugMode enables or disables comprehensive debug mode.
	// When enabled, additional debug information is collected and logged.
	DebugMode bool

	// PluginPaths specifies the paths to the plugins to load.
	// If empty, the framework will not load any plugins.
	PluginPaths []string
}

// InitOption defines a function type for configuring InitConfig through functional options.
// Each InitOption function modifies the InitConfig to set specific initialization parameters.
type InitOption func(*InitConfig)

// WithLibDir returns an InitOption that sets the library directory path for the MAA framework.
// The libDir parameter specifies the directory where the MAA dynamic library is located.
func WithLibDir(libDir string) InitOption {
	return func(ic *InitConfig) {
		ic.LibDir = libDir
	}
}

// WithLogDir returns an InitOption that sets the directory path for log files.
// The logDir parameter specifies where the MAA framework should write its log files.
func WithLogDir(logDir string) InitOption {
	return func(ic *InitConfig) {
		ic.LogDir = logDir
	}
}

// WithSaveDraw returns an InitOption that configures whether to save drawing information.
// When enabled is true, recognition results will be saved to LogDir/vision directory
// and RecoDetail will be able to retrieve draws for debugging.
func WithSaveDraw(enabled bool) InitOption {
	return func(ic *InitConfig) {
		ic.SaveDraw = enabled
	}
}

// WithStdoutLevel returns an InitOption that sets the logging level for standard output.
// The level parameter determines the verbosity of logs written to stdout.
func WithStdoutLevel(level LoggingLevel) InitOption {
	return func(ic *InitConfig) {
		ic.StdoutLevel = level
	}
}

// WithDebugMode returns an InitOption that enables or disables debug mode.
// When enabled is true, additional debug information will be collected and logged.
func WithDebugMode(enabled bool) InitOption {
	return func(ic *InitConfig) {
		ic.DebugMode = enabled
	}
}

func WithPluginPaths(path ...string) InitOption {
	return func(ic *InitConfig) {
		ic.PluginPaths = path
	}
}

func defaultInitConfig() InitConfig {
	return InitConfig{
		LibDir:      "",
		LogDir:      "./debug",
		SaveDraw:    false,
		StdoutLevel: LoggingLevelInfo,
		DebugMode:   false,
		PluginPaths: []string{},
	}
}

// Init loads the dynamic library related to the MAA framework and registers its related functions.
// It must be called before invoking any other MAA-related functions.
// Note: If this function is not called before other MAA functions, it will trigger a null pointer panic.
func Init(opts ...InitOption) error {

	if inited {
		return ErrAlreadyInitialized
	}

	cfg := defaultInitConfig()

	for _, opt := range opts {
		opt(&cfg)
	}

	if err := native.Init(cfg.LibDir); err != nil {
		return err
	}

	SetLogDir(cfg.LogDir)
	SetSaveDraw(cfg.SaveDraw)
	SetStdoutLevel(cfg.StdoutLevel)
	SetDebugMode(cfg.DebugMode)

	for _, path := range cfg.PluginPaths {
		LoadPlugin(path)
	}

	inited = true

	return nil
}

// IsInited checks if the MAA framework has been initialized.
func IsInited() bool {
	return inited
}

// Release releases the dynamic library resources of the MAA framework and unregisters its related functions.
// It must be called only after the framework has been initialized via Init.
func Release() error {

	if !inited {
		return ErrNotInitialized
	}

	if err := native.Release(); err != nil {
		return err
	}

	inited = false

	return nil
}

func setGlobalOption(key native.MaaGlobalOption, value unsafe.Pointer, valSize uintptr) bool {
	return native.MaaGlobalSetOption(key, value, uint64(valSize))
}

// SetLogDir sets the log directory.
func SetLogDir(path string) bool {
	if path == "" {
		return false
	}
	return setGlobalOption(native.MaaGlobalOption_LogDir, unsafe.Pointer(&[]byte(path)[0]), uintptr(len(path)))
}

// SetSaveDraw sets whether to save draw.
func SetSaveDraw(enabled bool) bool {
	return setGlobalOption(native.MaaGlobalOption_SaveDraw, unsafe.Pointer(&enabled), unsafe.Sizeof(enabled))
}

type LoggingLevel int32

// LoggingLevel
const (
	LoggingLevelOff LoggingLevel = iota
	LoggingLevelFatal
	LoggingLevelError
	LoggingLevelWarn
	LoggingLevelInfo
	LoggingLevelDebug
	LoggingLevelTrace
	LoggingLevelAll
)

// SetStdoutLevel sets the level of log output to stdout.
func SetStdoutLevel(level LoggingLevel) bool {
	return setGlobalOption(native.MaaGlobalOption_StdoutLevel, unsafe.Pointer(&level), unsafe.Sizeof(level))
}

// SetDebugMode sets whether to enable debug mode.
func SetDebugMode(enabled bool) bool {
	return setGlobalOption(native.MaaGlobalOption_DebugMode, unsafe.Pointer(&enabled), unsafe.Sizeof(enabled))
}

// LoadPlugin loads a plugin specified by path.
// The path may be a full filesystem path or just a plugin name.
// When only a name is provided, the function searches system directories and the current working directory for a matching plugin.
// If the path refers to a directory, plugins inside that directory are searched recursively.
func LoadPlugin(path string) bool {
	return native.MaaGlobalLoadPlugin(path)
}
