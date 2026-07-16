package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpen(t *testing.T) {
	dir, _ := os.MkdirTemp("", "oc-sync-db-test-*")
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "test.db")

	d, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	if err := d.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestOpenRW(t *testing.T) {
	dir, _ := os.MkdirTemp("", "oc-sync-db-test-*")
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "test.db")

	d, err := OpenRW(path)
	if err != nil {
		t.Fatalf("OpenRW: %v", err)
	}
	defer d.Close()

	if err := d.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}

	var n int
	if err := d.QueryRow("SELECT 1").Scan(&n); err != nil {
		t.Fatalf("query: %v", err)
	}
	if n != 1 {
		t.Fatalf("want 1, got %d", n)
	}
}

func TestOpenRO_rejectsWrite(t *testing.T) {
	dir, _ := os.MkdirTemp("", "oc-sync-db-test-*")
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "test.db")

	d, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	_, err = d.Exec("CREATE TABLE t (x int)")
	if err == nil {
		t.Error("expected error on write to read-only db")
	}
}
