package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Mapping stores a learned path mapping between machines for a project.
type Mapping struct {
	RemoteHostname  string `yaml:"remote_hostname"`
	RemoteProjectID string `yaml:"remote_project_id"`
	RemoteWorktree  string `yaml:"remote_worktree"`
	LocalWorktree   string `yaml:"local_worktree"`
	LocalProjectID  string `yaml:"local_project_id"`
}

// Config holds all configuration for oc-sync.
type Config struct {
	DBPath   string    `yaml:"db_path"`
	SyncDir  string    `yaml:"sync_dir"`
	Hostname string    `yaml:"hostname"`
	Mappings []Mapping `yaml:"mappings,omitempty"`
}

// Defaults returns a Config with sensible defaults.
func Defaults() *Config {
	hostname, _ := os.Hostname()
	return &Config{
		DBPath:   expandPath("~/.local/share/opencode/opencode.db"),
		SyncDir:  expandPath("~/Sync/oc-sync"),
		Hostname: hostname,
	}
}

// ConfigPath returns the default config file location.
func ConfigPath() string {
	if p := os.Getenv("OC_SYNC_CONFIG"); p != "" {
		return p
	}
	return expandPath("~/.config/oc-sync/config.yaml")
}

// Load reads config from disk, falling back to defaults.
func Load() (*Config, error) {
	return Defaults(), nil
}

func Save(cfg *Config) error {
	return fmt.Errorf("not implemented")
}

func expandPath(p string) string {
	if len(p) > 0 && p[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return filepath.Join(home, p[1:])
	}
	return p
}
