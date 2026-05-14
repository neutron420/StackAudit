package main

import (
	
	"fmt"


	"stack/internal/fix"
	"stack/internal/output"

	"github.com/spf13/cobra"
)

var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Generate fixes with confirmation",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runFix(cmd)
	},
}

func runFix(cmd *cobra.Command) error {
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

	plan, err := fix.BuildPlan(cmd.Context(), cfg.RootPath, ruleSet)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()

	if len(plan.Actions) == 0 {
		fmt.Fprintln(w, output.RenderFixEmpty(plan))
		return nil
	}

	fmt.Fprintln(w, output.RenderFixPlan(plan))

	if cfg.NoTUI {
		fmt.Fprintln(w, "Run 'stack fix' directly in terminal to apply fixes interactively.")
		return nil
	}

	confirmed, err := output.Confirm("Apply these fixes now?")
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Fprintln(w, "No changes applied.")
		return nil
	}

	results := fix.ApplyPlan(plan)
	fmt.Fprintln(w, output.RenderFixResults(plan, results))
	return nil
}
