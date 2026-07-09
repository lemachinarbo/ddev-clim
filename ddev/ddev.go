package ddev

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type Project struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	AppRoot    string `json:"approot"`
	ShortRoot  string `json:"shortroot"`
	Type       string `json:"type"`
	PrimaryURL string `json:"primary_url"`
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

func Poweroff() error {
	cmd := exec.Command("ddev", "poweroff")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to power off ddev: %w", err)
	}
	return nil
}

