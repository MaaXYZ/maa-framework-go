package maa

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
)

var (
	inited                  bool
	lifecycleMu             sync.RWMutex
	shutdownNativeLibraries = native.Shutdown

	// ErrSetLogDir is returned when SetLogDir fails at the native layer.
	ErrSetLogDir = errors.New("failed to set log directory")
	// ErrSetSaveDraw is returned when SetSaveDraw fails at the native layer.
	ErrSetSaveDraw = errors.New("failed to set save draw option")
	// ErrSetStdoutLevel is returned when SetStdoutLevel fails at the native layer.
	ErrSetStdoutLevel = errors.New("failed to set stdout level")
	// ErrSetDebugMode is returned when SetDebugMode fails at the native layer.
	ErrSetDebugMode = errors.New("failed to set debug mode")
	// ErrSetSaveOnError is returned when SetSaveOnError fails at the native layer.
	ErrSetSaveOnError = errors.New("failed to set save on error option")
	// ErrSetDrawQuality is returned when SetDrawQuality fails at the native layer.
	ErrSetDrawQuality = errors.New("failed to set draw quality")
	// ErrSetRecoImageCacheLimit is returned when SetRecoImageCacheLimit fails at the native layer.
	ErrSetRecoImageCacheLimit = errors.New("failed to set recognition image cache limit")
	// ErrLoadPlugin is returned when LoadPlugin cannot load any plugin for the given path.
	ErrLoadPlugin = errors.New("failed to load plugin")
	// ErrEmptyLogDir is returned by SetLogDir when the path is empty.
	ErrEmptyLogDir = errors.New("log directory path is empty")
)

// LibraryLoadError represents an error that occurs when loading a MAA dynamic library.
// This error type provides detailed information about which library failed to load,
// including the library name, the full path attempted, and the underlying system error.
type LibraryLoadError = native.LibraryLoadError

// SymbolLookupError reports a MaaFramework library that lacks a symbol required
// by this binding. Use errors.As to inspect the library, symbol, and version
// expectation after Init fails.
type SymbolLookupError = native.SymbolLookupError

// ErrLibraryInUse reports an attempt to release the libraries while native
// objects remain alive, the Agent Server has not been shut down, or its service
// thread has been detached.
var ErrLibraryInUse = errors.New("maa: cannot release libraries while native objects or the agent server are active")

// initConfig contains configuration options for initializing the MAA framework.
// It specifies various settings that control the framework's behavior,
// logging, debugging, and resource locations.
type initConfig struct {
	// LibDir specifies the directory path where MAA dynamic libraries are located.
	// If empty, libraries are located using the platform loader's current search configuration.
	LibDir string

	// LogDir specifies the directory where log files will be written.
	// Nil means Init will not set this option.
	LogDir *string

	// SaveDraw controls whether to save recognition results to LogDir/vision.
	// When enabled, RecoDetail will be able to retrieve draws for debugging purposes.
	// Nil means Init will not set this option.
	SaveDraw *bool

	// StdoutLevel sets the logging verbosity level for standard output.
	// Controls which log messages are displayed on the console.
	// Nil means Init will not set this option.
	StdoutLevel *LoggingLevel

	// DebugMode enables or disables comprehensive debug mode.
	// When enabled, additional debug information is collected and logged.
	// Nil means Init will not set this option.
	DebugMode *bool

	// PluginPaths specifies the paths to the plugins to load.
	// Nil means Init will not process plugin loading.
	PluginPaths *[]string

	// JSONEncoder sets a custom JSON encoder for the framework.
	// Nil means Init will not change the current encoder.
	JSONEncoder JSONEncoder

	// JSONDecoder sets a custom JSON decoder for the framework.
	// Nil means Init will not change the current decoder.
	JSONDecoder JSONDecoder
}

// InitOption defines a function type for configuring initialization through functional options.
// Use package-provided WithXxx helpers to construct options.
type InitOption func(*initConfig)

// WithLibDir returns an InitOption that sets the library directory path for the MAA framework.
// The libDir parameter specifies the directory where the MAA dynamic library is located.
// An empty directory uses the platform loader's current search configuration.
//
// On Windows, when Init applies a nonempty directory, it calls SetDllDirectoryW
// to change the DLL search directory for the whole process. Once that call
// succeeds, neither Release nor initialization failure rollback restores the
// previous configuration. It remains in effect until the process exits or
// another DLL search configuration change replaces it. Later Init calls with
// an empty directory leave the current configuration unchanged.
// For unpackaged, unprotected Win32 processes, the change can also affect the
// DLL search order of subsequently started child processes.
func WithLibDir(libDir string) InitOption {
	return func(ic *initConfig) {
		ic.LibDir = libDir
	}
}

// WithLogDir returns an InitOption that sets the directory path for log files.
// The logDir parameter specifies where the MAA framework should write its log files.
func WithLogDir(logDir string) InitOption {
	return func(ic *initConfig) {
		ic.LogDir = &logDir
	}
}

