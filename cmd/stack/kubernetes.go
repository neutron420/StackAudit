package main

import (
	"context"
	"fmt"
	"os"

	"stack/internal/kubernetes"
	"stack/internal/output"
	"stack/internal/scanner"

	"github.com/spf13/cobra"
)

var kubernetesCmd = &cobra.Command{
	Use:     "kubernetes",
	Aliases: []string{"k8s"},
	Short:   "Scan Kubernetes manifests",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runKubernetesScan(cmd.Context())
	},
}

func runKubernetesScan(ctx context.Context) error {
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

	report, err := scanner.Run(ctx, cfg.RootPath, ruleSet, []scanner.Module{kubernetes.NewScanner()}, options)
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
