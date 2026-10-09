package adb

import (
	"testing"
)

func TestScreencapMethod_String(t *testing.T) {
	tests := []struct {
		name     string
		method   ScreencapMethod
		expected string
	}{
		{"None", ScreencapNone, ""},
		{"EncodeToFileAndPull", ScreencapEncodeToFileAndPull, "EncodeToFileAndPull"},
		{"Encode", ScreencapEncode, "Encode"},
		{"RawWithGzip", ScreencapRawWithGzip, "RawWithGzip"},
		{"RawByNetcat", ScreencapRawByNetcat, "RawByNetcat"},
		{"MinicapDirect", ScreencapMinicapDirect, "MinicapDirect"},
		{"MinicapStream", ScreencapMinicapStream, "MinicapStream"},
		{"EmulatorExtras", ScreencapEmulatorExtras, "EmulatorExtras"},
		{"All", ScreencapAll, "All"},
		{"Default", ScreencapDefault, "Default"},
		{"Unknown", ScreencapMethod(999), "999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.method.String(); got != tt.expected {
				t.Errorf("ScreencapMethod.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestInputMethod_String(t *testing.T) {
	tests := []struct {
		name     string
		method   InputMethod
		expected string
	}{
		{"None", InputNone, ""},
		{"AdbShell", InputAdbShell, "AdbShell"},
		{"MinitouchAndAdbKey", InputMinitouchAndAdbKey, "MinitouchAndAdbKey"},
		{"Maatouch", InputMaatouch, "Maatouch"},
		{"EmulatorExtras", InputEmulatorExtras, "EmulatorExtras"},
		{"All", InputAll, "All"},
		{"Default", InputDefault, "Default"},
		{"Unknown", InputMethod(999), "999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.method.String(); got != tt.expected {
				t.Errorf("InputMethod.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseScreencapMethod(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  ScreencapMethod
		expectErr bool
	}{
		{"Empty", "", ScreencapNone, false},
		{"EncodeToFileAndPull", "EncodeToFileAndPull", ScreencapEncodeToFileAndPull, false},
		{"Encode", "Encode", ScreencapEncode, false},
		{"RawWithGzip", "RawWithGzip", ScreencapRawWithGzip, false},
		{"RawByNetcat", "RawByNetcat", ScreencapRawByNetcat, false},
		{"MinicapDirect", "MinicapDirect", ScreencapMinicapDirect, false},
		{"MinicapStream", "MinicapStream", ScreencapMinicapStream, false},
		{"EmulatorExtras", "EmulatorExtras", ScreencapEmulatorExtras, false},
		{"All", "All", ScreencapAll, false},
		{"Default", "Default", ScreencapDefault, false},
		// Case insensitive
		{"LowerCase", "encode", ScreencapEncode, false},
		{"UpperCase", "ENCODE", ScreencapEncode, false},
		{"MixedCase", "eNcOdE", ScreencapEncode, false},
		// With whitespace
		{"WithSpaces", "  Encode  ", ScreencapEncode, false},
		// Numeric string
		{"NumericString", "4", ScreencapMethod(4), false},
		// Invalid
		{"Invalid", "invalid_method", ScreencapNone, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseScreencapMethod(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("ParseScreencapMethod() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ParseScreencapMethod() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseInputMethod(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  InputMethod
		expectErr bool
	}{
		{"Empty", "", InputNone, false},
		{"AdbShell", "AdbShell", InputAdbShell, false},
		{"MinitouchAndAdbKey", "MinitouchAndAdbKey", InputMinitouchAndAdbKey, false},
		{"Maatouch", "Maatouch", InputMaatouch, false},
		{"EmulatorExtras", "EmulatorExtras", InputEmulatorExtras, false},
		{"All", "All", InputAll, false},
		{"Default", "Default", InputDefault, false},
		// Case insensitive
		{"LowerCase", "adbshell", InputAdbShell, false},
		{"UpperCase", "ADBSHELL", InputAdbShell, false},
		{"MixedCase", "aDbShElL", InputAdbShell, false},
		// With whitespace
		{"WithSpaces", "  AdbShell  ", InputAdbShell, false},
		// Numeric string
		{"NumericString", "2", InputMethod(2), false},
		// Invalid
		{"Invalid", "invalid_method", InputNone, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInputMethod(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("ParseInputMethod() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ParseInputMethod() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestScreencapMethodRoundTrip(t *testing.T) {
	methods := []ScreencapMethod{
		ScreencapEncodeToFileAndPull,
		ScreencapEncode,
		ScreencapRawWithGzip,
		ScreencapRawByNetcat,
		ScreencapMinicapDirect,
		ScreencapMinicapStream,
		ScreencapEmulatorExtras,
		ScreencapAll,
		ScreencapDefault,
	}

	for _, m := range methods {
		t.Run(m.String(), func(t *testing.T) {
			str := m.String()
			parsed, err := ParseScreencapMethod(str)
			if err != nil {
				t.Errorf("ParseScreencapMethod(%q) error = %v", str, err)
				return
			}
			if parsed != m {
				t.Errorf("Round trip failed: %v -> %q -> %v", m, str, parsed)
			}
		})
	}
}

func TestInputMethodRoundTrip(t *testing.T) {
	methods := []InputMethod{
		InputAdbShell,
		InputMinitouchAndAdbKey,
		InputMaatouch,
		InputEmulatorExtras,
		InputAll,
		InputDefault,
	}

	for _, m := range methods {
		t.Run(m.String(), func(t *testing.T) {
			str := m.String()
			parsed, err := ParseInputMethod(str)
			if err != nil {
				t.Errorf("ParseInputMethod(%q) error = %v", str, err)
				return
			}
			if parsed != m {
				t.Errorf("Round trip failed: %v -> %q -> %v", m, str, parsed)
			}
		})
	}
}
