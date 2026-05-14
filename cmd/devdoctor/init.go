package main

import (
	"fmt"
	"strings"

	"devdoctor/internal/githooks"
	"devdoctor/internal/projectinit"

	"github.com/spf13/cobra"
)

var initConfig struct {
	Force    bool
	Hooks    bool
	Baseline bool
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create DevDoctor starter configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		written, err := projectinit.WriteStarter(cfg.RootPath, projectinit.Options{Force: initConfig.Force})
		if err != nil {
			return err
		}
		if initConfig.Baseline {
			paths, err := projectinit.WriteEmptyBaseline(cfg.RootPath, projectinit.Options{Force: initConfig.Force})
			if err != nil {
				return err
			}
			written = append(written, paths...)
		}
		if initConfig.Hooks {
			paths, err := githooks.Install(cfg.RootPath, nil, githooks.Options{})
			if err != nil {
				return err
			}
			written = append(written, paths...)
		}
		printInitResult(cmd, written)
		return nil
	},
}

var initGitHubActionsCmd = &cobra.Command{
	Use:   "github-actions",
	Short: "Create a GitHub Actions workflow for DevDoctor",
	RunE: func(cmd *cobra.Command, args []string) error {
		written, err := projectinit.WriteGitHubActions(cfg.RootPath, projectinit.Options{Force: initConfig.Force})
		if err != nil {
			return err
		}
		printInitResult(cmd, written)
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initConfig.Force, "force", false, "Overwrite existing starter files")
	initCmd.Flags().BoolVar(&initConfig.Hooks, "hooks", false, "Install DevDoctor pre-commit and pre-push hooks")
	initCmd.Flags().BoolVar(&initConfig.Baseline, "baseline", false, "Create an empty baseline file")
	initGitHubActionsCmd.Flags().BoolVar(&initConfig.Force, "force", false, "Overwrite an existing GitHub Actions workflow")
	initCmd.AddCommand(initGitHubActionsCmd)
}

func printInitResult(cmd *cobra.Command, paths []string) {
	if len(paths) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "DevDoctor starter files already exist")
		return
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s\n", strings.Join(paths, ", "))
}
