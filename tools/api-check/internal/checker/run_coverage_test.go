package checker

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRunNativeExclusions(t *testing.T) {
	for _, tt := range []struct {
		name, config string
		code         int
		want         string
	}{
		{"with reason", "native_exclusions:\n  MaaOmitted: Intentional omission\n", 0, "native_exclusion: MaaOmitted: Intentional omission"},
		{"stale", "native_exclusions:\n  MaaAbsent: Old omission\n", 1, "MaaAbsent: stale native exclusion"},
		{"blank reason", "native_exclusions:\n  MaaOmitted: ''\n", 2, "nonempty reason"},
		{"trimmed symbol", "native_exclusions:\n  ' MaaOmitted ': Reason\n", 2, "exact nonempty symbol"},
		{"numeric reason", "native_exclusions:\n  MaaOmitted: 123\n", 2, "must be a string"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := writeRepoFixtureWith(t, map[string]string{"deps/include/omission.h": "MAA_FRAMEWORK_API void MaaOmitted(void);\n"})
			path := writeFixtureFile(t, filepath.Join(dir, "exclusions.yaml"), tt.config)
			code, stdout, stderr := runChecker(t, dir, "--config", path)
			if code != tt.code || !strings.Contains(stdout+stderr, tt.want) {
				t.Fatalf("exit %d, want %d; output:\n%s%s", code, tt.code, stdout, stderr)
			}
		})
	}
}

func TestRunRejectsEmptyNativeInventory(t *testing.T) {
	dir := writeRepoFixtureWith(t, map[string]string{"deps/include/agent_client.h": ""})
	code, stdout, stderr := runChecker(t, dir)
	if code != 2 || !strings.Contains(stderr, "empty native inventory for agent_client") || strings.Contains(stdout, "PASS") {
		t.Fatalf("exit %d; output:\n%s%s", code, stdout, stderr)
	}
}
