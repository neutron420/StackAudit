package main

import (
	"context"
	"fmt"
	"os"

	"stack/internal/output"
	"stack/internal/redis"
	"stack/internal/scanner"

	"github.com/spf13/cobra"
)

var redisCmd = &cobra.Command{
	Use:   "redis",
	Short: "Scan Redis configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRedisScan(cmd.Context())
	},
}

func runRedisScan(ctx context.Context) error {
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

	report, err := scanner.Run(ctx, cfg.RootPath, ruleSet, []scanner.Module{redis.NewScanner()}, options)
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
	fmt.Fprintln(os.Stdout, formatted)
	return applyExitCode(report)
}
