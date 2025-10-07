package maa

import (
	"errors"

	"github.com/MaaXYZ/maa-framework-go/v2/internal/maa"
)

var (
	inited                bool
	ErrAlreadyInitialized = errors.New("maa framework already initialized")
)

type InitConfig struct {
	LibDir      string
	LogDir      string
	SaveDraw    bool
	StdoutLevel LoggingLevel
	DebugMode   bool
}

type InitOption func(*InitConfig)

func WithLibDir(libDir string) InitOption {
	return func(ic *InitConfig) {
		ic.LibDir = libDir
	}
}

func WithLogDir(logDir string) InitOption {
	return func(ic *InitConfig) {
		ic.LogDir = logDir
	}
}

func WithSaveDraw(enabled bool) InitOption {
	return func(ic *InitConfig) {
		ic.SaveDraw = enabled
	}
}

func WithStdoutLevel(level LoggingLevel) InitOption {
	return func(ic *InitConfig) {
		ic.StdoutLevel = level
	}
}

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
