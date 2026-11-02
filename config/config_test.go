package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfigFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("creating directory: %v", err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing config file: %v", err)
	}
	return path
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("creating directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing file: %v", err)
	}
}

func TestLoadConfig_InlineOnly(t *testing.T) {
	dir := t.TempDir()
	path := writeConfigFile(t, dir, "config.yml", `
listen: "127.0.0.1:9099"
exporters:
  node_exporter:
    path: "/node_exporter/metrics"
    url: "http://127.0.0.1:9100/metrics"
`)

	cfg, err := LoadConfigFromYAML(path)
	if err != nil {
		t.Fatalf("LoadConfigFromYAML: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(cfg.ExporterConfigs) != 1 {
		t.Fatalf("expected 1 exporter, got %d", len(cfg.ExporterConfigs))
	}
}

func TestLoadConfig_EmptyIncludeDir(t *testing.T) {
	dir := t.TempDir()
	includeDir := filepath.Join(dir, "exporters.d")
	if err := os.Mkdir(includeDir, 0755); err != nil {
		t.Fatalf("creating include dir: %v", err)
	}

	path := writeConfigFile(t, dir, "config.yml", `
listen: "127.0.0.1:9099"
exporters:
  include: exporters.d
  node_exporter:
    path: "/node_exporter/metrics"
    url: "http://127.0.0.1:9100/metrics"
`)

	cfg, err := LoadConfigFromYAML(path)
	if err != nil {
		t.Fatalf("LoadConfigFromYAML: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestLoadConfig_EmptyIncludeDir_NoExporters(t *testing.T) {
	dir := t.TempDir()
	includeDir := filepath.Join(dir, "exporters.d")
	if err := os.Mkdir(includeDir, 0755); err != nil {
		t.Fatalf("creating include dir: %v", err)
	}

	path := writeConfigFile(t, dir, "config.yml", `
listen: "127.0.0.1:9099"
exporters:
  include: exporters.d
`)

	cfg, err := LoadConfigFromYAML(path)
	if err != nil {
		t.Fatalf("LoadConfigFromYAML: %v", err)
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error when no exporters are defined")
	}
}

func TestLoadConfig_IncludeSingleFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "exporters.d", "mysql.yml"), `
mysqld_exporter:
  path: "/mysqld_exporter/metrics"
  url: "http://127.0.0.1:9104/metrics"
`)

	path := writeConfigFile(t, dir, "config.yml", `
listen: "127.0.0.1:9099"
exporters:
  include: exporters.d
`)

	cfg, err := LoadConfigFromYAML(path)
	if err != nil {
		t.Fatalf("LoadConfigFromYAML: %v", err)
	}
	if _, ok := cfg.ExporterConfigs["mysqld_exporter"]; !ok {
		t.Fatal("expected mysqld_exporter from include file")
	}
}

func TestLoadConfig_IncludeMultipleFiles_LastWins(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "exporters.d", "a.yml"), `
shared_exporter:
  path: "/old/metrics"
  url: "http://127.0.0.1:9100/metrics"
`)
	writeFile(t, filepath.Join(dir, "exporters.d", "b.yml"), `
shared_exporter:
  path: "/new/metrics"
  url: "http://127.0.0.1:9101/metrics"
other_exporter:
  path: "/other/metrics"
  url: "http://127.0.0.1:9102/metrics"
`)

	path := writeConfigFile(t, dir, "config.yml", `
listen: "127.0.0.1:9099"
exporters:
  include: exporters.d
`)

	cfg, err := LoadConfigFromYAML(path)
	if err != nil {
		t.Fatalf("LoadConfigFromYAML: %v", err)
	}

	shared := cfg.ExporterConfigs["shared_exporter"]
	if shared.Path == nil || *shared.Path != "/new/metrics" {
		t.Fatalf("expected later include file to win, got path %v", shared.Path)
	}
	if _, ok := cfg.ExporterConfigs["other_exporter"]; !ok {
		t.Fatal("expected other_exporter from include files")
	}
}

func TestLoadConfig_InlineAndInclude_InlineWins(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "exporters.d", "node.yml"), `
node_exporter:
  path: "/from/include/metrics"
  url: "http://127.0.0.1:9100/metrics"
`)

	path := writeConfigFile(t, dir, "config.yml", `
listen: "127.0.0.1:9099"
exporters:
  include: exporters.d
  node_exporter:
    path: "/from/inline/metrics"
    url: "http://127.0.0.1:9200/metrics"
`)

	cfg, err := LoadConfigFromYAML(path)
	if err != nil {
		t.Fatalf("LoadConfigFromYAML: %v", err)
	}

	node := cfg.ExporterConfigs["node_exporter"]
	if node.Path == nil || *node.Path != "/from/inline/metrics" {
		t.Fatalf("expected inline exporter to win, got path %v", node.Path)
	}
	if node.URL == nil || *node.URL != "http://127.0.0.1:9200/metrics" {
		t.Fatalf("expected inline exporter URL to win, got url %v", node.URL)
	}
}

