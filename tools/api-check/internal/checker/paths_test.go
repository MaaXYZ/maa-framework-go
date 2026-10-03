package checker

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadGoModulePathValid(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"plain", "module github.com/MaaXYZ/maa-framework-go/v4\n", repoRootModulePath},
		{"tabs and extra spaces", "module\t  github.com/MaaXYZ/maa-framework-go/v4   \n", repoRootModulePath},
		{"trailing comment", "module github.com/MaaXYZ/maa-framework-go/v4 // pinned\n", repoRootModulePath},
		{"leading comment", "// module example.com/fake\nmodule github.com/MaaXYZ/maa-framework-go/v4\n", repoRootModulePath},
		{"quoted", "module \"github.com/MaaXYZ/maa-framework-go/v4\"\n", repoRootModulePath},
		{"backquoted", "module `github.com/MaaXYZ/maa-framework-go/v4`\n", repoRootModulePath},
		{"comment after quoted", "module \"github.com/MaaXYZ/maa-framework-go/v4\" // http://example.com\n", repoRootModulePath},
		{"after other directives", "go 1.24\n\nrequire example.com/dep v1.0.0\n\nmodule github.com/MaaXYZ/maa-framework-go/v4\n", repoRootModulePath},
		{"module-like token ignored", "modules example.com/other\nmodule github.com/MaaXYZ/maa-framework-go/v4\n", repoRootModulePath},
		{"block comment line ignored", "/* module example.com/fake */\nmodule github.com/MaaXYZ/maa-framework-go/v4\n", repoRootModulePath},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeFixtureFile(t, filepath.Join(t.TempDir(), "go.mod"), tt.body)
			got, err := readGoModulePath(path)
			if err != nil {
				t.Fatalf("readGoModulePath() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("readGoModulePath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReadGoModulePathInvalid(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"empty file", "", "module declaration not found"},
		{"no module directive", "go 1.24\n", "module declaration not found"},
		{"module without path", "module\n", "requires exactly one path"},
		{"module with two paths", "module a b\n", "requires exactly one path"},
		{"unterminated double quote", "module \"example.com/x\n", "unterminated quoted string"},
		{"unterminated backquote", "module `example.com/x\n", "unterminated quoted string"},
		{"bad escape", "module \"example.com/\\x\"\n", "invalid quoted string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeFixtureFile(t, filepath.Join(t.TempDir(), "go.mod"), tt.body)
			_, err := readGoModulePath(path)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("readGoModulePath() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestReadGoModulePathInputErrors(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		_, err := readGoModulePath(filepath.Join(t.TempDir(), "go.mod"))
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("readGoModulePath() error = %v, want not-exist", err)
		}
	})

	t.Run("go.mod is a directory", func(t *testing.T) {
		_, err := readGoModulePath(t.TempDir())
		if err == nil || errors.Is(err, os.ErrNotExist) {
			t.Fatalf("readGoModulePath() error = %v, want read error", err)
		}
	})
}

