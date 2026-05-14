package main

import (
	"context"
	"fmt"
	"os"

	"devdoctor/internal/docker"
	"devdoctor/internal/output"
	"devdoctor/internal/scanner"

	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Scan Docker configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDockerScan(cmd.Context())
	},
}

func runDockerScan(ctx context.Context) error {
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

	modules := []scanner.Module{docker.NewScanner()}
	report, err := scanner.Run(ctx, cfg.RootPath, ruleSet, modules, options)
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
