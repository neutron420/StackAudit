package main

import (
	"context"
	"fmt"
	"os"

	"devdoctor/internal/cicd"
	"devdoctor/internal/custom"
	"devdoctor/internal/docker"
	"devdoctor/internal/env"
	"devdoctor/internal/kubernetes"
	"devdoctor/internal/output"
	"devdoctor/internal/postgres"
	"devdoctor/internal/redis"
	"devdoctor/internal/scanner"
	"devdoctor/internal/secrets"

	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Run a full project scan",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScan(cmd.Context())
	},
}

func runScan(ctx context.Context) error {
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

	runner := func() (scanner.Report, error) {
		modules := []scanner.Module{
			env.NewScanner(),
			secrets.NewScanner(),
			docker.NewScanner(),
			cicd.NewScanner(),
			kubernetes.NewScanner(),
			redis.NewScanner(),
			postgres.NewScanner(),
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

	formatted, err := output.Render(report, mode)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, formatted)
	if err := applyExitCode(report); err != nil {
		return err
	}
	return nil
}
