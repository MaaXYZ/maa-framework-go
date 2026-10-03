package checker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFixtureFile(t *testing.T, path string, body string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadConfigFromPathValid(t *testing.T) {
	path := writeFixtureFile(t, filepath.Join(t.TempDir(), "config.yaml"), strings.Join([]string{
		"header_dir: deps/include",
		"blacklist:",
		"  - MaaFoo",
		"  - MaaBar",
		"pipeline_schema: tools/pipeline.schema.json",
		"pipeline_exclusions:",
		"  action.Click.param.pressure: Intentional test omission",
		"",
	}, "\n"))

	cfg, err := loadConfigFromPath(path)
	if err != nil {
		t.Fatalf("loadConfigFromPath() error = %v", err)
	}
	if cfg.HeaderDir != "deps/include" {
		t.Errorf("HeaderDir = %q, want %q", cfg.HeaderDir, "deps/include")
	}
	if len(cfg.Blacklist) != 2 || cfg.Blacklist[0] != "MaaFoo" || cfg.Blacklist[1] != "MaaBar" {
		t.Errorf("Blacklist = %v, want [MaaFoo MaaBar]", cfg.Blacklist)
	}
	if cfg.PipelineSchema != "tools/pipeline.schema.json" {
		t.Errorf("PipelineSchema = %q", cfg.PipelineSchema)
	}
	if got := cfg.PipelineExclusions["action.Click.param.pressure"]; got != "Intentional test omission" {
		t.Errorf("PipelineExclusions = %v", cfg.PipelineExclusions)
	}
}

func TestLoadConfigFromPathEmptyAllowed(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty file", ""},
		{"blank lines", "\n\n   \n"},
		{"comment only", "# no configuration here\n"},
		{"empty document", "---\n"},
		{"null document", "null\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeFixtureFile(t, filepath.Join(t.TempDir(), "config.yaml"), tt.body)
			cfg, err := loadConfigFromPath(path)
			if err != nil {
				t.Fatalf("loadConfigFromPath() error = %v", err)
			}
			if cfg.HeaderDir != "" || len(cfg.Blacklist) != 0 || cfg.PipelineSchema != "" || len(cfg.PipelineExclusions) != 0 {
				t.Fatalf("expected zero Config, got %+v", cfg)
			}
		})
	}
}

