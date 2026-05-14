package main

import (
	"context"
	"fmt"
	"os"

	"devdoctor/internal/env"
	"devdoctor/internal/output"
	"devdoctor/internal/scanner"

	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Scan environment files",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runEnvScan(cmd.Context())
	},
}

func runEnvScan(ctx context.Context) error {
	mode, err := output.ParseMode(cfg.OutputMode)
	if err != nil {
		return err
	}

	ruleSet, err := loadRules()
	if err != nil {
		return err
	}

	modules := []scanner.Module{env.NewScanner()}
	report, err := scanner.Run(ctx, cfg.RootPath, ruleSet, modules, scanner.Options{ModuleTimeout: cfg.ModuleTimeout})
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
