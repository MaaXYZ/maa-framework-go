// Package checker implements the consistency checks behind the api-check
// command: it compares the repository's Go bindings with the C headers under
// deps/include across native API, CustomController, controller method,
// constant, and event coverage, and adds pipeline v2 type and field-name
// coverage when a schema is configured.
package checker

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Run executes the API consistency checker using the process arguments and
// standard streams. It is a compatibility wrapper around run.
func Run() int {
	return run(os.Args[1:], os.Stdout, os.Stderr)
}

// run executes the checker with explicit arguments and output streams so it
// can be tested without touching global flag or stream state. Usage,
// configuration, and path errors return 2, as does a check that fails to
// run; help returns 0, reported findings return 1, and a clean run returns 0.
func run(args []string, stdout io.Writer, stderr io.Writer) int {
	var (
		configPath         string
		headerDirFlag      string
		pipelineSchemaFlag string
		cliBlacklist       stringSliceFlag
	)

	flags := flag.NewFlagSet("api-check", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&configPath, "config", "", "Path to YAML config file")
	flags.StringVar(&headerDirFlag, "header-dir", "", "Directory of C headers")
	flags.StringVar(&pipelineSchemaFlag, "pipeline-schema", "", "Release pipeline JSON schema (enables v2 coverage)")
	flags.Var(&cliBlacklist, "blacklist", "Function name blacklist (repeatable)")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "Usage: api-check [flags]")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Flags:")
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintf(stderr, "unexpected positional argument(s): %s\n", strings.Join(flags.Args(), " "))
		flags.Usage()
		return 2
	}

	repoRoot, err := detectRepoRoot()
	if err != nil {
		fmt.Fprintf(stderr, "failed to detect repository root: %v\n", err)
		return 2
	}

	cfg, loadedConfigPath, err := resolveConfig(configPath, repoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 2
	}

	if strings.TrimSpace(headerDirFlag) != "" {
		cfg.HeaderDir = strings.TrimSpace(headerDirFlag)
	}
	if strings.TrimSpace(pipelineSchemaFlag) != "" {
		cfg.PipelineSchema = strings.TrimSpace(pipelineSchemaFlag)
	}
	resolvedHeaderDir := resolveHeaderDir(repoRoot, cfg.HeaderDir)
	controllerHeaderPath := filepath.Join(resolvedHeaderDir, controllerHeaderRel)
	maaDefHeaderPath := filepath.Join(resolvedHeaderDir, maaDefHeaderRel)
	customControllerPath := resolveFromRepoRoot(repoRoot, customControllerRel)
	adbControllerPath := resolveFromRepoRoot(repoRoot, adbControllerRel)
	win32ControllerPath := resolveFromRepoRoot(repoRoot, win32ControllerRel)
	resolvedNativeFiles, err := discoverNativeFiles(repoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "failed to discover native sources: %v\n", err)
		return 2
	}
	if err := validateRequiredPaths(
		repoRoot,
		resolvedHeaderDir,
		controllerHeaderPath,
		maaDefHeaderPath,
		customControllerPath,
		adbControllerPath,
		win32ControllerPath,
		resolvedNativeFiles,
	); err != nil {
		fmt.Fprintf(stderr, "failed to resolve required input paths: %v\n", err)
		return 2
	}

	if err := checkNativeLibraryEntries(repoRoot); err != nil {
		fmt.Fprintf(stderr, "failed to check native Library entries: %v\n", err)
		return 2
	}
	blacklistSet := mergeBlacklist(cfg.Blacklist, cliBlacklist)
	exclusions := []string{}
	for _, name := range sortedPipelineKeys(cfg.NativeExclusions) {
		reason := strings.TrimSpace(cfg.NativeExclusions[name])
		if strings.TrimSpace(name) != name || name == "" || reason == "" {
			fmt.Fprintln(stderr, "native_exclusions requires exact nonempty symbol names and a nonempty reason")
			return 2
		}
		blacklistSet[name] = struct{}{}
		exclusions = append(exclusions, fmt.Sprintf("native_exclusion: %s: %s", name, reason))
	}

	report := []string{
		fmt.Sprintf("repo_root: %s", filepath.Clean(repoRoot)),
		fmt.Sprintf("header_dir: %s", filepath.Clean(resolvedHeaderDir)),
		fmt.Sprintf("blacklist_size: %d", len(blacklistSet)),
	}
	report = append(report, exclusions...)
	if loadedConfigPath != "" {
		report = append([]string{fmt.Sprintf("config: %s", filepath.Clean(loadedConfigPath))}, report...)
	} else {
		report = append([]string{"config: <none> (using defaults)"}, report...)
	}

	var inventory nativeInventory
	nativeIssues, err := checkNativeAPICoverage(resolvedHeaderDir, resolvedNativeFiles, blacklistSet, &inventory)
	if err != nil {
		fmt.Fprintf(stderr, "failed to check native API coverage: %v\n", err)
		return 2
	}

	for _, module := range moduleOrder {
		if inventory.exported[module] == 0 || inventory.registered[module] == 0 {
			fmt.Fprintf(stderr, "empty native inventory for %s (C=%d, Go=%d)\n", module, inventory.exported[module], inventory.registered[module])
			return 2
		}
		report = append(report, fmt.Sprintf("native_inventory[%s]: C=%d Go=%d", module, inventory.exported[module], inventory.registered[module]))
	}
	report = append(report, fmt.Sprintf("constant_families: %d", len(constantFamilySpecs)))

	controllerIssues, err := checkCustomControllerConsistency(controllerHeaderPath, customControllerPath)
	if err != nil {
		fmt.Fprintf(stderr, "failed to check CustomController consistency: %v\n", err)
		return 2
	}

	methodIssues, err := checkControllerMethodCoverage(maaDefHeaderPath, adbControllerPath, win32ControllerPath)
	if err != nil {
		fmt.Fprintf(stderr, "failed to check controller method coverage: %v\n", err)
		return 2
	}

	callbackIssues, err := checkCallbackABICoverage(repoRoot, resolvedHeaderDir, resolvedNativeFiles)
	if err != nil {
		fmt.Fprintf(stderr, "failed to check callback ABI coverage: %v\n", err)
		return 2
	}
	constantIssues, err := checkConstantCoverage(repoRoot, resolvedHeaderDir)
	if err != nil {
		fmt.Fprintf(stderr, "failed to check constant coverage: %v\n", err)
		return 2
	}
	eventIssues, err := checkEventCoverage(repoRoot, resolvedHeaderDir)
	if err != nil {
		fmt.Fprintf(stderr, "failed to check event coverage: %v\n", err)
		return 2
	}
	issues := make([]issue, 0, len(nativeIssues)+len(controllerIssues)+len(methodIssues)+len(callbackIssues)+len(constantIssues)+len(eventIssues))
	issues = append(issues, nativeIssues...)
	issues = append(issues, controllerIssues...)
	issues = append(issues, methodIssues...)
	issues = append(issues, callbackIssues...)
	issues = append(issues, constantIssues...)
	issues = append(issues, eventIssues...)
	if cfg.PipelineSchema != "" {
		path := resolveFromRepoRoot(repoRoot, cfg.PipelineSchema)
		pipelineIssues, exclusions, err := checkPipelineCoverage(repoRoot, path, cfg.PipelineExclusions)
		if err != nil {
			fmt.Fprintf(stderr, "failed to check pipeline v2 coverage: %v\n", err)
			return 2
		}
		report = append(report, "pipeline_schema: "+path, "pipeline_scope: v2 types and field names", fmt.Sprintf("pipeline_exclusions: %d", len(exclusions)))
		report = append(report, exclusions...)
		issues = append(issues, pipelineIssues...)
	} else {
		if len(cfg.PipelineExclusions) != 0 {
			fmt.Fprintln(stderr, "pipeline_exclusions requires pipeline_schema")
			return 2
		}
		report = append(report, "pipeline_schema: <disabled>")
	}

	return writeReport(stdout, report, issues)
}

