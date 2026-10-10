package checker

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

// loadConfigFromPath reads and strictly decodes a YAML config file. Unknown
// fields, mistyped values, duplicate keys, and extra YAML documents are
// rejected; an empty file decodes to a zero Config.
func loadConfigFromPath(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	cfg := Config{}
	if err := decoder.Decode(&cfg); err != nil {
		if errors.Is(err, io.EOF) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("parse yaml %s: %w", path, err)
	}

	var extra any
	switch err := decoder.Decode(&extra); {
	case err == nil:
		return Config{}, fmt.Errorf("parse yaml %s: multiple YAML documents are not supported", path)
	case !errors.Is(err, io.EOF):
		return Config{}, fmt.Errorf("parse yaml %s: %w", path, err)
	}

	// yaml.v3 coerces boolean and numeric scalars into Go strings. Inspect the
	// source tags as well so mistyped paths and exclusion reasons cannot silently
	// change their meaning during decoding.
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return Config{}, fmt.Errorf("parse yaml %s: %w", path, err)
	}
	if err := validateConfigStringValues(&document); err != nil {
		return Config{}, fmt.Errorf("parse yaml %s: %w", path, err)
	}

	return cfg, nil
}

func validateConfigStringValues(document *yaml.Node) error {
	if len(document.Content) == 0 {
		return nil
	}
	root := document.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil // The strict decoder has already checked the root shape.
	}
	return validateConfigMappingStrings(root)
}

func validateConfigMappingStrings(root *yaml.Node) error {
	root = resolveYAMLNode(root)
	for i := 0; i < len(root.Content); i += 2 {
		key, value := root.Content[i], root.Content[i+1]
		switch key.Value {
		case "<<":
			value = resolveYAMLNode(value)
			if value.Kind == yaml.SequenceNode {
				for _, mapping := range value.Content {
					if err := validateConfigMappingStrings(mapping); err != nil {
						return err
					}
				}
			} else if err := validateConfigMappingStrings(value); err != nil {
				return err
			}
		case "header_dir", "pipeline_schema":
			if err := requireYAMLString(value, key.Value); err != nil {
				return err
			}
		case "blacklist":
			value = resolveYAMLNode(value)
			for j, name := range value.Content {
				if err := requireYAMLString(name, fmt.Sprintf("blacklist[%d]", j)); err != nil {
					return err
				}
			}
		case "pipeline_exclusions", "native_exclusions", "constant_exclusions":
			value = resolveYAMLNode(value)
			for j := 0; j < len(value.Content); j += 2 {
				name, reason := value.Content[j], value.Content[j+1]
				if err := requireYAMLString(name, key.Value+" key"); err != nil {
					return err
				}
				if err := requireYAMLString(reason, key.Value+"."+name.Value); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func resolveYAMLNode(node *yaml.Node) *yaml.Node {
	if node.Kind == yaml.AliasNode {
		return node.Alias
	}
	return node
}

func requireYAMLString(node *yaml.Node, name string) error {
	node = resolveYAMLNode(node)
	if node.Kind != yaml.ScalarNode || node.Tag != "!!str" {
		return fmt.Errorf("%s at line %d must be a string", name, node.Line)
	}
	return nil
}

func mergeConfig(dst *Config, src Config) {
	if strings.TrimSpace(src.HeaderDir) != "" {
		dst.HeaderDir = strings.TrimSpace(src.HeaderDir)
	}
	if len(src.Blacklist) > 0 {
		dst.Blacklist = append(dst.Blacklist, src.Blacklist...)
	}
	if strings.TrimSpace(src.PipelineSchema) != "" {
		dst.PipelineSchema = strings.TrimSpace(src.PipelineSchema)
	}
	if src.NativeExclusions != nil {
		dst.NativeExclusions = src.NativeExclusions
	}
	if src.PipelineExclusions != nil {
		dst.PipelineExclusions = src.PipelineExclusions
	}
	if src.ConstantExclusions != nil {
		dst.ConstantExclusions = src.ConstantExclusions
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
