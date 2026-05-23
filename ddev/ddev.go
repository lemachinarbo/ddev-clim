package ddev

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Project struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	AppRoot  string `json:"approot"`
	HttpUrl  string `json:"httpurl"`
	HttpsUrl string `json:"httpsurl"`
}

type ListOutput struct {
	Raw []Project `json:"raw"`
}

func GetProjects() ([]Project, error) {
	cmd := exec.Command("ddev", "list", "-j")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run ddev list: %w", err)
	}

	var listOutput ListOutput
	if err := json.Unmarshal(output, &listOutput); err != nil {
		return nil, fmt.Errorf("failed to parse ddev list output: %w", err)
	}

	return listOutput.Raw, nil
}

func ScanForProjects(root string) ([]Project, error) {
	var projects []Project
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() && info.Name() == ".ddev" {
			appRoot := filepath.Dir(path)
			// Check if it's a valid ddev project by looking for config.yaml
			if _, err := os.Stat(filepath.Join(path, "config.yaml")); err == nil {
				// We found one. We can use `ddev describe -j` to get details
				cmd := exec.Command("ddev", "describe", "-j")
				cmd.Dir = appRoot
				out, err := cmd.Output()
				if err == nil {
					var desc struct {
						Raw Project `json:"raw"`
					}
					if err := json.Unmarshal(out, &desc); err == nil {
						projects = append(projects, desc.Raw)
					} else {
						// Fallback if describe fails
						projects = append(projects, Project{
							Name:    filepath.Base(appRoot),
							AppRoot: appRoot,
							Status:  "unknown",
						})
					}
				}
			}
			return filepath.SkipDir // Don't go deeper into .ddev
		}
		return nil
	})
	return projects, err
}

func StartProject(appRoot string) error {
	cmd := exec.Command("ddev", "start")
	cmd.Dir = appRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start ddev in %s: %w", appRoot, err)
	}
	return nil
}

func StopProject(appRoot string) error {
	cmd := exec.Command("ddev", "stop")
	cmd.Dir = appRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop ddev in %s: %w", appRoot, err)
	}
	return nil
}
