package macos

import "testing"

// TestEnumValuesMatchCAbi pins the exported enum values against the C ABI in
// deps/include/MaaFramework/MaaDef.h.
func TestEnumValuesMatchCAbi(t *testing.T) {
	screencap := []struct {
		name   string
		method ScreencapMethod
		want   uint64
	}{
		{"ScreencapNone", ScreencapNone, 0},
		{"ScreencapScreenCaptureKit", ScreencapScreenCaptureKit, 1 << 0},
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
		{"InputGlobalEvent", InputGlobalEvent, 1 << 0},
		{"InputPostToPid", InputPostToPid, 1 << 1},
	}
	for _, tt := range input {
		if got := uint64(tt.method); got != tt.want {
			t.Errorf("%s = %#x, want %#x", tt.name, got, tt.want)
		}
	}
}
