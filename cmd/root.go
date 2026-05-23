package cmd

import (
	"fmt"
	"os"

	"ddev-clim/config"
	"ddev-clim/ui"

	"github.com/spf13/cobra"
)

var scanPath string

var rootCmd = &cobra.Command{
	Use:   "ddev-clim",
	Short: "A TUI for managing DDEV instances",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if scanPath != "" {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			cfg.ScanPath = scanPath
			return config.SaveConfig(cfg)
		}
		return nil
	},
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

func init() {
	rootCmd.PersistentFlags().StringVarP(&scanPath, "path", "p", "", "Folder to scan for DDEV instances")
}
