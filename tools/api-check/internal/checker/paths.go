package checker

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// detectRepoRoot walks up from the current working directory looking for the
// maa-framework-go repository root.
func detectRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get cwd: %w", err)
	}
	return detectRepoRootFrom(cwd)
}

// detectRepoRootFrom walks up from start looking for a directory whose go.mod
// declares repoRootModulePath, or a worktree-shaped root when go.mod is
// unavailable.
func detectRepoRootFrom(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("get absolute cwd: %w", err)
	}

	for {
		ok, err := isRepoRoot(dir)
		if err != nil {
			return "", err
		}
		if ok {
			return filepath.Clean(dir), nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", errors.New("repository root not found from current working directory")
}

func isRepoRoot(dir string) (bool, error) {
	modulePath, err := readGoModulePath(filepath.Join(dir, "go.mod"))
	if err == nil {
		return modulePath == repoRootModulePath, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("read go.mod under %s: %w", filepath.Clean(dir), err)
	}

	// Fallback for worktrees where go.mod may be unavailable.
	if pathExists(filepath.Join(dir, ".git")) &&
		pathExists(filepath.Join(dir, "tools", "api-check")) &&
		pathExists(filepath.Join(dir, "internal", "native")) &&
		pathExists(filepath.Join(dir, customControllerRel)) {
		return true, nil
	}
	return false, nil
}

// readGoModulePath extracts the module path from the first module directive in
// a go.mod file. The directive may use arbitrary whitespace, inline and
// line comments, and quoted paths.
func readGoModulePath(goModPath string) (string, error) {
	f, err := os.Open(goModPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields, err := goModLineFields(scanner.Text())
		if err != nil {
			return "", fmt.Errorf("parse %s: %w", goModPath, err)
		}
		if len(fields) == 0 || fields[0] != "module" {
			continue
		}
		if len(fields) != 2 {
			return "", fmt.Errorf("parse %s: module directive requires exactly one path", goModPath)
		}
		return fields[1], nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", errors.New("module declaration not found in go.mod")
}

// goModLineFields splits one go.mod line into tokens, honouring quoted strings
// and stripping // comments that are outside quotes.
func goModLineFields(line string) ([]string, error) {
	var fields []string
	for i := 0; i < len(line); {
		switch {
		case line[i] == ' ' || line[i] == '\t' || line[i] == '\r':
			i++
		case line[i] == '/' && i+1 < len(line) && line[i+1] == '/':
			return fields, nil
		case line[i] == '"' || line[i] == '`':
			quote := line[i]
			start := i
			i++
			closed := false
			for i < len(line) {
				if quote == '"' && line[i] == '\\' {
					i += 2
					continue
				}
				if line[i] == quote {
					i++
					closed = true
					break
				}
				i++
			}
			if !closed {
				return nil, errors.New("unterminated quoted string")
			}
			token, err := strconv.Unquote(line[start:i])
			if err != nil {
				return nil, fmt.Errorf("invalid quoted string %s: %w", line[start:i], err)
			}
			fields = append(fields, token)
		default:
			start := i
			for i < len(line) && line[i] != ' ' && line[i] != '\t' && line[i] != '\r' {
				i++
			}
			fields = append(fields, line[start:i])
		}
	}
	return fields, nil
}

func resolveFromRepoRoot(repoRoot string, maybeRel string) string {
	if filepath.IsAbs(maybeRel) {
		return filepath.Clean(maybeRel)
	}
	return filepath.Clean(filepath.Join(repoRoot, maybeRel))
}

func resolveNativeFiles(repoRoot string) map[string][]string {
	out := make(map[string][]string, len(nativeFilesByModule))
	for module, files := range nativeFilesByModule {
		resolved := make([]string, 0, len(files))
		for _, file := range files {
			resolved = append(resolved, resolveFromRepoRoot(repoRoot, file))
		}
		out[module] = resolved
	}
	return out
}

func resolveHeaderDir(repoRoot string, headerDir string) string {
	trimmed := strings.TrimSpace(headerDir)
	if trimmed == "" {
		trimmed = defaultHeaderDirRel
	}
	return resolveFromRepoRoot(repoRoot, trimmed)
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
