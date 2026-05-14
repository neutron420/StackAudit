package main

import (
	"fmt"
	"os"
	"strings"

	"devdoctor/internal/output"
	"devdoctor/internal/rules"
	"devdoctor/pkg/version"

	"github.com/spf13/cobra"
)

type appConfig struct {
	RootPath       string
	RulesPath      string
	RulePacks      []string
	OutputMode     string
	NoTUI          bool
	ExitCode       bool
	MinSeverity    string
	BaselinePath   string
	UpdateBaseline bool
	ModuleTimeouts []string
	PluginPaths    []string
}

var cfg appConfig

var rootCmd = &cobra.Command{
	Use:   "devdoctor",
	Short: "DevDoctor scans backend projects for production health issues",
	Long:  "DevDoctor is a local-first backend health scanner for environment, secrets, Docker, CI/CD, Kubernetes, Redis, PostgreSQL, and custom plugin checks.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = version.FullVersion()
	rootCmd.PersistentFlags().StringVarP(&cfg.RootPath, "path", "p", ".", "Project root path")
	rootCmd.PersistentFlags().StringVar(&cfg.RulesPath, "rules", "", "Path to rules YAML file")
	rootCmd.PersistentFlags().StringSliceVar(&cfg.RulePacks, "rule-pack", nil, "Rule pack names or YAML paths (repeatable)")
	rootCmd.PersistentFlags().StringVarP(&cfg.OutputMode, "output", "o", string(output.ModeTable), "Output mode: table|json|markdown|sarif")
	rootCmd.PersistentFlags().BoolVar(&cfg.NoTUI, "no-tui", false, "Disable TUI loading indicators")
	rootCmd.PersistentFlags().BoolVar(&cfg.ExitCode, "exit-code", false, "Return non-zero exit code when findings meet --min-severity")
	rootCmd.PersistentFlags().StringVar(&cfg.MinSeverity, "min-severity", "warning", "Minimum severity for non-zero exit code: critical|warning|info")
	rootCmd.PersistentFlags().StringVar(&cfg.BaselinePath, "baseline", ".devdoctor.baseline.json", "Baseline file path")
	rootCmd.PersistentFlags().BoolVar(&cfg.UpdateBaseline, "update-baseline", false, "Write a baseline file from the current scan")
	rootCmd.PersistentFlags().StringSliceVar(&cfg.ModuleTimeouts, "module-timeout", nil, "Module timeout budget: duration for all modules or module=duration (e.g. 2s, env=500ms,secrets=5s)")
	rootCmd.PersistentFlags().StringSliceVar(&cfg.PluginPaths, "plugin", nil, "Custom scanner plugin YAML path (repeatable)")

	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(envCmd)
	rootCmd.AddCommand(dockerCmd)
	rootCmd.AddCommand(ciCmd)
	rootCmd.AddCommand(secretsCmd)
	rootCmd.AddCommand(kubernetesCmd)
	rootCmd.AddCommand(redisCmd)
	rootCmd.AddCommand(postgresCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(fixCmd)
	rootCmd.AddCommand(hooksCmd)
	rootCmd.AddCommand(versionCmd)
}

func loadRules() (rules.RuleSet, error) {
	packs := normalizePacks(cfg.RulePacks)
	if cfg.RulesPath == "" && len(packs) == 0 {
		return rules.DefaultRuleSet(), nil
	}

	return rules.Load(cfg.RulesPath, packs)
}

func normalizePacks(values []string) []string {
	result := []string{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		parts := strings.Split(trimmed, ",")
		for _, part := range parts {
			name := strings.TrimSpace(part)
			if name != "" {
				result = append(result, name)
			}
		}
	}
	return result
}
