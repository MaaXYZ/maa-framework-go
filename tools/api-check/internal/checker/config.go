package checker

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func resolveConfig(configPath string) (Config, string, error) {
	cfg := Config{
		HeaderDir: defaultHeaderDir,
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

	if _, err := os.Stat(autoConfigPath); err == nil {
		loaded, err := loadConfigFromPath(autoConfigPath)
		if err != nil {
			return cfg, "", err
		}
		mergeConfig(&cfg, loaded)
		return cfg, autoConfigPath, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return cfg, "", fmt.Errorf("check auto config %s: %w", autoConfigPath, err)
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
