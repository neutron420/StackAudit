package main

import (
	"context"
	"fmt"
	"os"

	"stackaudit/internal/output"
	"stackaudit/internal/postgres"
	"stackaudit/internal/scanner"

	"github.com/spf13/cobra"
)

var postgresCmd = &cobra.Command{
	Use:     "postgres",
	Aliases: []string{"postgresql"},
	Short:   "Scan PostgreSQL configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPostgresScan(cmd.Context())
	},
}

func runPostgresScan(ctx context.Context) error {
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

	report, err := scanner.Run(ctx, cfg.RootPath, ruleSet, []scanner.Module{postgres.NewScanner()}, options)
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
	return applyExitCode(report)
}
