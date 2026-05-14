package main

import (
	"fmt"

	"stack/internal/cicd"
	"stack/internal/custom"
	"stack/internal/docker"
	"stack/internal/env"
	"stack/internal/kubernetes"
	"stack/internal/output"
	"stack/internal/postgres"
	"stack/internal/redis"
	"stack/internal/scanner"
	"stack/internal/secrets"

	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [module...]",
	Short: "Run a project health scan",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			cfg.Modules = args
		}
		return runScan(cmd)
	},
}

func runScan(cmd *cobra.Command) error {
	ctx := cmd.Context()
	mode, err := output.ParseMode(cfg.OutputMode)
	if err != nil {
		return err
	}

	ruleSet, err := loadRules()
	if err != nil {
		return err
	}
	options, err := scanOptions()
	if err != nil {
		return err
	}

	availableModules := map[string]scanner.Module{
		"env":        env.NewScanner(),
		"secrets":    secrets.NewScanner(),
		"docker":     docker.NewScanner(),
		"cicd":       cicd.NewScanner(),
		"kubernetes": kubernetes.NewScanner(),
		"redis":      redis.NewScanner(),
		"postgres":   postgres.NewScanner(),
	}

	selectedNames := cfg.Modules
	if len(selectedNames) == 0 && mode == output.ModeTable && !cfg.NoTUI {
		names := []string{}
		for name := range availableModules {
			names = append(names, name)
		}
		selectedNames, err = output.MultiSelect(names)
		if err != nil {
			if err.Error() == "canceled" {
				return nil
			}
			return err
		}
	}

	runner := func() (scanner.Report, error) {
		modules := []scanner.Module{}
		if len(selectedNames) > 0 {
			for _, name := range selectedNames {
				if m, ok := availableModules[name]; ok {
					modules = append(modules, m)
				}
			}
		} else {
			for _, m := range availableModules {
				modules = append(modules, m)
			}
		}

		customModules, err := custom.NewScanners(cfg.RootPath, cfg.PluginPaths)
		if err != nil {
			return scanner.Report{}, err
		}
		modules = append(modules, customModules...)
		return scanner.Run(ctx, cfg.RootPath, ruleSet, modules, options)
	}

	var report scanner.Report
	if mode == output.ModeTable && !cfg.NoTUI {
		report, err = output.RunScanSpinner("Running health scan", runner)
	} else {
		report, err = runner()
	}
	if err != nil {
		return err
	}

	report, err = applyBaseline(report)
	if err != nil {
		return err
	}

	formatted, err := output.Render(report, mode, !cfg.NoTUI)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), formatted)
	return nil
}