// WithSaveDraw returns an InitOption that configures whether to save drawing information.
// When enabled is true, recognition results will be saved to LogDir/vision directory
// and RecoDetail will be able to retrieve draws for debugging.
func WithSaveDraw(enabled bool) InitOption {
	return func(ic *initConfig) {
		ic.SaveDraw = &enabled
	}
}

// WithStdoutLevel returns an InitOption that sets the logging level for standard output.
// The level parameter determines the verbosity of logs written to stdout.
func WithStdoutLevel(level LoggingLevel) InitOption {
	return func(ic *initConfig) {
		ic.StdoutLevel = &level
	}
}

// WithDebugMode returns an InitOption that enables or disables debug mode.
// When enabled is true, additional debug information will be collected and logged.
func WithDebugMode(enabled bool) InitOption {
	return func(ic *initConfig) {
		ic.DebugMode = &enabled
	}
}

// WithPluginPaths returns an InitOption that sets the plugin paths loaded
// during Init. Init copies the paths when applying the option; changes to
// the argument slice before Init affect the paths used. Each path follows
// the LoadPlugin resolution rules; passing no paths yields an empty list
// that loads nothing.
func WithPluginPaths(path ...string) InitOption {
	return func(ic *initConfig) {
		pluginPaths := append([]string(nil), path...)
		ic.PluginPaths = &pluginPaths
	}
}

// WithJSONEncoder returns an InitOption that sets a custom JSON encoder.
// The encoder must satisfy the compatibility requirements documented by [JSONEncoder].
func WithJSONEncoder(encoder JSONEncoder) InitOption {
	if encoder == nil {
		panic("json encoder cannot be nil")
	}
	return func(ic *initConfig) {
		ic.JSONEncoder = encoder
	}
}

// WithJSONDecoder returns an InitOption that sets a custom JSON decoder.
func WithJSONDecoder(decoder JSONDecoder) InitOption {
	if decoder == nil {
		panic("json decoder cannot be nil")
	}
	return func(ic *initConfig) {
		ic.JSONDecoder = decoder
	}
}

// Init loads MaaFramework, MaaToolkit, MaaAgentServer, and MaaAgentClient and
// registers their functions. All four libraries must come from the same
// MaaFramework release compatible with this binding.
//
// It must be called before invoking any other MAA-related functions.
// Once Init has succeeded, later calls to Init are no-ops and any options
// passed to them are discarded.
//
// A missing library is reported by [LibraryLoadError], and a missing required
// symbol by [SymbolLookupError]. Use errors.As to inspect these errors.
// On failure, Init attempts to unload the libraries it opened. If cleanup
// fails, its error is included in the returned error, and handles for libraries
// that could not be unloaded are retained. Call [Release] to retry cleanup
// before calling Init again.
//
// Calls to Init and Release are serialized. Other MAA-related functions must
// not run concurrently with Init or Release.
//
// On Windows, DLL search configuration changed through WithLibDir is not
// restored if initialization fails; see [WithLibDir] for its process-wide scope.
//
// Note: If this function is not called before other MAA functions, it will trigger a null pointer panic.
func Init(opts ...InitOption) (err error) {
	lifecycleMu.Lock()
	defer lifecycleMu.Unlock()

	if inited {
		return nil
	}

	cfg := initConfig{}

	for _, opt := range opts {
		opt(&cfg)
	}

	if err := native.Initialize(cfg.LibDir); err != nil {
		return err
	}

	success := false
	defer func() {
		if !success {
			if shutdownErr := shutdownNativeLibraries(); shutdownErr != nil {
				err = errors.Join(err, fmt.Errorf("failed to roll back native initialization: %w", shutdownErr))
			}
		}
	}()

	if cfg.LogDir != nil {
		if err := SetLogDir(*cfg.LogDir); err != nil {
			return err
		}
	}
	if cfg.SaveDraw != nil {
		if err := SetSaveDraw(*cfg.SaveDraw); err != nil {
			return err
		}
	}
	if cfg.StdoutLevel != nil {
		if err := SetStdoutLevel(*cfg.StdoutLevel); err != nil {
			return err
		}
	}
	if cfg.DebugMode != nil {
		if err := SetDebugMode(*cfg.DebugMode); err != nil {
			return err
		}
	}

	if cfg.PluginPaths != nil {
		for _, path := range *cfg.PluginPaths {
			if err := LoadPlugin(path); err != nil {
				return err
			}
		}
	}

	if cfg.JSONEncoder != nil {
		SetJSONEncoder(cfg.JSONEncoder)
	}
	if cfg.JSONDecoder != nil {
		SetJSONDecoder(cfg.JSONDecoder)
	}

	inited = true
	success = true

	return nil
}

