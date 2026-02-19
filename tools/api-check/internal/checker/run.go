package checker

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Run() int {
	var (
		configPath    string
		headerDirFlag string
		cliBlacklist  stringSliceFlag
	)

	flag.StringVar(&configPath, "config", "", "Path to YAML config file")
	flag.StringVar(&headerDirFlag, "header-dir", "", "Directory of C headers")
	flag.Var(&cliBlacklist, "blacklist", "Function name blacklist (repeatable)")
	flag.Parse()

	repoRoot, err := detectRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to detect repository root: %v\n", err)
		return 2
	}

	cfg, loadedConfigPath, err := resolveConfig(configPath, repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return 2
	}

	if strings.TrimSpace(headerDirFlag) != "" {
		cfg.HeaderDir = strings.TrimSpace(headerDirFlag)
	}
	resolvedHeaderDir := resolveHeaderDir(repoRoot, cfg.HeaderDir)
	controllerHeaderPath := filepath.Join(resolvedHeaderDir, controllerHeaderRel)
	customControllerPath := resolveFromRepoRoot(repoRoot, customControllerRel)
	resolvedNativeFiles := resolveNativeFiles(repoRoot)
	if err := validateRequiredPaths(repoRoot, resolvedHeaderDir, controllerHeaderPath, customControllerPath, resolvedNativeFiles); err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve required input paths: %v\n", err)
		return 2
	}

	blacklistSet := mergeBlacklist(cfg.Blacklist, cliBlacklist)

	report := []string{
		fmt.Sprintf("repo_root: %s", filepath.Clean(repoRoot)),
		fmt.Sprintf("header_dir: %s", filepath.Clean(resolvedHeaderDir)),
		fmt.Sprintf("blacklist_size: %d", len(blacklistSet)),
	}
	if loadedConfigPath != "" {
		report = append([]string{fmt.Sprintf("config: %s", filepath.Clean(loadedConfigPath))}, report...)
	} else {
		report = append([]string{"config: <none> (using defaults)"}, report...)
	}

	nativeIssues, err := checkNativeAPICoverage(resolvedHeaderDir, resolvedNativeFiles, blacklistSet)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to check native API coverage: %v\n", err)
		return 2
	}

	controllerIssues, err := checkCustomControllerConsistency(controllerHeaderPath, customControllerPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to check CustomController consistency: %v\n", err)
		return 2
	}

	issues := make([]issue, 0, len(nativeIssues)+len(controllerIssues))
	issues = append(issues, nativeIssues...)
	issues = append(issues, controllerIssues...)

	return printReport(report, issues)
}

func validateRequiredPaths(
	repoRoot string,
	headerDir string,
	controllerHeaderPath string,
	customControllerPath string,
	nativeFiles map[string][]string,
) error {
	if err := requireDir(headerDir); err != nil {
		return fmt.Errorf("repo_root=%q header-dir %q: %w", filepath.Clean(repoRoot), filepath.Clean(headerDir), err)
	}
	if err := requireFile(controllerHeaderPath); err != nil {
		return fmt.Errorf("repo_root=%q controller header %q: %w", filepath.Clean(repoRoot), filepath.Clean(controllerHeaderPath), err)
	}
	if err := requireFile(customControllerPath); err != nil {
		return fmt.Errorf("repo_root=%q custom controller source %q: %w", filepath.Clean(repoRoot), filepath.Clean(customControllerPath), err)
	}
	for module, files := range nativeFiles {
		for _, file := range files {
			if err := requireFile(file); err != nil {
				return fmt.Errorf("repo_root=%q native source [%s] %q: %w", filepath.Clean(repoRoot), module, filepath.Clean(file), err)
			}
		}
	}
	return nil
}

func requireDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory")
	}
	return nil
}

func requireFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("is a directory")
	}
	return nil
}
