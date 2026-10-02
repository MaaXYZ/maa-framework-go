package checker

import (
	"encoding/json"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPipelineConfigLoadAndMerge(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	text := "pipeline_schema: deps/tools/pipeline.schema.json\npipeline_exclusions:\n  action.Click.param.pressure: Intentional test omission\n"
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	src, err := loadConfigFromPath(path)
	if err != nil {
		t.Fatal(err)
	}
	dst := Config{HeaderDir: "existing/headers"}
	mergeConfig(&dst, src)
	if dst.HeaderDir != "existing/headers" || dst.PipelineSchema != "deps/tools/pipeline.schema.json" || dst.PipelineExclusions["action.Click.param.pressure"] != "Intentional test omission" {
		t.Fatalf("unexpected config: %+v", dst)
	}
}

// Run in a subprocess to exercise CLI flags and exit codes without changing
// the testing process's global flags, current directory, or stdout/stderr.
func TestPipelineRunHelper(t *testing.T) {
	if os.Getenv("MAA_PIPELINE_RUN_HELPER") != "1" {
		return
	}
	var args []string
	if err := json.Unmarshal([]byte(os.Getenv("MAA_PIPELINE_RUN_ARGS")), &args); err != nil {
		t.Fatal(err)
	}
	flag.CommandLine = flag.NewFlagSet("api-check", flag.ExitOnError)
	os.Args = append([]string{"api-check"}, args...)
	os.Exit(Run())
}

func writePipelineRunFixture(t *testing.T) (string, string) {
	t.Helper()
	s := pipelineFixtureSchema()
	fixtureProps(s, "Click")["pressure"] = schemaObject{"type": "integer"}
	dir, schema := writePipelineFixture(t, s, pipelineFixtureGo)
	files := map[string]string{
		"go.mod":                              "module " + repoRootModulePath + "\n\ngo 1.24\n",
		customControllerRel:                   "package maa\ntype CustomController interface { Foo() }\n",
		"deps/include/" + controllerHeaderRel: "struct MaaCustomControllerCallbacks { void (*foo)(void* trans_arg); };\n",
		"deps/include/" + maaDefHeaderRel:     "/* No method constants in this fixture. */\n",
		adbControllerRel:                      "package adb\n",
		win32ControllerRel:                    "package win32\n",
	}
	for _, paths := range nativeFilesByModule {
		for _, path := range paths {
			files[path] = "package native\n"
		}
	}
	for path, data := range files {
		path = filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return dir, schema
}

func TestPipelineCLI(t *testing.T) {
	dir, schema := writePipelineRunFixture(t)
	config := filepath.Join(dir, "enabled.yaml")
	if err := os.WriteFile(config, []byte("pipeline_schema: missing.json\n"), 0600); err != nil {
		t.Fatal(err)
	}
	blankReason := filepath.Join(dir, "blank.yaml")
	if err := os.WriteFile(blankReason, []byte("pipeline_schema: pipeline.schema.json\npipeline_exclusions:\n  action.Click.param.pressure: ''\n"), 0600); err != nil {
		t.Fatal(err)
	}
	disabled := filepath.Join(dir, "disabled.yaml")
	if err := os.WriteFile(disabled, []byte("pipeline_exclusions:\n  action.Click.param.pressure: Test omission\n"), 0600); err != nil {
		t.Fatal(err)
	}
	broken := filepath.Join(dir, "broken.json")
	if err := os.WriteFile(broken, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	unrecognized := filepath.Join(dir, "unrecognized.json")
	if err := os.WriteFile(unrecognized, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		args   []string
		code   int
		output string
	}{
		{"disabled by default", nil, 0, "pipeline_schema: <disabled>"},
		{"missing pressure", []string{"--pipeline-schema", schema}, 1, "action.Click.param.pressure: missing Go field"},
		{"missing schema from config", []string{"--config", config}, 2, "failed to check pipeline v2 coverage: read schema"},
		{"CLI overrides config", []string{"--config", config, "--pipeline-schema", schema}, 1, "action.Click.param.pressure: missing Go field"},
		{"invalid schema", []string{"--pipeline-schema", broken}, 2, "parse schema"},
		{"unrecognized schema", []string{"--pipeline-schema", unrecognized}, 2, "missing object $defs"},
		{"exclusions without schema", []string{"--config", disabled}, 2, "pipeline_exclusions requires pipeline_schema"},
		{"empty exclusion reason", []string{"--config", blankReason}, 2, "nonempty reason"},
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := json.Marshal(tt.args)
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(binary, "-test.run=^TestPipelineRunHelper$")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "MAA_PIPELINE_RUN_HELPER=1", "MAA_PIPELINE_RUN_ARGS="+string(args))
			output, err := cmd.CombinedOutput()
			code := 0
			if err != nil {
				if exit, ok := err.(*exec.ExitError); ok {
					code = exit.ExitCode()
				} else {
					t.Fatal(err)
				}
			}
			if code != tt.code || !strings.Contains(string(output), tt.output) {
				t.Fatalf("got exit %d, want %d; output:\n%s", code, tt.code, output)
			}
		})
	}
}

func TestPipelineCLIUnsupportedMarshal(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct{ name, body string }{{"unused closure", pipelineUnusedClosureMarshal}, {"ignored helper", pipelineIgnoredHelperMarshal}} {
		t.Run(tt.name, func(t *testing.T) {
			dir, schema := writePipelineRunFixture(t)
			if err := os.WriteFile(filepath.Join(dir, "pipeline.go"), []byte(replaceFixtureMarshal(pipelineFixtureGo, tt.body)), 0600); err != nil {
				t.Fatal(err)
			}
			args, err := json.Marshal([]string{"--pipeline-schema", schema})
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(binary, "-test.run=^TestPipelineRunHelper$")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "MAA_PIPELINE_RUN_HELPER=1", "MAA_PIPELINE_RUN_ARGS="+string(args))
			output, err := cmd.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 2 || !strings.Contains(string(output), "expected a direct JSON helper call") {
				t.Fatalf("got %v; output:\n%s", err, output)
			}
		})
	}
}