func TestLoadConfigFromPathRejectsInvalidYAML(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"unknown key", "unknown_key: value\n", "field unknown_key not found"},
		{"unknown nested key", "pipeline_exclusions:\n  a: b\nunknown: 1\n", "field unknown not found"},
		{"mistyped scalar as list", "blacklist: not-a-list\n", "cannot unmarshal"},
		{"mistyped list as scalar", "header_dir:\n  - a\n", "cannot unmarshal"},
		{"mistyped exclusions", "pipeline_exclusions:\n  - a\n", "cannot unmarshal"},
		{"mistyped schema", "pipeline_schema: [a]\n", "cannot unmarshal"},
		{"numeric header", "header_dir: 1\n", "header_dir at line 1 must be a string"},
		{"boolean schema", "pipeline_schema: false\n", "pipeline_schema at line 1 must be a string"},
		{"numeric blacklist entry", "blacklist: [MaaFoo, 1]\n", "blacklist[1] at line 1 must be a string"},
		{"null blacklist entry", "blacklist: [null]\n", "blacklist[0] at line 1 must be a string"},
		{"numeric exclusion key", "pipeline_exclusions: {1: reason}\n", "pipeline_exclusions key at line 1 must be a string"},
		{"boolean exclusion reason", "pipeline_exclusions: {action.Click.param.pressure: false}\n", "pipeline_exclusions.action.Click.param.pressure at line 1 must be a string"},
		{"aliased numeric scalar", "header_dir: &wrong 1\npipeline_schema: *wrong\n", "header_dir at line 1 must be a string"},
		{"merged numeric scalar", "<<: {header_dir: 1}\n", "header_dir at line 1 must be a string"},
		{"sequence merged numeric scalar", "<<: [{header_dir: 1}, {pipeline_schema: correct.json}]\n", "header_dir at line 1 must be a string"},
		{"duplicate key", "header_dir: a\nheader_dir: b\n", "already defined"},
		{"duplicate blacklist key", "blacklist:\n  - a\nblacklist:\n  - b\n", "already defined"},
		{"duplicate exclusion key", "pipeline_exclusions:\n  a: x\n  a: y\n", "already defined"},
		{"extra document", "header_dir: a\n---\nheader_dir: b\n", "multiple YAML documents"},
		{"extra empty document", "header_dir: a\n---\n", "multiple YAML documents"},
		{"malformed yaml", "header_dir: [\n", "parse yaml"},
		{"sequence document", "- a\n- b\n", "cannot unmarshal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeFixtureFile(t, filepath.Join(t.TempDir(), "config.yaml"), tt.body)
			_, err := loadConfigFromPath(path)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("loadConfigFromPath() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestLoadConfigFromPathStringAliases(t *testing.T) {
	path := writeFixtureFile(t, filepath.Join(t.TempDir(), "config.yaml"), "header_dir: &path \"1\"\npipeline_schema: *path\nblacklist: [\"false\"]\npipeline_exclusions: {\"1\": \"false\"}\n")
	cfg, err := loadConfigFromPath(path)
	if err != nil {
		t.Fatalf("loadConfigFromPath() error = %v", err)
	}
	if cfg.HeaderDir != "1" || cfg.PipelineSchema != "1" || cfg.Blacklist[0] != "false" || cfg.PipelineExclusions["1"] != "false" {
		t.Fatalf("string aliases and quoted scalar values were not preserved: %+v", cfg)
	}
}

func TestLoadConfigFromPathMissing(t *testing.T) {
	_, err := loadConfigFromPath(filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil || !strings.Contains(err.Error(), "read") {
		t.Fatalf("loadConfigFromPath() error = %v, want read error", err)
	}
}

func TestResolveConfigPrecedence(t *testing.T) {
	t.Run("cwd config wins over root fallback", func(t *testing.T) {
		root := t.TempDir()
		writeFixtureFile(t, filepath.Join(root, apiCheckConfigPathRel), "header_dir: from-root\n")
		cwd := t.TempDir()
		writeFixtureFile(t, filepath.Join(cwd, autoConfigFileName), "header_dir: from-cwd\n")
		t.Chdir(cwd)

		cfg, path, err := resolveConfig("", root)
		if err != nil {
			t.Fatalf("resolveConfig() error = %v", err)
		}
		if cfg.HeaderDir != "from-cwd" || path != autoConfigFileName {
			t.Fatalf("got HeaderDir=%q path=%q, want cwd config", cfg.HeaderDir, path)
		}
	})

	t.Run("root fallback used without cwd config", func(t *testing.T) {
		root := t.TempDir()
		fallback := writeFixtureFile(t, filepath.Join(root, apiCheckConfigPathRel), "header_dir: from-root\n")
		cwd := t.TempDir()
		t.Chdir(cwd)

		cfg, path, err := resolveConfig("", root)
		if err != nil {
			t.Fatalf("resolveConfig() error = %v", err)
		}
		if cfg.HeaderDir != "from-root" || path != resolveFromRepoRoot(root, apiCheckConfigPathRel) {
			t.Fatalf("got HeaderDir=%q path=%q, want fallback %q", cfg.HeaderDir, path, fallback)
		}
	})

	t.Run("explicit config wins over both", func(t *testing.T) {
		root := t.TempDir()
		writeFixtureFile(t, filepath.Join(root, apiCheckConfigPathRel), "header_dir: from-root\n")
		cwd := t.TempDir()
		writeFixtureFile(t, filepath.Join(cwd, autoConfigFileName), "header_dir: from-cwd\n")
		explicit := writeFixtureFile(t, filepath.Join(cwd, "explicit.yaml"), "header_dir: from-explicit\n")
		t.Chdir(cwd)

		cfg, path, err := resolveConfig(explicit, root)
		if err != nil {
			t.Fatalf("resolveConfig() error = %v", err)
		}
		if cfg.HeaderDir != "from-explicit" || path != explicit {
			t.Fatalf("got HeaderDir=%q path=%q, want explicit config", cfg.HeaderDir, path)
		}
	})

	t.Run("no config yields defaults", func(t *testing.T) {
		cwd := t.TempDir()
		t.Chdir(cwd)

		cfg, path, err := resolveConfig("", t.TempDir())
		if err != nil {
			t.Fatalf("resolveConfig() error = %v", err)
		}
		if path != "" || cfg.HeaderDir != "" || len(cfg.Blacklist) != 0 {
			t.Fatalf("got path=%q cfg=%+v, want defaults", path, cfg)
		}
	})

	t.Run("invalid cwd config reported", func(t *testing.T) {
		cwd := t.TempDir()
		writeFixtureFile(t, filepath.Join(cwd, autoConfigFileName), "unknown_key: 1\n")
		t.Chdir(cwd)

		if _, _, err := resolveConfig("", t.TempDir()); err == nil || !strings.Contains(err.Error(), "field unknown_key not found") {
			t.Fatalf("resolveConfig() error = %v, want unknown field error", err)
		}
	})

	t.Run("invalid root fallback reported", func(t *testing.T) {
		root := t.TempDir()
		writeFixtureFile(t, filepath.Join(root, apiCheckConfigPathRel), "header_dir: a\nheader_dir: b\n")
		cwd := t.TempDir()
		t.Chdir(cwd)

		if _, _, err := resolveConfig("", root); err == nil || !strings.Contains(err.Error(), "already defined") {
			t.Fatalf("resolveConfig() error = %v, want duplicate key error", err)
		}
	})

	t.Run("config path that is a directory reported", func(t *testing.T) {
		cwd := t.TempDir()
		if err := os.Mkdir(filepath.Join(cwd, autoConfigFileName), 0700); err != nil {
			t.Fatal(err)
		}
		t.Chdir(cwd)

		if _, _, err := resolveConfig("", t.TempDir()); err == nil {
			t.Fatal("resolveConfig() error = nil, want read error for directory config")
		}
	})

	t.Run("missing explicit config reported", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "missing.yaml")
		_, path, err := resolveConfig(missing, t.TempDir())
		if err == nil || !strings.Contains(err.Error(), "read") {
			t.Fatalf("resolveConfig() error = %v, want read error", err)
		}
		if path != "" {
			t.Fatalf("path = %q, want empty on error", path)
		}
	})
}

