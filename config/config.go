package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	RunningProjects []string `json:"running_projects"`
}

func (c *Config) AddProject(name string) {
	for _, p := range c.RunningProjects {
		if p == name {
			return
		}
	}
	c.RunningProjects = append(c.RunningProjects, name)
}

func (c *Config) RemoveProject(name string) {
	var newProjects []string
	for _, p := range c.RunningProjects {
		if p != name {
			newProjects = append(newProjects, p)
		}
	}
	c.RunningProjects = newProjects
}

func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".config", "ddev-clim")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return "", err
		}
	}
	return filepath.Join(configDir, "config.json"), nil
}

func LoadConfig() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{RunningProjects: []string{}}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
