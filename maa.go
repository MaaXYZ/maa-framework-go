package maa

import (
	"errors"

	"github.com/MaaXYZ/maa-framework-go/v2/internal/maa"
)

var (
	inited                bool
	ErrAlreadyInitialized = errors.New("maa framework already initialized")
	ErrNotInitialized     = errors.New("maa framework not initialized")
)

type InitConfig struct {
	LibDir      string
	LogDir      string
	SaveDraw    bool
	StdoutLevel LoggingLevel
	DebugMode   bool
}

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
// When enabled is true, the framework will save drawing debug information.
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

func defaultInitConfig() InitConfig {
	return InitConfig{
		LibDir:      "",
		LogDir:      "./debug",
		SaveDraw:    false,
		StdoutLevel: LoggingLevelInfo,
		DebugMode:   false,
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

	if err := maa.Init(cfg.LibDir); err != nil {
		return err
	}

	SetLogDir(cfg.LogDir)
	SetSaveDraw(cfg.SaveDraw)
	SetStdoutLevel(cfg.StdoutLevel)
	SetDebugMode(cfg.DebugMode)

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

	if err := maa.Release(); err != nil {
		return err
	}

	inited = false

	return nil
}
