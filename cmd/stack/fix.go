package main

import (
	"context"
	"fmt"
	"os"

	"stack/internal/fix"
	"stack/internal/output"

	"github.com/spf13/cobra"
)

var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Generate fixes with confirmation",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runFix(cmd.Context())
	},
}

func runFix(ctx context.Context) error {
	mode, err := output.ParseMode(cfg.OutputMode)
	if err != nil {
		return err
	}
	if mode != output.ModeTable {
		return fmt.Errorf("fix supports only table output")
	}

	ruleSet, err := loadRules()
	if err != nil {
		return err
	}

	plan, err := fix.BuildPlan(ctx, cfg.RootPath, ruleSet)
	if err != nil {
		return err
	}

	if len(plan.Actions) == 0 {
		fmt.Fprintln(os.Stdout, output.RenderFixEmpty(plan))
		return nil
	}

	fmt.Fprintln(os.Stdout, output.RenderFixPlan(plan))
	confirmed, err := output.Confirm("Apply these fixes now?")
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Fprintln(os.Stdout, "No changes applied.")
		return nil
	}

	results := fix.ApplyPlan(plan)
	fmt.Fprintln(os.Stdout, output.RenderFixResults(plan, results))
	return nil
}
