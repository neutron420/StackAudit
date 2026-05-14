package main

import (
	"fmt"
	"os"
	"strings"

	"stack/internal/output"
	"stack/internal/rules"
	"stack/pkg/version"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type appConfig struct {
	RootPath       string
	RulesPath      string
	ConfigPath     string
	RulePacks      []string
	OutputMode     string
	NoTUI          bool
	ExitCode       bool
	MinSeverity    string
	BaselinePath   string
	UpdateBaseline bool
	ModuleTimeouts []string
	PluginPaths    []string
	Modules        []string
}

var cfg appConfig

var rootCmd = &cobra.Command{
	Use:   "stack",
	Short: "Stack scans backend projects for production health issues",
	Long:  "Stack is a local-first backend health scanner for environment, secrets, Docker, CI/CD, Kubernetes, Redis, PostgreSQL, and custom plugin checks.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return output.RunSandbox(func(shellArgs []string) string {
				if len(shellArgs) > 0 && (shellArgs[0] == "stack" || shellArgs[0] == "./stack" || shellArgs[0] == "stack.exe") {
					shellArgs = shellArgs[1:]
				}
				if len(shellArgs) == 0 {
					return ""
				}

				if len(shellArgs) == 1 && (shellArgs[0] == "cls" || shellArgs[0] == "clear") {
					return "__CLEAR__"
				}

				subCmd, subArgs, err := cmd.Find(shellArgs)
				if err != nil {
					return fmt.Sprintf("Error: %v\n", err)
				}
				
				if subCmd.Name() == "stack" && len(shellArgs) > 0 {
					return fmt.Sprintf("Error: unknown command %q\n", shellArgs[0])
				}
				
				// Capture output
				var buf strings.Builder
				subCmd.SetOut(&buf)
				subCmd.SetErr(&buf)
				
				// Set context, reset flags, and DISABLE TUI for sub-execution
				// (Running a TUI inside a TUI causes memory corruption)
				subCmd.SetContext(cmd.Context())
				
				// Reset all flags to their default state for a clean execution
				subCmd.Flags().VisitAll(func(f *pflag.Flag) {
					if slice, ok := f.Value.(pflag.SliceValue); ok {
						slice.Replace(nil)
					} else {
						f.Value.Set(f.DefValue)
					}
					f.Changed = false
				})

				// Re-load config to ensure fresh state for every command run in the workbench
				if err := applyConfigFile(subCmd); err != nil {
					return fmt.Sprintf("Error loading config: %v\n", err)
				}

				if err := subCmd.Flags().Parse(subArgs); err != nil {
					return fmt.Sprintf("Error parsing flags: %v\n", err)
				}
				
				oldNoTUI := cfg.NoTUI
				cfg.NoTUI = true
				defer func() { cfg.NoTUI = oldNoTUI }()
				
				var cmdErr error
				if subCmd.RunE != nil {
					cmdErr = subCmd.RunE(subCmd, subArgs)
				} else if subCmd.Run != nil {
					subCmd.Run(subCmd, subArgs)
				} else {
					cmdErr = subCmd.Help()
				}

				if cmdErr != nil {
					return fmt.Sprintf("Error: %v\n", cmdErr)
				}
				
				return buf.String()
			})
		}
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
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		err := applyConfigFile(cmd)
		if err != nil {
			return err
		}

		// Only print banner for human-readable table output and if TUI is enabled
		if cfg.OutputMode == string(output.ModeTable) && !cfg.NoTUI && cmd.Name() != "version" {
			output.PrintBanner()
		}
		return nil
	}
	rootCmd.PersistentFlags().StringVarP(&cfg.RootPath, "path", "p", ".", "Project root path")
	rootCmd.PersistentFlags().StringVar(&cfg.ConfigPath, "config", ".stack.yaml", "Path to stack config file")
	rootCmd.PersistentFlags().StringVar(&cfg.RulesPath, "rules", "", "Path to rules YAML file")
	rootCmd.PersistentFlags().StringSliceVar(&cfg.RulePacks, "rule-pack", nil, "Rule pack names or YAML paths (repeatable)")
	rootCmd.PersistentFlags().StringVarP(&cfg.OutputMode, "output", "o", string(output.ModeTable), "Output mode: table|json|markdown|sarif|html")
	rootCmd.PersistentFlags().BoolVar(&cfg.NoTUI, "no-tui", false, "Disable TUI loading indicators")
	rootCmd.PersistentFlags().BoolVar(&cfg.ExitCode, "exit-code", false, "Return non-zero exit code when findings meet --min-severity")
	rootCmd.PersistentFlags().StringVar(&cfg.MinSeverity, "min-severity", "warning", "Minimum severity for non-zero exit code: critical|warning|info")
	rootCmd.PersistentFlags().StringVar(&cfg.BaselinePath, "baseline", ".stack.baseline.json", "Baseline file path")
	rootCmd.PersistentFlags().BoolVar(&cfg.UpdateBaseline, "update-baseline", false, "Write a baseline file from the current scan")
	rootCmd.PersistentFlags().StringSliceVar(&cfg.ModuleTimeouts, "module-timeout", nil, "Module timeout budget: duration for all modules or module=duration (e.g. 2s, env=500ms,secrets=5s)")
	rootCmd.PersistentFlags().StringSliceVar(&cfg.PluginPaths, "plugin", nil, "Custom scanner plugin YAML path (repeatable)")
	rootCmd.PersistentFlags().StringSliceVar(&cfg.Modules, "module", nil, "Specific modules to run (repeatable, e.g. env,docker)")

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
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(hooksCmd)
	rootCmd.AddCommand(exitCmd)
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
