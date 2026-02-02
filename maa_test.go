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
