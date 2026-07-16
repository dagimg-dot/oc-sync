package cli

import (
	"testing"
)

func TestFormatTokens(t *testing.T) {
	tests := []struct {
		total int64
		want  string
	}{
		{0, "0"},
		{500, "500"},
		{999, "999"},
		{1000, "1.0K"},
		{1500, "1.5K"},
		{12345, "12.3K"},
		{100000, "100.0K"},
	}
	for _, tt := range tests {
		got := formatTokens(tt.total)
		if got != tt.want {
			t.Errorf("formatTokens(%d) = %q, want %q", tt.total, got, tt.want)
		}
	}
}

func TestShorten(t *testing.T) {
	tests := []struct {
		s    string
		n    int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "he..."},
		{"abc", 3, "abc"},
		{"abcdef", 5, "ab..."},
	}
	for _, tt := range tests {
		got := shorten(tt.s, tt.n)
		if got != tt.want {
			t.Errorf("shorten(%q, %d) = %q, want %q", tt.s, tt.n, got, tt.want)
		}
	}
}

func TestFormatTime(t *testing.T) {
	if got := formatTime(0); got != "-" {
		t.Errorf("formatTime(0) = %q, want %q", got, "-")
	}
	if got := formatTime(1000000); got == "" {
		t.Error("formatTime(1000000) returned empty")
	}
}

func TestVersionGetters(t *testing.T) {
	if Version() == "" {
		t.Error("Version() should not be empty")
	}
}
