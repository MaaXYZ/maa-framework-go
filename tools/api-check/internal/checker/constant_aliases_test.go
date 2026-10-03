package checker

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestPublicConstantAliases(t *testing.T) {
	for _, tt := range []struct{ name, alias, want string }{
		{"matching", "native.MaaGamepadType_Xbox360", ""},
		{"wrong native member", "native.MaaGamepadType_DualShock4", "constant value mismatch: Xbox360"},
		{"unknown native member", "native.MaaGamepadType_Missing", "failed to evaluate Go constant: Xbox360"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			writeFixtureFile(t, filepath.Join(root, "internal/native/framework.go"), "package native\ntype MaaGamepadType uint64\nconst MaaGamepadType_Xbox360 MaaGamepadType = 0\nconst MaaGamepadType_DualShock4 MaaGamepadType = 1\n")
			writeFixtureFile(t, filepath.Join(root, "internal/native/toolkit.go"), "package native\n")
			writeFixtureFile(t, filepath.Join(root, "controller.go"), "package maa\ntype GamepadType = native.MaaGamepadType\nconst GamepadTypeXbox360 GamepadType = "+tt.alias+"\nconst GamepadTypeDualShock4 GamepadType = native.MaaGamepadType_DualShock4\n")
			evaluation, err := evaluatePublicConstants(root, "controller.go")
			if err != nil {
				t.Fatal(err)
			}
			env, err := evaluateCConstSources([]cConstSource{{content: "#define MaaGamepadType_Xbox360 0ULL\n#define MaaGamepadType_DualShock4 1ULL\n"}})
			if err != nil {
				t.Fatal(err)
			}
			spec := constantFamilySpecs[0]
			issues := compareConstantFamily(spec, env, evaluation)
			if tt.want == "" && len(issues) != 0 || tt.want != "" && !strings.Contains(issueText(issues), tt.want) {
				t.Fatalf("issues = %v, want %q", issues, tt.want)
			}
		})
	}
}
