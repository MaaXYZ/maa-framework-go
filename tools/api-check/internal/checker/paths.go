package checker

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func detectRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get cwd: %w", err)
	}

	dir, err := filepath.Abs(cwd)
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

func readGoModulePath(goModPath string) (string, error) {
	f, err := os.Open(goModPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", errors.New("module declaration not found in go.mod")
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
