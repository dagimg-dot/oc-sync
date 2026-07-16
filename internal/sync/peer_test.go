package sync

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestPeerFiles_empty(t *testing.T) {
	dir := t.TempDir()
	files, err := PeerFiles(dir, "myhost")
	if err != nil {
		t.Fatalf("PeerFiles: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("want 0 files, got %d", len(files))
	}
}

func TestPeerFiles_skipsSelf(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "myhost"), 0755)
	os.WriteFile(filepath.Join(dir, "myhost", "s1.json"), []byte("{}"), 0644)

	files, err := PeerFiles(dir, "myhost")
	if err != nil {
		t.Fatalf("PeerFiles: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("want 0 files (own host filtered), got %d", len(files))
	}
}

func TestPeerFiles_findsPeers(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "peer1"), 0755)
	os.WriteFile(filepath.Join(dir, "peer1", "s1.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(dir, "peer1", "s2.json"), []byte("{}"), 0644)

	files, err := PeerFiles(dir, "myhost")
	if err != nil {
		t.Fatalf("PeerFiles: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("want 2 files, got %d", len(files))
	}
	if files[0].Machine != "peer1" {
		t.Errorf("want machine 'peer1', got %q", files[0].Machine)
	}
}

func TestPeerFiles_skipsNonJSON(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "peer1"), 0755)
	os.WriteFile(filepath.Join(dir, "peer1", "s1.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(dir, "peer1", "notes.txt"), []byte("hello"), 0644)

	files, err := PeerFiles(dir, "myhost")
	if err != nil {
		t.Fatalf("PeerFiles: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("want 1 file (txt skipped), got %d", len(files))
	}
}

func TestPeerFiles_multiplePeers(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "peer1"), 0755)
	os.WriteFile(filepath.Join(dir, "peer1", "s1.json"), []byte("{}"), 0644)
	os.MkdirAll(filepath.Join(dir, "peer2"), 0755)
	os.WriteFile(filepath.Join(dir, "peer2", "s2.json"), []byte("{}"), 0644)

	files, err := PeerFiles(dir, "myhost")
	if err != nil {
		t.Fatalf("PeerFiles: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("want 2 files, got %d", len(files))
	}
}

func TestPendingExports_allExported(t *testing.T) {
	dir := t.TempDir()
	myDir := filepath.Join(dir, "myhost")
	os.MkdirAll(myDir, 0755)

	db := setupDB(t, []string{
		`INSERT INTO session (id, project_id, slug, directory, title, version, time_created, time_updated) VALUES ('s1', 'p1', 'a', '/r', 's1', '1', 1, 1)`,
		`INSERT INTO session (id, project_id, slug, directory, title, version, time_created, time_updated) VALUES ('s2', 'p1', 'b', '/r', 's2', '1', 1, 2)`,
	})

	os.WriteFile(filepath.Join(myDir, "s1.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(myDir, "s2.json"), []byte("{}"), 0644)

	n, err := PendingExports(db, myDir)
	if err != nil {
		t.Fatalf("PendingExports: %v", err)
	}
	if n != 0 {
		t.Fatalf("want 0 pending exports, got %d", n)
	}
}

func TestPendingExports_somePending(t *testing.T) {
	dir := t.TempDir()
	myDir := filepath.Join(dir, "myhost")
	os.MkdirAll(myDir, 0755)

	db := setupDB(t, []string{
		`INSERT INTO session (id, project_id, slug, directory, title, version, time_created, time_updated) VALUES ('s1', 'p1', 'a', '/r', 's1', '1', 1, 1)`,
		`INSERT INTO session (id, project_id, slug, directory, title, version, time_created, time_updated) VALUES ('s2', 'p1', 'b', '/r', 's2', '1', 1, 2)`,
	})

	os.WriteFile(filepath.Join(myDir, "s1.json"), []byte("{}"), 0644)

	n, err := PendingExports(db, myDir)
	if err != nil {
		t.Fatalf("PendingExports: %v", err)
	}
	if n != 1 {
		t.Fatalf("want 1 pending export, got %d", n)
	}
}

func TestPendingImports(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "peer1"), 0755)
	os.WriteFile(filepath.Join(dir, "peer1", "s1.json"), []byte("{}"), 0644)

	n, err := PendingImports(dir, "myhost")
	if err != nil {
		t.Fatalf("PendingImports: %v", err)
	}
	if n != 1 {
		t.Fatalf("want 1 pending import, got %d", n)
	}
}

func setupDB(t *testing.T, seeds []string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE session (id TEXT PRIMARY KEY, project_id TEXT NOT NULL DEFAULT '', slug TEXT NOT NULL DEFAULT '', directory TEXT NOT NULL DEFAULT '', title TEXT NOT NULL DEFAULT '', version TEXT NOT NULL DEFAULT '', time_created INTEGER NOT NULL DEFAULT 0, time_updated INTEGER NOT NULL DEFAULT 0)`); err != nil {
		t.Fatalf("schema: %v", err)
	}
	for _, s := range seeds {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	return db
}
