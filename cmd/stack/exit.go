package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var exitCmd = &cobra.Command{
	Use:   "exit",
	Short: "Exit the stack sandbox",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Goodbye!")
	},
}
