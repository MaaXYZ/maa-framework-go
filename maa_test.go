package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetLogDir(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		expectedErr error
	}{
		{
			name:        "ValidPath",
			path:        "./test/debug",
			expectedErr: nil,
		},
		{
			name:        "EmptyPath",
			path:        "",
			expectedErr: ErrEmptyLogDir,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SetLogDir(tc.path)
			require.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestSetSaveDraw(t *testing.T) {
	testCases := []struct {
		name        string
		enabled     bool
		expectedErr error
	}{
		{
			name:        "EnableSaveDraw",
			enabled:     true,
			expectedErr: nil,
		},
		{
			name:        "DisableSaveDraw",
			enabled:     false,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SetSaveDraw(tc.enabled)
			require.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestSetStdoutLevel(t *testing.T) {
	testCases := []struct {
		name        string
		level       LoggingLevel
		expectedErr error
	}{
		{
			name:        "SetLevelOff",
			level:       LoggingLevelOff,
			expectedErr: nil,
		},
		{
			name:        "SetLevelFatal",
			level:       LoggingLevelFatal,
			expectedErr: nil,
		},
		{
			name:        "SetLevelError",
			level:       LoggingLevelError,
			expectedErr: nil,
		},
		{
			name:        "SetLevelWarn",
			level:       LoggingLevelWarn,
			expectedErr: nil,
		},
		{
			name:        "SetLevelInfo",
			level:       LoggingLevelInfo,
			expectedErr: nil,
		},
		{
			name:        "SetLevelDebug",
			level:       LoggingLevelDebug,
			expectedErr: nil,
		},
		{
			name:        "SetLevelTrace",
			level:       LoggingLevelTrace,
			expectedErr: nil,
		},
		{
			name:        "SetLevelAll",
			level:       LoggingLevelAll,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SetStdoutLevel(tc.level)
			require.Equal(t, tc.expectedErr, err)
		})
	}

	SetStdoutLevel(LoggingLevelOff)
}

func TestSetDebugMode(t *testing.T) {
	testCases := []struct {
		name        string
		enabled     bool
		expectedErr error
	}{
		{
			name:        "EnableDebugMode",
			enabled:     true,
			expectedErr: nil,
		},
		{
			name:        "DisableDebugMode",
			enabled:     false,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SetDebugMode(tc.enabled)
			require.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestSetSaveOnError(t *testing.T) {
	testCases := []struct {
		name        string
		enabled     bool
		expectedErr error
	}{
		{
			name:        "EnableSaveOnError",
			enabled:     true,
			expectedErr: nil,
		},
		{
			name:        "DisableSaveOnError",
			enabled:     false,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SetSaveOnError(tc.enabled)
			require.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestSetDrawQuality(t *testing.T) {
	testCases := []struct {
		name        string
		quality     int32
		expectedErr error
	}{
		{
			name:        "BelowRange",
			quality:     -1,
			expectedErr: ErrSetDrawQuality,
		},
		{
			name:        "MinBound",
			quality:     0,
			expectedErr: nil,
		},
		{
			name:        "Default",
			quality:     85,
			expectedErr: nil,
		},
		{
			name:        "MaxBound",
			quality:     100,
			expectedErr: nil,
		},
		{
			name:        "AboveRange",
			quality:     101,
			expectedErr: ErrSetDrawQuality,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SetDrawQuality(tc.quality)
			require.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestSetRecoImageCacheLimit(t *testing.T) {
	testCases := []struct {
		name        string
		limit       uint64
		expectedErr error
	}{
		{
			name:        "Zero",
			limit:       0,
			expectedErr: nil,
		},
		{
			name:        "Default",
			limit:       4096,
			expectedErr: nil,
		},
		{
			name:        "Large",
			limit:       1 << 20,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SetRecoImageCacheLimit(tc.limit)
			require.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestWithPluginPaths(t *testing.T) {
	paths := []string{"./test/plugin_a", "./test/plugin_b"}
	opt := WithPluginPaths(paths...)

	cfg := initConfig{}
	opt(&cfg)
	require.NotNil(t, cfg.PluginPaths)
	require.Equal(t, []string{"./test/plugin_a", "./test/plugin_b"}, *cfg.PluginPaths)

	paths[0] = "./test/mutated"
	require.Equal(t, "./test/plugin_a", (*cfg.PluginPaths)[0],
		"the option must copy the argument slice")

	WithPluginPaths()(&cfg)
	require.NotNil(t, cfg.PluginPaths)
	require.Empty(t, *cfg.PluginPaths)
}

func TestLoadPlugin(t *testing.T) {
	err := LoadPlugin("./test/maa-nonexistent-plugin")
	require.Equal(t, ErrLoadPlugin, err)
}
