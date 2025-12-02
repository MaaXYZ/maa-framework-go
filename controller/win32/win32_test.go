package win32

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
		{"GDI", ScreencapGDI, "GDI"},
		{"FramePool", ScreencapFramePool, "FramePool"},
		{"DXGIDesktopDup", ScreencapDXGIDesktopDup, "DXGIDesktopDup"},
		{"DXGIDesktopDupWindow", ScreencapDXGIDesktopDupWindow, "DXGIDesktopDupWindow"},
		{"PrintWindow", ScreencapPrintWindow, "PrintWindow"},
		{"ScreenDC", ScreencapScreenDC, "ScreenDC"},
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
		{"Seize", InputSeize, "Seize"},
		{"SendMessage", InputSendMessage, "SendMessage"},
		{"PostMessage", InputPostMessage, "PostMessage"},
		{"LegacyEvent", InputLegacyEvent, "LegacyEvent"},
		{"PostThreadMessage", InputPostThreadMessage, "PostThreadMessage"},
		{"SendMessageWithCursorPos", InputSendMessageWithCursorPos, "SendMessageWithCursorPos"},
		{"PostMessageWithCursorPos", InputPostMessageWithCursorPos, "PostMessageWithCursorPos"},
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
		{"GDI", "GDI", ScreencapGDI, false},
		{"FramePool", "FramePool", ScreencapFramePool, false},
		{"DXGIDesktopDup", "DXGIDesktopDup", ScreencapDXGIDesktopDup, false},
		{"DXGIDesktopDupWindow", "DXGIDesktopDupWindow", ScreencapDXGIDesktopDupWindow, false},
		{"PrintWindow", "PrintWindow", ScreencapPrintWindow, false},
		{"ScreenDC", "ScreenDC", ScreencapScreenDC, false},
		// Case insensitive
		{"LowerCase", "gdi", ScreencapGDI, false},
		{"UpperCase", "GDI", ScreencapGDI, false},
		{"MixedCase", "GdI", ScreencapGDI, false},
		// With whitespace
		{"WithSpaces", "  GDI  ", ScreencapGDI, false},
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
		{"Seize", "Seize", InputSeize, false},
		{"SendMessage", "SendMessage", InputSendMessage, false},
		{"PostMessage", "PostMessage", InputPostMessage, false},
		{"LegacyEvent", "LegacyEvent", InputLegacyEvent, false},
		{"PostThreadMessage", "PostThreadMessage", InputPostThreadMessage, false},
		{"SendMessageWithCursorPos", "SendMessageWithCursorPos", InputSendMessageWithCursorPos, false},
		{"PostMessageWithCursorPos", "PostMessageWithCursorPos", InputPostMessageWithCursorPos, false},
		// Case insensitive
		{"LowerCase", "seize", InputSeize, false},
		{"UpperCase", "SEIZE", InputSeize, false},
		{"MixedCase", "sEiZe", InputSeize, false},
		// With whitespace
		{"WithSpaces", "  Seize  ", InputSeize, false},
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
		ScreencapGDI,
		ScreencapFramePool,
		ScreencapDXGIDesktopDup,
		ScreencapDXGIDesktopDupWindow,
		ScreencapPrintWindow,
		ScreencapScreenDC,
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
		InputSeize,
		InputSendMessage,
		InputPostMessage,
		InputLegacyEvent,
		InputPostThreadMessage,
		InputSendMessageWithCursorPos,
		InputPostMessageWithCursorPos,
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
