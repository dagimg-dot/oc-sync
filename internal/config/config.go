package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Mapping struct {
	RemoteHostname  string `yaml:"remote_hostname"`
	RemoteProjectID string `yaml:"remote_project_id"`
	RemoteWorktree  string `yaml:"remote_worktree"`
	LocalWorktree   string `yaml:"local_worktree"`
	LocalProjectID  string `yaml:"local_project_id"`
}

type Config struct {
	DBPath   string    `yaml:"db_path"`
	SyncDir  string    `yaml:"sync_dir"`
	Hostname string    `yaml:"hostname"`
	Mappings []Mapping `yaml:"mappings,omitempty"`
}

func Defaults() *Config {
	hostname, _ := os.Hostname()
	return &Config{
		DBPath:   expandPath("~/.local/share/opencode/opencode.db"),
		SyncDir:  expandPath("~/Sync/oc-sync"),
		Hostname: hostname,
	}
}

func ConfigPath() string {
	if p := os.Getenv("OC_SYNC_CONFIG"); p != "" {
		return p
	}
	return expandPath("~/.config/oc-sync/config.yaml")
}

func Load() (*Config, error) {
	cfg := Defaults()

	p := ConfigPath()
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	cfg.DBPath = expandPath(cfg.DBPath)
	cfg.SyncDir = expandPath(cfg.SyncDir)
	if cfg.Hostname == "" {
		cfg.Hostname, _ = os.Hostname()
	}
	return cfg, nil
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
