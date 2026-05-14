package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestApplyConfigFilePopulatesUnsetConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".devdoctor.yaml")
	if err := os.WriteFile(path, []byte(`
root: ./service
rules: ./rules.yaml
rule_packs:
  - strict
output: html
no_tui: true
exit_code: true
min_severity: critical
baseline: ./baseline.json
module_timeouts:
  - env=1s
plugins:
  - ./plugin.yaml
`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	original := cfg
	t.Cleanup(func() { cfg = original })
	cfg = appConfig{
		ConfigPath:     path,
		OutputMode:     "table",
		MinSeverity:    "warning",
		BaselinePath:   ".devdoctor.baseline.json",
		ModuleTimeouts: nil,
	}

	cmd := &cobra.Command{Use: "scan"}
	if err := applyConfigFile(cmd); err != nil {
		t.Fatalf("applyConfigFile returned error: %v", err)
	}

	if cfg.RootPath != "./service" || cfg.RulesPath != "./rules.yaml" || cfg.OutputMode != "html" {
		t.Fatalf("cfg not populated from file: %+v", cfg)
	}
	if !cfg.NoTUI || !cfg.ExitCode {
		t.Fatalf("bool config not applied: %+v", cfg)
	}
	if len(cfg.RulePacks) != 1 || cfg.RulePacks[0] != "strict" {
		t.Fatalf("rule packs = %v", cfg.RulePacks)
	}
	if len(cfg.PluginPaths) != 1 || cfg.PluginPaths[0] != "./plugin.yaml" {
		t.Fatalf("plugins = %v", cfg.PluginPaths)
	}
}
