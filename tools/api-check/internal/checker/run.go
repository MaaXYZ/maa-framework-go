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

	cfg, loadedConfigPath, err := resolveConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return 2
	}

	if strings.TrimSpace(headerDirFlag) != "" {
		cfg.HeaderDir = strings.TrimSpace(headerDirFlag)
	}
	if cfg.HeaderDir == "" {
		cfg.HeaderDir = defaultHeaderDir
	}

	blacklistSet := mergeBlacklist(cfg.Blacklist, cliBlacklist)

	report := []string{
		fmt.Sprintf("header_dir: %s", filepath.Clean(cfg.HeaderDir)),
		fmt.Sprintf("blacklist_size: %d", len(blacklistSet)),
	}
	if loadedConfigPath != "" {
		report = append([]string{fmt.Sprintf("config: %s", filepath.Clean(loadedConfigPath))}, report...)
	} else {
		report = append([]string{"config: <none> (using defaults)"}, report...)
	}

	nativeIssues, err := checkNativeAPICoverage(cfg.HeaderDir, blacklistSet)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to check native API coverage: %v\n", err)
		return 2
	}

	controllerHeaderPath := filepath.Join(cfg.HeaderDir, controllerHeaderRel)
	controllerIssues, err := checkCustomControllerConsistency(controllerHeaderPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to check CustomController consistency: %v\n", err)
		return 2
	}

	issues := make([]issue, 0, len(nativeIssues)+len(controllerIssues))
	issues = append(issues, nativeIssues...)
	issues = append(issues, controllerIssues...)

	return printReport(report, issues)
}
