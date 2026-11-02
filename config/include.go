package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

const includeKey = "include"

type exportersSection struct {
	items        map[string]ExporterConfig
	includePaths []string
}

func (s *exportersSection) UnmarshalYAML(value *yaml.Node) error {
	if value == nil || value.Kind == 0 {
		return nil
	}
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("exporters must be a mapping")
	}

	s.items = make(map[string]ExporterConfig)

	for i := 0; i < len(value.Content); i += 2 {
		key := value.Content[i].Value
		node := value.Content[i+1]

		if key == includeKey {
			paths, err := parseIncludePaths(node)
			if err != nil {
				return fmt.Errorf("exporters.include: %w", err)
			}
			s.includePaths = append(s.includePaths, paths...)
			continue
		}

		var ec ExporterConfig
		if err := node.Decode(&ec); err != nil {
			return fmt.Errorf("exporters.%s: %w", key, err)
		}
		s.items[key] = ec
	}

	return nil
}

func (s *exportersSection) Resolve(baseDir string) error {
	if len(s.includePaths) == 0 {
		return nil
	}

	merged := make(map[string]ExporterConfig)

	for _, includePath := range s.includePaths {
		dir := resolveIncludePath(baseDir, includePath)
		fromDir, err := loadExporterConfigsFromDir(dir)
		if err != nil {
			return err
		}
		for name, cfg := range fromDir {
			merged[name] = cfg
		}
	}

	for name, cfg := range s.items {
		merged[name] = cfg
	}

	s.items = merged
	s.includePaths = nil
	return nil
}

func parseIncludePaths(node *yaml.Node) ([]string, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		if node.Value == "" {
			return nil, fmt.Errorf("must not be empty")
		}
		return []string{node.Value}, nil
	case yaml.SequenceNode:
		paths := make([]string, 0, len(node.Content))
		for _, item := range node.Content {
			if item.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("entries must be strings")
			}
			if item.Value == "" {
				return nil, fmt.Errorf("entries must not be empty")
			}
			paths = append(paths, item.Value)
		}
		return paths, nil
	default:
		return nil, fmt.Errorf("must be a string or list of strings")
	}
}

func resolveIncludePath(baseDir, includePath string) string {
	if filepath.IsAbs(includePath) {
		return includePath
	}
	return filepath.Join(baseDir, includePath)
}

func loadExporterConfigsFromDir(dir string) (map[string]ExporterConfig, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("include directory does not exist: %s", dir)
		}
		return nil, fmt.Errorf("reading include directory %s: %w", dir, err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}
	sort.Strings(files)

	result := make(map[string]ExporterConfig)
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("reading include file %s: %w", file, err)
		}

		var fragment map[string]ExporterConfig
		if err := yaml.Unmarshal(data, &fragment); err != nil {
			return nil, fmt.Errorf("parsing include file %s: %w", file, err)
		}
		for name, cfg := range fragment {
			result[name] = cfg
		}
	}

	return result, nil
}
