package main

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type fileConfig struct {
	RootPath       string   `yaml:"root"`
	RulesPath      string   `yaml:"rules"`
	RulePacks      []string `yaml:"rule_packs"`
	OutputMode     string   `yaml:"output"`
	NoTUI          *bool    `yaml:"no_tui"`
	ExitCode       *bool    `yaml:"exit_code"`
	MinSeverity    string   `yaml:"min_severity"`
	BaselinePath   string   `yaml:"baseline"`
	UpdateBaseline *bool    `yaml:"update_baseline"`
	ModuleTimeouts []string `yaml:"module_timeouts"`
	PluginPaths    []string `yaml:"plugins"`
}

func applyConfigFile(cmd *cobra.Command) error {
	if cmd.CommandPath() == rootCmd.CommandPath() || strings.HasPrefix(cmd.CommandPath(), rootCmd.CommandPath()+" init") {
		return nil
	}
	path := cfg.ConfigPath
	if path == "" {
		path = ".StackAudit.yaml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	parsed := fileConfig{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return err
	}

	flags := cmd.Flags()
	persistent := cmd.Root().PersistentFlags()
	changed := func(name string) bool {
		if flag := flags.Lookup(name); flag != nil && flag.Changed {
			return true
		}
		if flag := persistent.Lookup(name); flag != nil && flag.Changed {
			return true
		}
		return false
	}

	if parsed.RootPath != "" && !changed("path") {
		cfg.RootPath = parsed.RootPath
	}
	if parsed.RulesPath != "" && !changed("rules") {
		cfg.RulesPath = parsed.RulesPath
	}
	if len(parsed.RulePacks) > 0 && !changed("rule-pack") {
		cfg.RulePacks = parsed.RulePacks
	}
	if parsed.OutputMode != "" && !changed("output") {
		cfg.OutputMode = parsed.OutputMode
	}
	if parsed.NoTUI != nil && !changed("no-tui") {
		cfg.NoTUI = *parsed.NoTUI
	}
	if parsed.ExitCode != nil && !changed("exit-code") {
		cfg.ExitCode = *parsed.ExitCode
	}
	if parsed.MinSeverity != "" && !changed("min-severity") {
		cfg.MinSeverity = parsed.MinSeverity
	}
	if parsed.BaselinePath != "" && !changed("baseline") {
		cfg.BaselinePath = parsed.BaselinePath
	}
	if parsed.UpdateBaseline != nil && !changed("update-baseline") {
		cfg.UpdateBaseline = *parsed.UpdateBaseline
	}
	if len(parsed.ModuleTimeouts) > 0 && !changed("module-timeout") {
		cfg.ModuleTimeouts = parsed.ModuleTimeouts
	}
	if len(parsed.PluginPaths) > 0 && !changed("plugin") {
		cfg.PluginPaths = parsed.PluginPaths
	}
	return nil
}