// IsInited checks if the MAA framework has been initialized.
// It is safe to call concurrently with Init and Release.
func IsInited() bool {
	lifecycleMu.RLock()
	defer lifecycleMu.RUnlock()
	return inited
}

// Release releases the dynamic library resources of the MAA framework and unregisters its related functions.
// It returns ErrLibraryInUse while native objects remain alive or the Agent
// Server has not been shut down. After AgentServerDetach, Release remains
// blocked for the rest of the process, even after AgentServerJoin or
// AgentServerShutDown, because native thread exit cannot be confirmed.
// Calls to Init and Release are serialized.
// Other MAA-related functions must not run concurrently with Init or Release.
// If unloading fails, IsInited becomes false; call Release again to retry
// cleanup before calling Init.
// On Windows, Release does not restore the DLL search configuration changed
// during Init. A later Init with an empty library directory uses the process's
// current configuration; see [WithLibDir].
func Release() error {
	lifecycleMu.Lock()
	defer lifecycleMu.Unlock()

	phase := agentServerPhase(agentServerState.Load())
	if liveNativeObjects.Load() != 0 || phase != agentServerStopped && phase != agentServerClosed {
		return ErrLibraryInUse
	}

	// A failed shutdown can already have cleared some native function variables.
	// Mark the package unavailable even when cleanup must be retried.
	err := shutdownNativeLibraries()
	inited = false
	return err
}

func setGlobalOption(key native.MaaGlobalOption, value unsafe.Pointer, valSize uintptr) bool {
	return native.MaaGlobalSetOption(key, value, uint64(valSize))
}

// SetLogDir sets the log directory.
// An empty path is rejected with ErrEmptyLogDir.
func SetLogDir(path string) error {
	if path == "" {
		return ErrEmptyLogDir
	}
	if !setGlobalOption(native.MaaGlobalOption_LogDir, unsafe.Pointer(&[]byte(path)[0]), uintptr(len(path))) {
		return ErrSetLogDir
	}
	return nil
}

// SetSaveDraw sets whether to save draw.
func SetSaveDraw(enabled bool) error {
	if !setGlobalOption(native.MaaGlobalOption_SaveDraw, unsafe.Pointer(&enabled), unsafe.Sizeof(enabled)) {
		return ErrSetSaveDraw
	}
	return nil
}

// LoggingLevel defines the logging verbosity levels, mirroring the
// MaaLoggingLevelEnum values of MaaDef.h (Off = 0 through All = 7).
type LoggingLevel int32

// LoggingLevel values accepted by SetStdoutLevel, ordered by increasing verbosity.
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
func SetStdoutLevel(level LoggingLevel) error {
	if !setGlobalOption(native.MaaGlobalOption_StdoutLevel, unsafe.Pointer(&level), unsafe.Sizeof(level)) {
		return ErrSetStdoutLevel
	}
	return nil
}

// SetDebugMode sets whether to enable debug mode.
func SetDebugMode(enabled bool) error {
	if !setGlobalOption(native.MaaGlobalOption_DebugMode, unsafe.Pointer(&enabled), unsafe.Sizeof(enabled)) {
		return ErrSetDebugMode
	}
	return nil
}

// SetSaveOnError sets whether to save screenshot on error.
func SetSaveOnError(enabled bool) error {
	if !setGlobalOption(native.MaaGlobalOption_SaveOnError, unsafe.Pointer(&enabled), unsafe.Sizeof(enabled)) {
		return ErrSetSaveOnError
	}
	return nil
}

// SetDrawQuality sets image quality for draw images.
// Default value is 85, range: [0, 100].
// Values outside the range fail with ErrSetDrawQuality instead of being clamped.
func SetDrawQuality(quality int32) error {
	if !setGlobalOption(native.MaaGlobalOption_DrawQuality, unsafe.Pointer(&quality), unsafe.Sizeof(quality)) {
		return ErrSetDrawQuality
	}
	return nil
}

// SetRecoImageCacheLimit sets recognition image cache limit.
// Default value is 4096.
func SetRecoImageCacheLimit(limit uint64) error {
	if !setGlobalOption(native.MaaGlobalOption_RecoImageCacheLimit, unsafe.Pointer(&limit), unsafe.Sizeof(limit)) {
		return ErrSetRecoImageCacheLimit
	}
	return nil
}

// LoadPlugin loads a plugin specified by path.
// The path may be a full filesystem path or just a plugin name.
// A bare name is resolved against the directory of the loaded MaaFramework
// library first, then through the platform's dynamic-library search.
// If the path refers to a directory, plugins inside it are searched
// recursively, and the load fails with ErrLoadPlugin unless at least one
// plugin loads. Plugins in the "plugins" directory next to the MaaFramework
// library are loaded automatically when the library itself is loaded.
func LoadPlugin(path string) error {
	if !native.MaaGlobalLoadPlugin(path) {
		return ErrLoadPlugin
	}
	return nil
}
