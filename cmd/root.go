package cmd

import (
	"fmt"
	"os"

	"ddev-clim/ui"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ddev-clim",
	Short: "A TUI for managing DDEV instances",
	RunE: func(cmd *cobra.Command, args []string) error {
		return ui.StartTUI()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
