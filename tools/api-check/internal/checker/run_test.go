package checker

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const runFixtureGoMod = "module " + repoRootModulePath + "\n\ngo 1.24\n"

// repoFixtureFiles builds the minimal repository layout that satisfies every
// required input path checked by run.
func repoFixtureFiles() map[string]string {
	files := map[string]string{
		"go.mod": runFixtureGoMod,
		customControllerRel: `package maa
 type CustomController interface { Foo() }
 type MaaCustomControllerCallbacks struct { Foo uintptr }
 var customControllerCallbacksHandle = new(MaaCustomControllerCallbacks)
 func init() { customControllerCallbacksHandle.Foo = purego.NewCallback(_FooAgent) }
 func _FooAgent(handle uintptr) uintptr { return 0 }
`,
		"deps/include/" + controllerHeaderRel:    "struct MaaCustomControllerCallbacks { void (*foo)(void* trans_arg); };\n",
		"deps/include/" + maaDefHeaderRel:        "",
		"deps/include/" + maaToolkitDefHeaderRel: "",
		"deps/include/MaaFramework/MaaMsg.h":     "",
		"event.go":                               "package maa\n",
		adbControllerRel:                         "package adb\n",
		win32ControllerRel:                       "package win32\n",
	}
	for _, paths := range nativeFilesByModule {
		for _, path := range paths {
			files[path] = "package native\n"
		}
	}
	files["internal/native/native.go"] = `package native
var libraries = []Library{
 {handle: &maaFramework, entries: frameworkEntries},
 {handle: &maaToolkit, entries: toolkitEntries},
 {handle: &maaAgentServer, entries: agentServerEntries},
 {handle: &maaAgentClient, entries: agentClientEntries},
}
`
	for _, spec := range constantFamilySpecs {
		if files[spec.goFile] == "" {
			files[spec.goFile] = "package fixture\n"
		}
		files[spec.goFile] += fmt.Sprintf("const %sFixture = 0\n", spec.goPrefix)
		files["deps/include/"+maaDefHeaderRel] += fmt.Sprintf("#define %sFixture 0\n", spec.cPrefix)
	}
	for i, module := range moduleOrder {
		path := nativeFilesByModule[module][0]
		table := []string{"frameworkEntries", "toolkitEntries", "agentServerEntries", "agentClientEntries"}[i]
		symbol := "Maa" + []string{"Framework", "Toolkit", "AgentServer", "AgentClient"}[i] + "Fixture"
		files[path] += fmt.Sprintf("var %s func()\nvar %s = []Entry{{&%s, %q}}\n", symbol, table, symbol, symbol)
		macro := []string{"MAA_FRAMEWORK_API", "MAA_TOOLKIT_API", "MAA_AGENT_SERVER_API", "MAA_AGENT_CLIENT_API"}[i]
		files["deps/include/"+module+".h"] = macro + " void " + symbol + "(void);\n"
	}
	return files
}

func writeFixtureFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for path, data := range files {
		writeFixtureFile(t, filepath.Join(dir, path), data)
	}
}

func writeRepoFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFixtureFiles(t, dir, repoFixtureFiles())
	return dir
}

func writeRepoFixtureWith(t *testing.T, extra map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	files := repoFixtureFiles()
	for path, data := range extra {
		files[path] = data
	}
	writeFixtureFiles(t, dir, files)
	return dir
}