func validateRequiredPaths(
	repoRoot string,
	headerDir string,
	controllerHeaderPath string,
	maaDefHeaderPath string,
	customControllerPath string,
	adbControllerPath string,
	win32ControllerPath string,
	nativeFiles map[string][]string,
) error {
	if err := requireDir(headerDir); err != nil {
		return fmt.Errorf("repo_root=%q header-dir %q: %w", filepath.Clean(repoRoot), filepath.Clean(headerDir), err)
	}
	if err := requireFile(controllerHeaderPath); err != nil {
		return fmt.Errorf("repo_root=%q controller header %q: %w", filepath.Clean(repoRoot), filepath.Clean(controllerHeaderPath), err)
	}
	if err := requireFile(maaDefHeaderPath); err != nil {
		return fmt.Errorf("repo_root=%q MaaDef header %q: %w", filepath.Clean(repoRoot), filepath.Clean(maaDefHeaderPath), err)
	}
	if err := requireFile(customControllerPath); err != nil {
		return fmt.Errorf("repo_root=%q custom controller source %q: %w", filepath.Clean(repoRoot), filepath.Clean(customControllerPath), err)
	}
	if err := requireFile(adbControllerPath); err != nil {
		return fmt.Errorf("repo_root=%q adb controller source %q: %w", filepath.Clean(repoRoot), filepath.Clean(adbControllerPath), err)
	}
	if err := requireFile(win32ControllerPath); err != nil {
		return fmt.Errorf("repo_root=%q win32 controller source %q: %w", filepath.Clean(repoRoot), filepath.Clean(win32ControllerPath), err)
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
