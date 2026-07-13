package cmd

import (
	"fmt"
	"time"

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

		var projects []ddev.Project
		var errProjects error
		for i := 0; i < 5; i++ {
			projects, errProjects = ddev.GetProjects()
			if errProjects == nil {
				break
			}
			fmt.Printf("Attempt %d: failed to get projects, retrying in 2s... (%v)\n", i+1, errProjects)
			time.Sleep(2 * time.Second)
		}
		if errProjects != nil {
			return fmt.Errorf("failed to get projects after retries: %w", errProjects)
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
			// Run stop first to ensure a clean slate (removes stale containers/sockets)
			_ = ddev.StopProject(p.AppRoot)
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