func TestMergeConfig(t *testing.T) {
	dst := Config{
		HeaderDir:          "old/headers",
		Blacklist:          []string{"seed"},
		PipelineSchema:     "old.json",
		PipelineExclusions: map[string]string{"old": "reason"},
	}
	mergeConfig(&dst, Config{
		HeaderDir:          "new/headers",
		Blacklist:          []string{"extra"},
		PipelineSchema:     "new.json",
		PipelineExclusions: map[string]string{"new": "reason"},
	})
	if dst.HeaderDir != "new/headers" || dst.PipelineSchema != "new.json" {
		t.Fatalf("scalar overrides not applied: %+v", dst)
	}
	if len(dst.Blacklist) != 2 || dst.Blacklist[0] != "seed" || dst.Blacklist[1] != "extra" {
		t.Fatalf("blacklist not appended: %v", dst.Blacklist)
	}
	if len(dst.PipelineExclusions) != 1 || dst.PipelineExclusions["new"] != "reason" {
		t.Fatalf("exclusions not replaced: %v", dst.PipelineExclusions)
	}

	mergeConfig(&dst, Config{HeaderDir: "  ", PipelineSchema: "\t"})
	if dst.HeaderDir != "new/headers" || dst.PipelineSchema != "new.json" {
		t.Fatalf("blank overrides should be ignored: %+v", dst)
	}
}

func TestConfigMergeBlacklistTrimAndDedupe(t *testing.T) {
	got := mergeBlacklist(
		[]string{" MaaFoo ", "MaaFoo", "", "   ", "MaaBar"},
		[]string{"MaaBar", " MaaBaz\t"},
	)
	want := map[string]struct{}{"MaaFoo": {}, "MaaBar": {}, "MaaBaz": {}}
	if len(got) != len(want) {
		t.Fatalf("mergeBlacklist() = %v, want %v", got, want)
	}
	for name := range want {
		if _, ok := got[name]; !ok {
			t.Fatalf("mergeBlacklist() missing %q in %v", name, got)
		}
	}
	for name := range got {
		if _, ok := want[name]; !ok {
			t.Fatalf("mergeBlacklist() unexpected %q in %v", name, got)
		}
	}
}
