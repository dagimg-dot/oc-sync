package export

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/dagimg-dot/oc-sync/internal/types"
)

const testSchema = `
CREATE TABLE IF NOT EXISTS project (
	id TEXT PRIMARY KEY,
	worktree TEXT NOT NULL,
	vcs TEXT,
	name TEXT,
	time_created INTEGER NOT NULL,
	time_updated INTEGER NOT NULL,
	sandboxes TEXT NOT NULL DEFAULT '[]'
);
CREATE TABLE IF NOT EXISTS session (
	id TEXT PRIMARY KEY,
	project_id TEXT NOT NULL,
	parent_id TEXT,
	slug TEXT NOT NULL,
	directory TEXT NOT NULL,
	title TEXT NOT NULL,
	version TEXT NOT NULL DEFAULT '',
	time_created INTEGER NOT NULL,
	time_updated INTEGER NOT NULL,
	agent TEXT,
	model TEXT,
	path TEXT,
	cost REAL DEFAULT 0 NOT NULL,
	tokens_input INTEGER DEFAULT 0 NOT NULL,
	tokens_output INTEGER DEFAULT 0 NOT NULL
);
CREATE TABLE IF NOT EXISTS message (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL,
	time_created INTEGER NOT NULL,
	time_updated INTEGER NOT NULL,
	data TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS part (
	id TEXT PRIMARY KEY,
	message_id TEXT NOT NULL,
	session_id TEXT NOT NULL,
	time_created INTEGER NOT NULL,
	time_updated INTEGER NOT NULL,
	data TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS todo (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL,
	content TEXT NOT NULL,
	status TEXT NOT NULL,
	priority TEXT NOT NULL,
	position INTEGER NOT NULL,
	time_created INTEGER NOT NULL,
	time_updated INTEGER NOT NULL
);
`

func setupDB(t *testing.T, seeds []string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if _, err := db.Exec(testSchema); err != nil {
		t.Fatalf("schema: %v", err)
	}
	for _, s := range seeds {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	return db
}

func TestRandID(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := randID()
		if len(id) != 16 {
			t.Fatalf("randID() = %q (len %d), want 16 hex chars", id, len(id))
		}
		if ids[id] {
			t.Fatalf("collision on iteration %d: %s", i, id)
		}
		ids[id] = true
	}
}

func TestWriteExport(t *testing.T) {
	dir := t.TempDir()
	exp := &types.SessionExport{
		Version: 1,
		Source:  "testhost",
		Session: types.Session{ID: "ses_test", Title: "Test", ProjectID: "p1"},
		Project: types.Project{ID: "p1", Worktree: "/repo"},
	}

	if err := writeExport(dir, exp); err != nil {
		t.Fatalf("writeExport: %v", err)
	}

	path := filepath.Join(dir, "ses_test.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("export file not created")
	}

	data, _ := os.ReadFile(path)
	var decoded types.SessionExport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.Session.ID != "ses_test" {
		t.Errorf("want session ID 'ses_test', got %q", decoded.Session.ID)
	}
	if decoded.Version != 1 {
		t.Errorf("want version 1, got %d", decoded.Version)
	}
}

func TestWriteExport_atomicTemp(t *testing.T) {
	dir := t.TempDir()
	exp := &types.SessionExport{
		Version: 1,
		Session: types.Session{ID: "ses_atomic", Title: "T", ProjectID: "p1"},
		Project: types.Project{ID: "p1", Worktree: "/r"},
	}

	if err := writeExport(dir, exp); err != nil {
		t.Fatalf("writeExport: %v", err)
	}

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("leftover .tmp file: %s", e.Name())
		}
	}
}

