package main

import (
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Alias for scan",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScan(cmd.Context())
	},
}
