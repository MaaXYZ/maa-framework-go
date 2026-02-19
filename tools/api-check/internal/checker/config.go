package checker

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func resolveConfig(configPath string, repoRoot string) (Config, string, error) {
	cfg := Config{
		HeaderDir: "",
		Blacklist: []string{},
	}

	if strings.TrimSpace(configPath) != "" {
		path := strings.TrimSpace(configPath)
		loaded, err := loadConfigFromPath(path)
		if err != nil {
			return cfg, "", err
		}
		mergeConfig(&cfg, loaded)
		return cfg, path, nil
	}

	if _, err := os.Stat(autoConfigFileName); err == nil {
		loaded, err := loadConfigFromPath(autoConfigFileName)
		if err != nil {
			return cfg, "", err
		}
		mergeConfig(&cfg, loaded)
		return cfg, autoConfigFileName, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return cfg, "", fmt.Errorf("check auto config %s: %w", autoConfigFileName, err)
	}

	fallbackPath := resolveFromRepoRoot(repoRoot, apiCheckConfigPathRel)
	if _, err := os.Stat(fallbackPath); err == nil {
		loaded, err := loadConfigFromPath(fallbackPath)
		if err != nil {
			return cfg, "", err
		}
		mergeConfig(&cfg, loaded)
		return cfg, fallbackPath, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return cfg, "", fmt.Errorf("check auto config %s: %w", fallbackPath, err)
	}

	return cfg, "", nil
}

func loadConfigFromPath(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}
	cfg := Config{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse yaml %s: %w", path, err)
	}
	return cfg, nil
}

func mergeConfig(dst *Config, src Config) {
	if strings.TrimSpace(src.HeaderDir) != "" {
		dst.HeaderDir = strings.TrimSpace(src.HeaderDir)
	}
	if len(src.Blacklist) > 0 {
		dst.Blacklist = append(dst.Blacklist, src.Blacklist...)
	}
}

func mergeBlacklist(configBlacklist []string, cliBlacklist []string) map[string]struct{} {
	blacklistSet := make(map[string]struct{})
	for _, n := range configBlacklist {
		if name := strings.TrimSpace(n); name != "" {
			blacklistSet[name] = struct{}{}
		}
	}
	for _, n := range cliBlacklist {
		if name := strings.TrimSpace(n); name != "" {
			blacklistSet[name] = struct{}{}
		}
	}
	return blacklistSet
}
