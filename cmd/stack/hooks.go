package main

import (
	"fmt"
	"strings"

	"stack/internal/githooks"

	"github.com/spf13/cobra"
)

var hookConfig struct {
	Hooks   []string
	Command string
}

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Manage stack git hooks",
}

var hooksInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install stack pre-commit and pre-push hooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		paths, err := githooks.Install(cfg.RootPath, normalizedHooks(), githooks.Options{Command: hookConfig.Command})
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Installed stack hook block in %s\n", strings.Join(paths, ", "))
		return nil
	},
}

var hooksUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove stack-managed hook blocks",
	RunE: func(cmd *cobra.Command, args []string) error {
		paths, err := githooks.Uninstall(cfg.RootPath, normalizedHooks())
		if err != nil {
			return err
		}
		if len(paths) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No stack hook blocks found")
			return nil
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Removed stack hook block from %s\n", strings.Join(paths, ", "))
		return nil
	},
}

func init() {
	hooksInstallCmd.Flags().StringSliceVar(&hookConfig.Hooks, "hook", nil, "Hook to manage: pre-commit or pre-push (repeatable)")
	hooksInstallCmd.Flags().StringVar(&hookConfig.Command, "command", "", "Command to run from the git hook")
	hooksUninstallCmd.Flags().StringSliceVar(&hookConfig.Hooks, "hook", nil, "Hook to manage: pre-commit or pre-push (repeatable)")
	hooksCmd.AddCommand(hooksInstallCmd)
	hooksCmd.AddCommand(hooksUninstallCmd)
}

func normalizedHooks() []string {
	result := []string{}
	for _, value := range hookConfig.Hooks {
		for _, part := range strings.Split(value, ",") {
			hook := strings.TrimSpace(part)
			if hook != "" {
				result = append(result, hook)
			}
		}
	}
	return result
}
