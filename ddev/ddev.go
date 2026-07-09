package ddev

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
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

type ServiceInfo struct {
	Status string `json:"status"`
}

type DescribeRaw struct {
	Name            string                 `json:"name"`
	Status          string                 `json:"status"`
	PhpVersion      string                 `json:"php_version"`
	WebserverType   string                 `json:"webserver_type"`
	DatabaseType    string                 `json:"database_type"`
	DatabaseVersion string                 `json:"database_version"`
	MutagenEnabled  bool                   `json:"mutagen_enabled"`
	Services        map[string]ServiceInfo `json:"services"`
}

type DescribeOutput struct {
	Raw DescribeRaw `json:"raw"`
}

type CommandStream struct {
	Reader io.Reader
	Cmd    *exec.Cmd
	Mu     sync.Mutex
	Err    error
	Done   bool
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

func DescribeProject(appRoot string) (DescribeRaw, error) {
	cmd := exec.Command("ddev", "describe", "-j")
	cmd.Dir = appRoot
	output, err := cmd.Output()
	if err != nil {
		return DescribeRaw{}, fmt.Errorf("failed to run ddev describe: %w", err)
	}

	var desc DescribeOutput
	if err := json.Unmarshal(output, &desc); err != nil {
		return DescribeRaw{}, fmt.Errorf("failed to parse ddev describe output: %w", err)
	}

	return desc.Raw, nil
}

func StartProject(appRoot string) error {
	cmd := exec.Command("ddev", "start")
	cmd.Dir = appRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		outStr := strings.TrimSpace(string(output))
		if outStr == "" {
			outStr = err.Error()
		}
		return fmt.Errorf("failed to start: %s", outStr)
	}
	return nil
}

func StopProject(appRoot string) error {
	cmd := exec.Command("ddev", "stop")
	cmd.Dir = appRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		outStr := strings.TrimSpace(string(output))
		if outStr == "" {
			outStr = err.Error()
		}
		return fmt.Errorf("failed to stop: %s", outStr)
	}
	return nil
}

func StartCommandStream(appRoot string, action string) (*CommandStream, error) {
	cmd := exec.Command("ddev", action)
	cmd.Dir = appRoot

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = cmd.Stdout // Merge stderr into stdout

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	stream := &CommandStream{
		Reader: stdout,
		Cmd:    cmd,
	}

	go func() {
		state, waitErr := cmd.Process.Wait()
		stream.Mu.Lock()
		if waitErr != nil {
			stream.Err = waitErr
		} else if !state.Success() {
			stream.Err = fmt.Errorf("exit status %d", state.ExitCode())
		}
		stream.Done = true
		stream.Mu.Unlock()
		_ = stdout.Close()
	}()

	return stream, nil
}

func Poweroff() error {
	cmd := exec.Command("ddev", "poweroff")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to power off ddev: %w", err)
	}
	return nil
}