// runChecker runs the checker from dir and captures its streams.
func runChecker(t *testing.T, dir string, args ...string) (int, string, string) {
	t.Helper()
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// repoRootForTest mirrors run's own root detection so path assertions stay
// correct when the temporary directory is reached through a symlink.
func repoRootForTest(t *testing.T) string {
	t.Helper()
	root, err := detectRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestRunCleanRepoExitsZero(t *testing.T) {
	dir := writeRepoFixture(t)
	code, stdout, stderr := runChecker(t, dir)
	if code != 0 {
		t.Fatalf("run() = %d, want 0; stderr:\n%s\nstdout:\n%s", code, stderr, stdout)
	}
	root := repoRootForTest(t)
	for _, want := range []string{
		"config: <none> (using defaults)",
		"repo_root: " + root,
		"header_dir: " + filepath.Join(root, defaultHeaderDirRel),
		"blacklist_size: 0",
		"pipeline_schema: <disabled>",
		"PASS: no inconsistencies found.",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout)
		}
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestRunCPlusPlusGuardUsesCConstantValue(t *testing.T) {
	files := repoFixtureFiles()
	headerPath := "deps/include/" + maaDefHeaderRel
	goPath := "internal/native/framework.go"
	goContent := files[goPath] + "const MaaInferenceExecutionProvider_CPU = 99\n"
	dir := writeRepoFixtureWith(t, map[string]string{
		headerPath: files[headerPath] + `enum MaaInferenceExecutionProviderEnum {
#if defined(__cplusplus)
    MaaInferenceExecutionProvider_CPU = 1,
#else
    MaaInferenceExecutionProvider_CPU = 99,
#endif
};
`,
		goPath: goContent,
	})

	code, stdout, stderr := runChecker(t, dir)
	if code != 0 || !strings.Contains(stdout, "PASS: no inconsistencies found.") || stderr != "" {
		t.Errorf("run() with the C value = %d, want 0; stderr:\n%s\nstdout:\n%s", code, stderr, stdout)
	}

	writeFixtureFile(t, filepath.Join(dir, goPath), strings.Replace(goContent, "MaaInferenceExecutionProvider_CPU = 99", "MaaInferenceExecutionProvider_CPU = 1", 1))
	code, stdout, stderr = runChecker(t, dir)
	if code != 1 || !strings.Contains(stdout, "[native.execution_provider] constant value mismatch: CPU (go=1 c=99)") || stderr != "" {
		t.Fatalf("run() with the C++ value = %d, want constant mismatch and exit 1; stderr:\n%s\nstdout:\n%s", code, stderr, stdout)
	}
	if strings.Contains(stdout, "PASS: no inconsistencies found.") {
		t.Fatalf("run() passed despite the C value mismatch:\n%s", stdout)
	}
}

func TestRunMixedUnknownCGuardsExitTwo(t *testing.T) {
	tests := []struct {
		name      string
		condition string
		goValue   int
	}{
		{"defined AND", "!defined(__cplusplus) && defined(MAA_REVIEW_UNKNOWN)", 1},
		{"defined OR", "defined(__cplusplus) || defined(MAA_REVIEW_UNKNOWN)", 99},
		{"bare unknown operand", "!defined(__cplusplus) && MAA_REVIEW_UNKNOWN", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := repoFixtureFiles()
			headerPath := "deps/include/" + maaDefHeaderRel
			goPath := "internal/native/framework.go"
			dir := writeRepoFixtureWith(t, map[string]string{
				headerPath: files[headerPath] + fmt.Sprintf(`#define MAA_REVIEW_UNKNOWN 1
enum MaaInferenceExecutionProviderEnum {
#if %s
    MaaInferenceExecutionProvider_CPU = 99,
#else
    MaaInferenceExecutionProvider_CPU = 1,
#endif
};
`, tt.condition),
				goPath: files[goPath] + fmt.Sprintf("const MaaInferenceExecutionProvider_CPU = %d\n", tt.goValue),
			})

			code, stdout, stderr := runChecker(t, dir)
			if code != 2 || !strings.Contains(stderr, "conflicting C constant MaaInferenceExecutionProvider_CPU") {
				t.Fatalf("run() = %d, want conflicting C constant and exit 2; stderr:\n%s\nstdout:\n%s", code, stderr, stdout)
			}
			if strings.Contains(stdout, "PASS: no inconsistencies found.") {
				t.Fatalf("run() passed despite conflicting C constant values:\n%s", stdout)
			}
		})
	}
}

func TestRunUnknownConditionalMacroConflictExitsTwo(t *testing.T) {
	files := repoFixtureFiles()
	headerPath := "deps/include/" + maaDefHeaderRel
	goPath := "internal/native/framework.go"
	header := `#if MAA_REVIEW_UNKNOWN
#define MaaInferenceExecutionProvider_CPU 1
#else
#define MaaInferenceExecutionProvider_CPU 1
#endif
`
	dir := writeRepoFixtureWith(t, map[string]string{
		headerPath: files[headerPath] + header,
		goPath:     files[goPath] + "const MaaInferenceExecutionProvider_CPU = 1\n",
	})
	code, stdout, stderr := runChecker(t, dir)
	if code != 0 || stderr != "" || !strings.Contains(stdout, "PASS: no inconsistencies found.") {
		t.Fatalf("equivalent branch macros: exit=%d, stderr=%s, stdout=%s", code, stderr, stdout)
	}
	// Keep the last declaration matching Go, isolating the declaration that
	// the old map extractor silently overwrote.
	header = strings.Replace(header, "MaaInferenceExecutionProvider_CPU 1", "MaaInferenceExecutionProvider_CPU 2", 1)
	writeFixtureFile(t, filepath.Join(dir, headerPath), files[headerPath]+header)
	code, stdout, stderr = runChecker(t, dir)
	if code != 2 || !strings.Contains(stderr, "conflicting C constant MaaInferenceExecutionProvider_CPU") || strings.Contains(stdout, "PASS:") {
		t.Fatalf("conflicting branch macros: exit=%d, stderr=%s, stdout=%s", code, stderr, stdout)
	}
}

func TestRunUnknownConditionalEnumSequencingExitsTwo(t *testing.T) {
	files := repoFixtureFiles()
	headerPath := "deps/include/" + maaDefHeaderRel
	goPath := "internal/native/framework.go"
	explicitEnum := `enum MaaInferenceExecutionProviderEnum {
    MaaInferenceExecutionProvider_A = 0,
#if MAA_REVIEW_UNKNOWN
    MaaInferenceExecutionProvider_B = 1,
#else
    MaaInferenceExecutionProvider_C = 1,
#endif
    MaaInferenceExecutionProvider_Tail = 2,
};
`
	dir := writeRepoFixtureWith(t, map[string]string{
		headerPath: files[headerPath] + explicitEnum,
		goPath: files[goPath] + `const (
    MaaInferenceExecutionProvider_A = 0
    MaaInferenceExecutionProvider_B = 1
    MaaInferenceExecutionProvider_C = 1
    MaaInferenceExecutionProvider_Tail = 2
)
`,
	})

	code, stdout, stderr := runChecker(t, dir)
	if code != 0 || !strings.Contains(stdout, "PASS: no inconsistencies found.") || stderr != "" {
		t.Fatalf("run() with explicit unknown branches = %d, want 0; stderr:\n%s\nstdout:\n%s", code, stderr, stdout)
	}

	implicitEnum := strings.NewReplacer(
		"MaaInferenceExecutionProvider_B = 1", "MaaInferenceExecutionProvider_B",
		"MaaInferenceExecutionProvider_C = 1", "MaaInferenceExecutionProvider_C",
		"MaaInferenceExecutionProvider_Tail = 2", "MaaInferenceExecutionProvider_Tail",
	).Replace(explicitEnum)
	writeFixtureFile(t, filepath.Join(dir, headerPath), files[headerPath]+implicitEnum)
	code, stdout, stderr = runChecker(t, dir)
	if code != 2 || !strings.Contains(stderr, "unknown conditional enum sequencing") {
		t.Fatalf("run() with implicit unknown branches = %d, want sequencing diagnostic and exit 2; stderr:\n%s\nstdout:\n%s", code, stderr, stdout)
	}
	if strings.Contains(stdout, "PASS: no inconsistencies found.") {
		t.Fatalf("run() passed despite unsupported conditional enum sequencing:\n%s", stdout)
	}
}

func TestRunUnknownCGuardElifReachability(t *testing.T) {
	for _, tc := range []struct {
		name   string
		header string
	}{
		{
			name: "known true elif consumes fallthrough",
			header: `enum MaaInferenceExecutionProviderEnum {
#if MAA_REVIEW_UNKNOWN
    MaaInferenceExecutionProvider_CPU = 1,
#elif !defined(__cplusplus)
    MaaInferenceExecutionProvider_CPU = 1,
#else
    MaaInferenceExecutionProvider_CPU = 99,
#endif
};
`,
		},
		{
			name: "known false elif is omitted",
			header: `enum MaaInferenceExecutionProviderEnum {
#if MAA_REVIEW_UNKNOWN
    MaaInferenceExecutionProvider_CPU = 1,
#elif defined(__cplusplus)
    MaaInferenceExecutionProvider_CPU = 99,
#else
    MaaInferenceExecutionProvider_CPU = 1,
#endif
};
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := repoFixtureFiles()
			headerPath := "deps/include/" + maaDefHeaderRel
			goPath := "internal/native/framework.go"
			dir := writeRepoFixtureWith(t, map[string]string{
				headerPath: files[headerPath] + tc.header,
				goPath:     files[goPath] + "const MaaInferenceExecutionProvider_CPU = 1\n",
			})

			code, stdout, stderr := runChecker(t, dir)
			if code != 0 || !strings.Contains(stdout, "PASS: no inconsistencies found.") || stderr != "" {
				t.Fatalf("run() with unreachable divergent branch = %d, want 0; stderr:\n%s\nstdout:\n%s", code, stderr, stdout)
			}

			conflictingHeader := strings.Replace(tc.header, "MaaInferenceExecutionProvider_CPU = 1", "MaaInferenceExecutionProvider_CPU = 2", 1)
			writeFixtureFile(t, filepath.Join(dir, headerPath), files[headerPath]+conflictingHeader)
			code, stdout, stderr = runChecker(t, dir)
			if code != 2 || !strings.Contains(stderr, "conflicting C constant MaaInferenceExecutionProvider_CPU") {
				t.Fatalf("run() with divergent reachable branches = %d, want conflict and exit 2; stderr:\n%s\nstdout:\n%s", code, stderr, stdout)
			}
			if strings.Contains(stdout, "PASS: no inconsistencies found.") {
				t.Fatalf("run() passed despite conflicting reachable branches:\n%s", stdout)
			}
		})
	}
}

func TestRunConfigPrecedence(t *testing.T) {
	t.Run("cwd config wins over root fallback", func(t *testing.T) {
		dir := writeRepoFixture(t)
		writeFixtureFile(t, filepath.Join(dir, apiCheckConfigPathRel), "header_dir: missing-headers\n")
		nested := filepath.Join(dir, "nested")
		writeFixtureFile(t, filepath.Join(nested, autoConfigFileName), "header_dir: deps/include\n")

		code, stdout, stderr := runChecker(t, nested)
		if code != 0 {
			t.Fatalf("run() = %d, want 0; stderr:\n%s", code, stderr)
		}
		if !strings.Contains(stdout, "config: config.yaml\n") {
			t.Fatalf("cwd config not reported:\n%s", stdout)
		}
		if !strings.Contains(stdout, "repo_root: "+repoRootForTest(t)+"\n") {
			t.Fatalf("root not detected from subdirectory:\n%s", stdout)
		}
	})

	t.Run("root fallback used without cwd config", func(t *testing.T) {
		dir := writeRepoFixture(t)
		writeFixtureFile(t, filepath.Join(dir, apiCheckConfigPathRel), "header_dir: deps/include\n")
		nested := filepath.Join(dir, "nested", "deep")
		if err := os.MkdirAll(nested, 0700); err != nil {
			t.Fatal(err)
		}

		code, stdout, stderr := runChecker(t, nested)
		if code != 0 {
			t.Fatalf("run() = %d, want 0; stderr:\n%s", code, stderr)
		}
		root := repoRootForTest(t)
		if !strings.Contains(stdout, "config: "+resolveFromRepoRoot(root, apiCheckConfigPathRel)+"\n") {
			t.Fatalf("root fallback config not reported:\n%s", stdout)
		}
		if !strings.Contains(stdout, "header_dir: "+filepath.Join(root, defaultHeaderDirRel)+"\n") {
			t.Fatalf("header_dir not resolved against repo root:\n%s", stdout)
		}
	})
}

func TestRunExplicitConfigAndHeaderOverride(t *testing.T) {
	dir := writeRepoFixture(t)
	writeFixtureFile(t, filepath.Join(dir, autoConfigFileName), "header_dir: missing-headers\n")

	code, _, stderr := runChecker(t, dir)
	if code != 2 || !strings.Contains(stderr, "failed to resolve required input paths") {
		t.Fatalf("run() = %d, stderr = %q; want input error", code, stderr)
	}

	code, stdout, stderr := runChecker(t, dir, "--header-dir", "deps/include")
	if code != 0 {
		t.Fatalf("run() with --header-dir = %d, stderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "config: config.yaml\n") {
		t.Fatalf("auto config should still be loaded:\n%s", stdout)
	}

	explicit := writeFixtureFile(t, filepath.Join(dir, "explicit.yaml"), "header_dir: deps/include\n")
	code, stdout, stderr = runChecker(t, dir, "--config", explicit)
	if code != 0 {
		t.Fatalf("run() with --config = %d, stderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "config: "+filepath.Clean(explicit)+"\n") {
		t.Fatalf("explicit config not reported:\n%s", stdout)
	}

	code, _, stderr = runChecker(t, dir, "--config", filepath.Join(dir, "missing.yaml"))
	if code != 2 || !strings.Contains(stderr, "failed to load config") || !strings.Contains(stderr, "read") {
		t.Fatalf("run() with missing --config = %d, stderr = %q; want load error", code, stderr)
	}
}

func TestRunPipelineSchemaOverride(t *testing.T) {
	dir, schema := writePipelineRunFixture(t)

	code, stdout, stderr := runChecker(t, dir)
	if code != 0 || !strings.Contains(stdout, "pipeline_schema: <disabled>") {
		t.Fatalf("run() = %d, output:\n%s\nstderr:\n%s", code, stdout, stderr)
	}

	code, stdout, stderr = runChecker(t, dir, "--pipeline-schema", schema)
	if code != 1 || !strings.Contains(stdout, "action.Click.param.pressure: missing Go field") {
		t.Fatalf("run() with schema = %d, output:\n%s\nstderr:\n%s", code, stdout, stderr)
	}

	writeFixtureFile(t, filepath.Join(dir, autoConfigFileName), "pipeline_schema: missing.json\n")
	code, _, stderr = runChecker(t, dir)
	if code != 2 || !strings.Contains(stderr, "failed to check pipeline v2 coverage") {
		t.Fatalf("run() with broken config schema = %d, stderr = %q", code, stderr)
	}

	code, stdout, stderr = runChecker(t, dir, "--pipeline-schema", schema)
	if code != 1 || !strings.Contains(stdout, "action.Click.param.pressure: missing Go field") {
		t.Fatalf("run() with --pipeline-schema override = %d, output:\n%s\nstderr:\n%s", code, stdout, stderr)
	}

	writeFixtureFile(t, filepath.Join(dir, autoConfigFileName), "pipeline_exclusions:\n  action.Click.param.pressure: reason\n")
	code, _, stderr = runChecker(t, dir)
	if code != 2 || !strings.Contains(stderr, "pipeline_exclusions requires pipeline_schema") {
		t.Fatalf("run() with exclusions only = %d, stderr = %q", code, stderr)
	}
}

func TestRunRootRelativeSchemaPath(t *testing.T) {
	dir, schema := writePipelineRunFixture(t)
	nested := filepath.Join(dir, "nested", "deep")
	if err := os.MkdirAll(nested, 0700); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runChecker(t, nested, "--pipeline-schema", filepath.Base(schema))
	if code != 1 {
		t.Fatalf("run() = %d, want 1; stderr:\n%s\nstdout:\n%s", code, stderr, stdout)
	}
	want := resolveFromRepoRoot(repoRootForTest(t), filepath.Base(schema))
	if !strings.Contains(stdout, "pipeline_schema: "+want+"\n") {
		t.Fatalf("pipeline schema not resolved against repo root:\n%s", stdout)
	}
}

func TestRunDifferencesExitOne(t *testing.T) {
	dir := writeRepoFixtureWith(t, map[string]string{
		customControllerRel: strings.Replace(repoFixtureFiles()[customControllerRel], "Foo() }", "Foo(); Bar() }", 1),
	})

	code, stdout, stderr := runChecker(t, dir)
	if code != 1 {
		t.Fatalf("run() = %d, want 1; stderr:\n%s\nstdout:\n%s", code, stderr, stdout)
	}
	for _, want := range []string{
		"## " + sectionController,
		"Go interface method not found in C callbacks: Bar",
		"FAIL: found 1 inconsistency(s).",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout)
		}
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestRunInputErrorsExitTwo(t *testing.T) {
	t.Run("missing header dir", func(t *testing.T) {
		dir := writeRepoFixture(t)
		code, _, stderr := runChecker(t, dir, "--header-dir", "missing/headers")
		if code != 2 || !strings.Contains(stderr, "failed to resolve required input paths") {
			t.Fatalf("run() = %d, stderr = %q", code, stderr)
		}
	})

	t.Run("missing custom controller source", func(t *testing.T) {
		dir := t.TempDir()
		files := repoFixtureFiles()
		delete(files, customControllerRel)
		writeFixtureFiles(t, dir, files)

		code, _, stderr := runChecker(t, dir)
		if code != 2 || !strings.Contains(stderr, "custom controller source") {
			t.Fatalf("run() = %d, stderr = %q", code, stderr)
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		dir := writeRepoFixture(t)
		bad := writeFixtureFile(t, filepath.Join(dir, "bad.yaml"), "unknown_key: 1\n")
		code, _, stderr := runChecker(t, dir, "--config", bad)
		if code != 2 || !strings.Contains(stderr, "failed to load config") || !strings.Contains(stderr, "field unknown_key not found") {
			t.Fatalf("run() = %d, stderr = %q", code, stderr)
		}
	})

	t.Run("multiple document config", func(t *testing.T) {
		dir := writeRepoFixture(t)
		bad := writeFixtureFile(t, filepath.Join(dir, "multi.yaml"), "header_dir: a\n---\nheader_dir: b\n")
		code, _, stderr := runChecker(t, dir, "--config", bad)
		if code != 2 || !strings.Contains(stderr, "multiple YAML documents") {
			t.Fatalf("run() = %d, stderr = %q", code, stderr)
		}
	})

	t.Run("not a repository", func(t *testing.T) {
		code, _, stderr := runChecker(t, t.TempDir())
		if code != 2 || !strings.Contains(stderr, "failed to detect repository root") {
			t.Fatalf("run() = %d, stderr = %q", code, stderr)
		}
	})
}

func TestRunHelpExitsZero(t *testing.T) {
	dir := writeRepoFixture(t)
	for _, arg := range []string{"--help", "-h"} {
		t.Run(arg, func(t *testing.T) {
			code, stdout, stderr := runChecker(t, dir, arg)
			if code != 0 {
				t.Fatalf("run(%s) = %d, want 0", arg, code)
			}
			if !strings.Contains(stderr, "Usage: api-check") || !strings.Contains(stderr, "-config") {
				t.Fatalf("help output missing flags:\n%s", stderr)
			}
			if stdout != "" {
				t.Fatalf("help wrote to stdout: %q", stdout)
			}
		})
	}
}

func TestRunInvalidFlagsExitTwo(t *testing.T) {
	dir := writeRepoFixture(t)

	t.Run("unknown flag", func(t *testing.T) {
		code, stdout, stderr := runChecker(t, dir, "--not-a-flag")
		if code != 2 || !strings.Contains(stderr, "flag provided but not defined") {
			t.Fatalf("run() = %d, stderr = %q", code, stderr)
		}
		if stdout != "" {
			t.Fatalf("stdout = %q, want empty", stdout)
		}
	})

	t.Run("missing flag value", func(t *testing.T) {
		code, _, stderr := runChecker(t, dir, "--config")
		if code != 2 || !strings.Contains(stderr, "flag needs an argument") {
			t.Fatalf("run() = %d, stderr = %q", code, stderr)
		}
	})
}

func TestRunPositionalArgumentsExitTwo(t *testing.T) {
	dir := writeRepoFixture(t)
	tests := []struct {
		name string
		args []string
	}{
		{"positional after flags", []string{"--header-dir", "deps/include", "extra"}},
		{"positional before flags", []string{"extra", "--header-dir", "deps/include"}},
		{"positional after terminator", []string{"--", "extra"}},
		{"positional before config error", []string{"--config", "does-not-exist.yaml", "extra"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr := runChecker(t, dir, tt.args...)
			if code != 2 || !strings.Contains(stderr, "unexpected positional argument") {
				t.Fatalf("run(%v) = %d, stderr = %q", tt.args, code, stderr)
			}
			if strings.Contains(stderr, "failed to load config") {
				t.Fatalf("positional error should precede config loading:\n%s", stderr)
			}
			if stdout != "" {
				t.Fatalf("stdout = %q, want empty", stdout)
			}
		})
	}
}

func TestRunIgnoresGlobalFlagState(t *testing.T) {
	dir := writeRepoFixture(t)

	original := flag.CommandLine
	defer func() { flag.CommandLine = original }()
	flag.CommandLine = flag.NewFlagSet("poisoned", flag.ExitOnError)
	flag.CommandLine.Bool("poison", false, "must be ignored by run")

	code, stdout, stderr := runChecker(t, dir, "--header-dir", "deps/include")
	if code != 0 {
		t.Fatalf("run() = %d, stderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "PASS: no inconsistencies found.") {
		t.Fatalf("run() did not use its own flag set:\n%s", stdout)
	}
}

func TestRunBlacklistTrimAndDedupe(t *testing.T) {
	dir := writeRepoFixtureWith(t, map[string]string{"deps/include/omitted.h": "MAA_FRAMEWORK_API void MaaFoo(void);\nMAA_FRAMEWORK_API void MaaBar(void);\n"})

	code, stdout, stderr := runChecker(t, dir, "--blacklist", " MaaFoo ", "--blacklist", "MaaFoo", "--blacklist", "   ")
	if code != 1 || !strings.Contains(stdout, "MaaBar") {
		t.Fatalf("run() = %d, stderr:\n%s\nstdout:\n%s", code, stderr, stdout)
	}
	if !strings.Contains(stdout, "blacklist_size: 1\n") {
		t.Fatalf("CLI blacklist not trimmed/deduped:\n%s", stdout)
	}

	cfg := writeFixtureFile(t, filepath.Join(dir, "blacklist.yaml"), "blacklist:\n  - \" MaaFoo \"\n  - MaaBar\n")
	code, stdout, stderr = runChecker(t, dir, "--config", cfg, "--blacklist", " MaaBar ")
	if code != 0 {
		t.Fatalf("run() = %d, stderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "blacklist_size: 2\n") {
		t.Fatalf("config and CLI blacklist not merged/deduped:\n%s", stdout)
	}
}