func TestDetectRepoRootFrom(t *testing.T) {
	t.Run("root with matching module", func(t *testing.T) {
		root := t.TempDir()
		writeFixtureFile(t, filepath.Join(root, "go.mod"), "module "+repoRootModulePath+"\n")

		got, err := detectRepoRootFrom(root)
		if err != nil {
			t.Fatalf("detectRepoRootFrom() error = %v", err)
		}
		if got != filepath.Clean(root) {
			t.Fatalf("detectRepoRootFrom() = %q, want %q", got, filepath.Clean(root))
		}
	})

	t.Run("subdirectory walks up to root", func(t *testing.T) {
		root := t.TempDir()
		writeFixtureFile(t, filepath.Join(root, "go.mod"), "module "+repoRootModulePath+"\n")
		nested := filepath.Join(root, "tools", "api-check", "internal")
		if err := os.MkdirAll(nested, 0700); err != nil {
			t.Fatal(err)
		}

		got, err := detectRepoRootFrom(nested)
		if err != nil {
			t.Fatalf("detectRepoRootFrom() error = %v", err)
		}
		if got != filepath.Clean(root) {
			t.Fatalf("detectRepoRootFrom() = %q, want %q", got, filepath.Clean(root))
		}
	})

	t.Run("nested module does not shadow outer root", func(t *testing.T) {
		root := t.TempDir()
		writeFixtureFile(t, filepath.Join(root, "go.mod"), "module "+repoRootModulePath+"\n")
		nested := filepath.Join(root, "examples", "nested")
		writeFixtureFile(t, filepath.Join(nested, "go.mod"), "module example.com/nested\n")

		got, err := detectRepoRootFrom(nested)
		if err != nil {
			t.Fatalf("detectRepoRootFrom() error = %v", err)
		}
		if got != filepath.Clean(root) {
			t.Fatalf("detectRepoRootFrom() = %q, want outer root %q", got, filepath.Clean(root))
		}
	})

	t.Run("worktree fallback without go.mod", func(t *testing.T) {
		root := t.TempDir()
		for _, rel := range []string{".git", filepath.Join("tools", "api-check"), filepath.Join("internal", "native")} {
			if err := os.MkdirAll(filepath.Join(root, rel), 0700); err != nil {
				t.Fatal(err)
			}
		}
		writeFixtureFile(t, filepath.Join(root, customControllerRel), "package maa\n")

		got, err := detectRepoRootFrom(filepath.Join(root, "internal", "native"))
		if err != nil {
			t.Fatalf("detectRepoRootFrom() error = %v", err)
		}
		if got != filepath.Clean(root) {
			t.Fatalf("detectRepoRootFrom() = %q, want worktree root %q", got, filepath.Clean(root))
		}
	})

	t.Run("worktree markers incomplete", func(t *testing.T) {
		root := t.TempDir()
		writeFixtureFile(t, filepath.Join(root, ".git"), "gitdir: elsewhere\n")
		if err := os.MkdirAll(filepath.Join(root, "tools", "api-check"), 0700); err != nil {
			t.Fatal(err)
		}

		_, err := detectRepoRootFrom(root)
		if err == nil || !strings.Contains(err.Error(), "repository root not found") {
			t.Fatalf("detectRepoRootFrom() error = %v, want repository-root-not-found", err)
		}
	})

	t.Run("mismatched module without markers", func(t *testing.T) {
		root := t.TempDir()
		writeFixtureFile(t, filepath.Join(root, "go.mod"), "module example.com/other\n")

		_, err := detectRepoRootFrom(root)
		if err == nil || !strings.Contains(err.Error(), "repository root not found") {
			t.Fatalf("detectRepoRootFrom() error = %v, want repository-root-not-found", err)
		}
	})

	t.Run("missing module declaration is an input error", func(t *testing.T) {
		root := t.TempDir()
		writeFixtureFile(t, filepath.Join(root, "go.mod"), "go 1.24\n")

		_, err := detectRepoRootFrom(root)
		if err == nil || !strings.Contains(err.Error(), "module declaration not found") {
			t.Fatalf("detectRepoRootFrom() error = %v, want module-declaration error", err)
		}
	})

	t.Run("go.mod directory is an input error", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, "go.mod"), 0700); err != nil {
			t.Fatal(err)
		}

		_, err := detectRepoRootFrom(root)
		if err == nil || !strings.Contains(err.Error(), "read go.mod") {
			t.Fatalf("detectRepoRootFrom() error = %v, want read-go.mod error", err)
		}
	})
}

