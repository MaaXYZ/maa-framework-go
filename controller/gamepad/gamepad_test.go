package gamepad

import "testing"

// TestButtonValuesMatchCAbi pins the button constants against the C ABI in
// deps/include/MaaFramework/MaaDef.h, including the DualShock 4 aliases and
// the two DS4-only buttons.
func TestButtonValuesMatchCAbi(t *testing.T) {
	buttons := []struct {
		name   string
		button Button
		want   int32
	}{
		{"ButtonA", ButtonA, 0x1000},
		{"ButtonB", ButtonB, 0x2000},
		{"ButtonX", ButtonX, 0x4000},
		{"ButtonY", ButtonY, 0x8000},
		{"ButtonLB", ButtonLB, 0x0100},
		{"ButtonRB", ButtonRB, 0x0200},
		{"ButtonLeftThumb", ButtonLeftThumb, 0x0040},
		{"ButtonRightThumb", ButtonRightThumb, 0x0080},
		{"ButtonStart", ButtonStart, 0x0010},
		{"ButtonBack", ButtonBack, 0x0020},
		{"ButtonGuide", ButtonGuide, 0x0400},
		{"ButtonDpadUp", ButtonDpadUp, 0x0001},
		{"ButtonDpadDown", ButtonDpadDown, 0x0002},
		{"ButtonDpadLeft", ButtonDpadLeft, 0x0004},
		{"ButtonDpadRight", ButtonDpadRight, 0x0008},
		{"ButtonPS", ButtonPS, 0x10000},
		{"ButtonTouchpad", ButtonTouchpad, 0x20000},
	}
	for _, tt := range buttons {
		if got := int32(tt.button); got != tt.want {
			t.Errorf("%s = %#x, want %#x", tt.name, got, tt.want)
		}
	}

	aliases := []struct {
		name   string
		alias  Button
		target Button
	}{
		{"ButtonCross", ButtonCross, ButtonA},
		{"ButtonCircle", ButtonCircle, ButtonB},
		{"ButtonSquare", ButtonSquare, ButtonX},
		{"ButtonTriangle", ButtonTriangle, ButtonY},
		{"ButtonL1", ButtonL1, ButtonLB},
		{"ButtonR1", ButtonR1, ButtonRB},
		{"ButtonL3", ButtonL3, ButtonLeftThumb},
		{"ButtonR3", ButtonR3, ButtonRightThumb},
		{"ButtonOptions", ButtonOptions, ButtonStart},
		{"ButtonShare", ButtonShare, ButtonBack},
	}
	for _, tt := range aliases {
		if tt.alias != tt.target {
			t.Errorf("%s = %#x, want the Xbox equivalent %#x", tt.name, int32(tt.alias), int32(tt.target))
		}
	}
}

// TestTouchValuesMatchCAbi pins the touch contact constants against the C
// ABI in deps/include/MaaFramework/MaaDef.h.
func TestTouchValuesMatchCAbi(t *testing.T) {
	touches := []struct {
		name  string
		touch Touch
		want  int32
	}{
		{"TouchLeftStick", TouchLeftStick, 0},
		{"TouchRightStick", TouchRightStick, 1},
		{"TouchLeftTrigger", TouchLeftTrigger, 2},
		{"TouchRightTrigger", TouchRightTrigger, 3},
	}
	for _, tt := range touches {
		if got := int32(tt.touch); got != tt.want {
			t.Errorf("%s = %d, want %d", tt.name, got, tt.want)
		}
	}
}
