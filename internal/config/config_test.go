package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	tests := []struct {
		input string
		want  string
	}{
		{"~/.config/oc-sync", filepath.Join(home, ".config/oc-sync")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}
	for _, tt := range tests {
		got := expandPath(tt.input)
		if got != tt.want {
			t.Errorf("expandPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	if cfg.DBPath == "" {
		t.Error("Defaults() returned empty DBPath")
	}
	if cfg.SyncDir == "" {
		t.Error("Defaults() returned empty SyncDir")
	}
	if cfg.Hostname == "" {
		t.Error("Defaults() returned empty Hostname")
	}
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()
	if path == "" {
		t.Error("ConfigPath() returned empty")
	}
}
