package maa

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSetLogDir(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "ValidPath",
			path:     "./test/debug",
			expected: true,
		},
		{
			name:     "EmptyPath",
			path:     "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SetLogDir(tc.path)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestSetSaveDraw(t *testing.T) {
	testCases := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "EnableSaveDraw",
			enabled:  true,
			expected: true,
		},
		{
			name:     "DisableSaveDraw",
			enabled:  false,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SetSaveDraw(tc.enabled)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestSetRecording(t *testing.T) {
	testCases := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "EnableRecording",
			enabled:  true,
			expected: true,
		},
		{
			name:     "DisableRecording",
			enabled:  false,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SetRecording(tc.enabled)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestSetStdoutLevel(t *testing.T) {
	testCases := []struct {
		name     string
		level    LoggingLevel
		expected bool
	}{
		{
			name:     "SetLevelOff",
			level:    LoggingLevelOff,
			expected: true,
		},
		{
			name:     "SetLevelFatal",
			level:    LoggingLevelFatal,
			expected: true,
		},
		{
			name:     "SetLevelError",
			level:    LoggingLevelError,
			expected: true,
		},
		{
			name:     "SetLevelWarn",
			level:    LoggingLevelWarn,
			expected: true,
		},
		{
			name:     "SetLevelInfo",
			level:    LoggingLevelInfo,
			expected: true,
		},
		{
			name:     "SetLevelDebug",
			level:    LoggingLevelDebug,
			expected: true,
		},
		{
			name:     "SetLevelTrace",
			level:    LoggingLevelTrace,
			expected: true,
		},
		{
			name:     "SetLevelAll",
			level:    LoggingLevelAll,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SetStdoutLevel(tc.level)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestSetDebugMode(t *testing.T) {
	testCases := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "EnableDebugMode",
			enabled:  true,
			expected: true,
		},
		{
			name:     "DisableDebugMode",
			enabled:  false,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SetDebugMode(tc.enabled)
			require.Equal(t, tc.expected, result)
		})
	}
}