func TestLoadConfig_IncludeMultipleDirectories_LastWins(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "first.d", "a.yml"), `
shared_exporter:
  path: "/first/metrics"
  url: "http://127.0.0.1:9100/metrics"
`)
	writeFile(t, filepath.Join(dir, "second.d", "a.yml"), `
shared_exporter:
  path: "/second/metrics"
  url: "http://127.0.0.1:9101/metrics"
`)

	path := writeConfigFile(t, dir, "config.yml", `
listen: "127.0.0.1:9099"
exporters:
  include:
    - first.d
    - second.d
`)

	cfg, err := LoadConfigFromYAML(path)
	if err != nil {
		t.Fatalf("LoadConfigFromYAML: %v", err)
	}

	shared := cfg.ExporterConfigs["shared_exporter"]
	if shared.Path == nil || *shared.Path != "/second/metrics" {
		t.Fatalf("expected later include directory to win, got path %v", shared.Path)
	}
}

func TestLoadConfig_InvalidFragment(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "exporters.d", "bad.yml"), `{not yaml`)

	path := writeConfigFile(t, dir, "config.yml", `
listen: "127.0.0.1:9099"
exporters:
  include: exporters.d
`)

	_, err := LoadConfigFromYAML(path)
	if err == nil {
		t.Fatal("expected error for invalid include file")
	}
}

func TestLoadConfig_MissingIncludeDir(t *testing.T) {
	dir := t.TempDir()
	path := writeConfigFile(t, dir, "config.yml", `
listen: "127.0.0.1:9099"
exporters:
  include: missing.d
`)

	_, err := LoadConfigFromYAML(path)
	if err == nil {
		t.Fatal("expected error for missing include directory")
	}
}

func TestLoadConfig_RelativeIncludePath(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "nested", "exporters.d", "node.yml"), `
node_exporter:
  path: "/node_exporter/metrics"
  url: "http://127.0.0.1:9100/metrics"
`)

	path := writeConfigFile(t, filepath.Join(dir, "nested"), "config.yml", `
listen: "127.0.0.1:9099"
exporters:
  include: exporters.d
`)

	cfg, err := LoadConfigFromYAML(path)
	if err != nil {
		t.Fatalf("LoadConfigFromYAML: %v", err)
	}
	if _, ok := cfg.ExporterConfigs["node_exporter"]; !ok {
		t.Fatal("expected node_exporter from relative include path")
	}
}

func TestLoadConfig_InvalidGlobalConfig(t *testing.T) {
	dir := t.TempDir()
	includeDir := filepath.Join(dir, "exporters.d")
	if err := os.Mkdir(includeDir, 0755); err != nil {
		t.Fatalf("creating include dir: %v", err)
	}

	path := writeConfigFile(t, dir, "config.yml", `
exporters:
  include: exporters.d
`)

	cfg, err := LoadConfigFromYAML(path)
	if err != nil {
		t.Fatalf("LoadConfigFromYAML: %v", err)
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for missing listen and exporters")
	}
}

func TestLoadConfig_SkipsNonYAMLAndHiddenFiles(t *testing.T) {
	dir := t.TempDir()
	includeDir := filepath.Join(dir, "exporters.d")
	if err := os.Mkdir(includeDir, 0755); err != nil {
		t.Fatalf("creating include dir: %v", err)
	}
	writeFile(t, filepath.Join(includeDir, ".hidden.yml"), `
hidden_exporter:
  path: "/hidden/metrics"
  url: "http://127.0.0.1:9100/metrics"
`)
	writeFile(t, filepath.Join(includeDir, "notes.txt"), "ignore me")
	writeFile(t, filepath.Join(includeDir, "valid.yml"), `
valid_exporter:
  path: "/valid/metrics"
  url: "http://127.0.0.1:9101/metrics"
`)

	path := writeConfigFile(t, dir, "config.yml", `
listen: "127.0.0.1:9099"
exporters:
  include: exporters.d
`)

	cfg, err := LoadConfigFromYAML(path)
	if err != nil {
		t.Fatalf("LoadConfigFromYAML: %v", err)
	}
	if _, ok := cfg.ExporterConfigs["hidden_exporter"]; ok {
		t.Fatal("did not expect hidden include file to be loaded")
	}
	if _, ok := cfg.ExporterConfigs["valid_exporter"]; !ok {
		t.Fatal("expected valid include file to be loaded")
	}
}