func TestQuerySession(t *testing.T) {
	db := setupDB(t, []string{
		`INSERT INTO project (id, worktree, time_created, time_updated, sandboxes) VALUES ('p1', '/repo', 1, 1, '[]')`,
		`INSERT INTO session (id, project_id, slug, directory, title, version, time_created, time_updated, agent, model, tokens_input, tokens_output) VALUES ('s1', 'p1', 'slug', '/repo', 'Test', '1', 100, 200, 'opencode', 'gpt4', 500, 200)`,
	})

	s, err := querySession(db, "s1")
	if err != nil {
		t.Fatalf("querySession: %v", err)
	}
	if s.ID != "s1" || s.Title != "Test" || s.TokensInput != 500 {
		t.Errorf("unexpected session: %+v", s)
	}
}

func TestQuerySession_missing(t *testing.T) {
	db := setupDB(t, nil)
	_, err := querySession(db, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing session")
	}
}

func TestQueryProject(t *testing.T) {
	db := setupDB(t, []string{
		`INSERT INTO project (id, worktree, vcs, name, time_created, time_updated, sandboxes) VALUES ('p1', '/repo', 'git', 'test', 100, 200, '[]')`,
	})
	p, err := queryProject(db, "p1")
	if err != nil {
		t.Fatalf("queryProject: %v", err)
	}
	if p.ID != "p1" || p.VCS != "git" || p.Name != "test" {
		t.Errorf("unexpected project: %+v", p)
	}
}

func TestQueryMessages(t *testing.T) {
	db := setupDB(t, []string{
		`INSERT INTO project (id, worktree, time_created, time_updated, sandboxes) VALUES ('p1', '/r', 1, 1, '[]')`,
		`INSERT INTO session (id, project_id, slug, directory, title, version, time_created, time_updated) VALUES ('s1', 'p1', 'a', '/r', 't', '1', 1, 2)`,
		`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES ('m1', 's1', 100, 100, '{"role":"user"}')`,
		`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES ('m2', 's1', 200, 200, '{"role":"assistant"}')`,
	})
	msgs, err := queryMessages(db, "s1")
	if err != nil {
		t.Fatalf("queryMessages: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("want 2 messages, got %d", len(msgs))
	}
	if msgs[0].Data != `{"role":"user"}` {
		t.Errorf("first message data mismatch: %q", msgs[0].Data)
	}
}

func TestQueryParts(t *testing.T) {
	db := setupDB(t, []string{
		`INSERT INTO project (id, worktree, time_created, time_updated, sandboxes) VALUES ('p1', '/r', 1, 1, '[]')`,
		`INSERT INTO session (id, project_id, slug, directory, title, version, time_created, time_updated) VALUES ('s1', 'p1', 'a', '/r', 't', '1', 1, 2)`,
		`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES ('m1', 's1', 1, 1, '{}')`,
		`INSERT INTO part (id, message_id, session_id, time_created, time_updated, data) VALUES ('p1', 'm1', 's1', 1, 1, '{"type":"text","text":"hello"}')`,
	})
	parts, err := queryParts(db, "s1")
	if err != nil {
		t.Fatalf("queryParts: %v", err)
	}
	if len(parts) != 1 {
		t.Fatalf("want 1 part, got %d", len(parts))
	}
}

func TestQueryTodos(t *testing.T) {
	db := setupDB(t, []string{
		`INSERT INTO project (id, worktree, time_created, time_updated, sandboxes) VALUES ('p1', '/r', 1, 1, '[]')`,
		`INSERT INTO session (id, project_id, slug, directory, title, version, time_created, time_updated) VALUES ('s1', 'p1', 'a', '/r', 't', '1', 1, 2)`,
		`INSERT INTO todo (id, session_id, content, status, priority, position, time_created, time_updated) VALUES ('t1', 's1', 'fix it', 'pending', 'high', 0, 1, 1)`,
	})
	todos, err := queryTodos(db, "s1")
	if err != nil {
		t.Fatalf("queryTodos: %v", err)
	}
	if len(todos) != 1 {
		t.Fatalf("want 1 todo, got %d", len(todos))
	}
	if todos[0].Content != "fix it" {
		t.Errorf("want content 'fix it', got %q", todos[0].Content)
	}
}
