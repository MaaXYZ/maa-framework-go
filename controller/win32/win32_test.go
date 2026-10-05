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
		{"DXGI_DesktopDup", ScreencapDXGIDesktopDup, "DXGI_DesktopDup"},
		{"DXGI_DesktopDup_Window", ScreencapDXGIDesktopDupWindow, "DXGI_DesktopDup_Window"},
		{"PrintWindow", ScreencapPrintWindow, "PrintWindow"},
		{"ScreenDC", ScreencapScreenDC, "ScreenDC"},
		{"All", ScreencapAll, "All"},
		{"Foreground", ScreencapForeground, "Foreground"},
		{"Background", ScreencapBackground, "Background"},
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
		{"SendMessageWithWindowPos", InputSendMessageWithWindowPos, "SendMessageWithWindowPos"},
		{"PostMessageWithWindowPos", InputPostMessageWithWindowPos, "PostMessageWithWindowPos"},
		{"Interception", InputInterception, "Interception"},
		{"AnchoredTouch", InputAnchoredTouch, "AnchoredTouch"},
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
		{"DXGI_DesktopDup", "DXGI_DesktopDup", ScreencapDXGIDesktopDup, false},
		{"DXGI_DesktopDup_Window", "DXGI_DesktopDup_Window", ScreencapDXGIDesktopDupWindow, false},
		{"All", "All", ScreencapAll, false},
		{"Foreground", "Foreground", ScreencapForeground, false},
		{"Background", "Background", ScreencapBackground, false},
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
		{"SendMessageWithWindowPos", "SendMessageWithWindowPos", InputSendMessageWithWindowPos, false},
		{"PostMessageWithWindowPos", "PostMessageWithWindowPos", InputPostMessageWithWindowPos, false},
		{"Interception", "Interception", InputInterception, false},
		{"AnchoredTouch", "AnchoredTouch", InputAnchoredTouch, false},
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
		InputSendMessageWithWindowPos,
		InputPostMessageWithWindowPos,
		InputInterception,
		InputAnchoredTouch,
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

// TestEnumValuesMatchCAbi pins the exported enum values against the C ABI in
// deps/include/MaaFramework/MaaDef.h. The String/Parse round-trip tests above
// cannot catch a silent renumbering of the underlying bits.
func TestEnumValuesMatchCAbi(t *testing.T) {
	screencap := []struct {
		name   string
		method ScreencapMethod
		want   uint64
	}{
		{"ScreencapNone", ScreencapNone, 0},
		{"ScreencapGDI", ScreencapGDI, 1 << 0},
		{"ScreencapFramePool", ScreencapFramePool, 1 << 1},
		{"ScreencapDXGIDesktopDup", ScreencapDXGIDesktopDup, 1 << 2},
		{"ScreencapDXGIDesktopDupWindow", ScreencapDXGIDesktopDupWindow, 1 << 3},
		{"ScreencapPrintWindow", ScreencapPrintWindow, 1 << 4},
		{"ScreencapScreenDC", ScreencapScreenDC, 1 << 5},
		{"ScreencapAll", ScreencapAll, ^uint64(0)},
		// Foreground = DXGI_DesktopDup_Window | ScreenDC
		{"ScreencapForeground", ScreencapForeground, 1<<3 | 1<<5},
		// Background = FramePool | PrintWindow
		{"ScreencapBackground", ScreencapBackground, 1<<1 | 1<<4},
	}
	for _, tt := range screencap {
		if got := uint64(tt.method); got != tt.want {
			t.Errorf("%s = %#x, want %#x", tt.name, got, tt.want)
		}
	}

	input := []struct {
		name   string
		method InputMethod
		want   uint64
	}{
		{"InputNone", InputNone, 0},
		{"InputSeize", InputSeize, 1 << 0},
		{"InputSendMessage", InputSendMessage, 1 << 1},
		{"InputPostMessage", InputPostMessage, 1 << 2},
		{"InputLegacyEvent", InputLegacyEvent, 1 << 3},
		{"InputPostThreadMessage", InputPostThreadMessage, 1 << 4},
		{"InputSendMessageWithCursorPos", InputSendMessageWithCursorPos, 1 << 5},
		{"InputPostMessageWithCursorPos", InputPostMessageWithCursorPos, 1 << 6},
		{"InputSendMessageWithWindowPos", InputSendMessageWithWindowPos, 1 << 7},
		{"InputPostMessageWithWindowPos", InputPostMessageWithWindowPos, 1 << 8},
		{"InputInterception", InputInterception, 1 << 9},
		{"InputAnchoredTouch", InputAnchoredTouch, 1 << 10},
	}
	for _, tt := range input {
		if got := uint64(tt.method); got != tt.want {
			t.Errorf("%s = %#x, want %#x", tt.name, got, tt.want)
		}
	}
}

// TestScreencapMethod_CanonicalSpellings pins the DXGI spellings against the
// upstream bindings (PiCli Configurator.cpp, Python define.py, NodeJS
// constant.cpp): String emits the underscore spellings and Parse accepts both
// them and this binding's legacy underscore-free spellings.
func TestScreencapMethod_CanonicalSpellings(t *testing.T) {
	if got := ScreencapDXGIDesktopDup.String(); got != "DXGI_DesktopDup" {
		t.Errorf("ScreencapDXGIDesktopDup.String() = %q, want %q", got, "DXGI_DesktopDup")
	}
	if got := ScreencapDXGIDesktopDupWindow.String(); got != "DXGI_DesktopDup_Window" {
		t.Errorf("ScreencapDXGIDesktopDupWindow.String() = %q, want %q", got, "DXGI_DesktopDup_Window")
	}

	for _, tt := range []struct {
		input    string
		expected ScreencapMethod
	}{
		{"DXGI_DesktopDup", ScreencapDXGIDesktopDup},
		{"DXGI_DesktopDup_Window", ScreencapDXGIDesktopDupWindow},
		{"DXGIDesktopDup", ScreencapDXGIDesktopDup},
		{"DXGIDesktopDupWindow", ScreencapDXGIDesktopDupWindow},
	} {
		got, err := ParseScreencapMethod(tt.input)
		if err != nil {
			t.Errorf("ParseScreencapMethod(%q) error = %v", tt.input, err)
			continue
		}
		if got != tt.expected {
			t.Errorf("ParseScreencapMethod(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