func TestDetectRepoRootUsesWorkingDirectory(t *testing.T) {
	root := t.TempDir()
	writeFixtureFile(t, filepath.Join(root, "go.mod"), "module "+repoRootModulePath+"\n")
	nested := filepath.Join(root, "tools", "api-check")
	if err := os.MkdirAll(nested, 0700); err != nil {
		t.Fatal(err)
	}
	t.Chdir(nested)

	got, err := detectRepoRoot()
	if err != nil {
		t.Fatalf("detectRepoRoot() error = %v", err)
	}
	wantInfo, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	gotInfo, err := os.Stat(got)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(gotInfo, wantInfo) {
		t.Fatalf("detectRepoRoot() = %q, want the directory %q", got, root)
	}
}

func TestResolveFromRepoRoot(t *testing.T) {
	root := t.TempDir()
	absolute := filepath.Join(root, "elsewhere", "..", "absolute")

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"relative", filepath.Join("a", "b"), filepath.Join(root, "a", "b")},
		{"relative with parent", filepath.Join("a", "..", "b"), filepath.Join(root, "b")},
		{"absolute", absolute, filepath.Clean(absolute)},
		{"dot", ".", filepath.Clean(root)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveFromRepoRoot(root, tt.input); got != tt.want {
				t.Fatalf("resolveFromRepoRoot(%q, %q) = %q, want %q", root, tt.input, got, tt.want)
			}
		})
	}
}

func TestResolveHeaderDirPath(t *testing.T) {
	root := t.TempDir()
	absolute := filepath.Join(root, "custom", "include")

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty uses default", "", filepath.Join(root, defaultHeaderDirRel)},
		{"blank uses default", "   \t", filepath.Join(root, defaultHeaderDirRel)},
		{"relative", filepath.Join("deps", "include"), filepath.Join(root, "deps", "include")},
		{"trimmed", "  deps/include  ", filepath.Join(root, "deps", "include")},
		{"absolute", absolute, filepath.Clean(absolute)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveHeaderDir(root, tt.input); got != tt.want {
				t.Fatalf("resolveHeaderDir(%q, %q) = %q, want %q", root, tt.input, got, tt.want)
			}
		})
	}
}

func TestResolveNativeFilePaths(t *testing.T) {
	root := t.TempDir()
	got := resolveNativeFiles(root)
	if len(got) != len(nativeFilesByModule) {
		t.Fatalf("resolveNativeFiles() has %d modules, want %d", len(got), len(nativeFilesByModule))
	}
	for module, files := range nativeFilesByModule {
		resolved := got[module]
		if len(resolved) != len(files) {
			t.Fatalf("module %s has %d files, want %d", module, len(resolved), len(files))
		}
		for i, file := range files {
			if want := filepath.Join(root, file); resolved[i] != want {
				t.Fatalf("module %s file %d = %q, want %q", module, i, resolved[i], want)
			}
		}
	}
}

func TestPathExists(t *testing.T) {
	dir := t.TempDir()
	file := writeFixtureFile(t, filepath.Join(dir, "file.txt"), "x")
	if !pathExists(dir) || !pathExists(file) {
		t.Fatal("pathExists() = false for existing paths")
	}
	if pathExists(filepath.Join(dir, "missing")) {
		t.Fatal("pathExists() = true for missing path")
	}
}

func TestRequirePathKinds(t *testing.T) {
	dir := t.TempDir()
	file := writeFixtureFile(t, filepath.Join(dir, "file.txt"), "x")

	if err := requireDir(dir); err != nil {
		t.Fatalf("requireDir(dir) error = %v", err)
	}
	if err := requireDir(file); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("requireDir(file) error = %v, want not-a-directory", err)
	}
	if err := requireDir(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("requireDir(missing) error = nil")
	}

	if err := requireFile(file); err != nil {
		t.Fatalf("requireFile(file) error = %v", err)
	}
	if err := requireFile(dir); err == nil || !strings.Contains(err.Error(), "is a directory") {
		t.Fatalf("requireFile(dir) error = %v, want is-a-directory", err)
	}
	if err := requireFile(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("requireFile(missing) error = nil")
	}
}
