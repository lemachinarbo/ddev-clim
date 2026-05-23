package cmd

import (
	"fmt"

	"ddev-clim/config"
	"ddev-clim/ddev"

	"github.com/spf13/cobra"
)

var autostartCmd = &cobra.Command{
	Use:   "autostart",
	Short: "Automatically start previously running DDEV projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		projects, err := ddev.GetProjects()
		if err != nil {
			return fmt.Errorf("failed to get projects: %w", err)
		}

		projectMap := make(map[string]ddev.Project)
		for _, p := range projects {
			projectMap[p.Name] = p
		}

		fmt.Println("Starting saved DDEV session...")
		for _, name := range cfg.RunningProjects {
			p, ok := projectMap[name]
			if !ok {
				fmt.Printf("Skipping %s: project not found in ddev list\n", name)
				continue
			}

			if p.Status == "running" {
				fmt.Printf("%s is already running\n", name)
				continue
			}

			fmt.Printf("Starting %s...\n", name)
			if err := ddev.StartProject(p.AppRoot); err != nil {
				fmt.Printf("Error starting %s: %v\n", name, err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(autostartCmd)
}
