package list

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
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
	slug TEXT NOT NULL,
	directory TEXT NOT NULL,
	title TEXT NOT NULL,
	version TEXT NOT NULL DEFAULT '',
	time_created INTEGER NOT NULL,
	time_updated INTEGER NOT NULL,
	agent TEXT,
	model TEXT,
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
`

func setupDB(t *testing.T, seeds []string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
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

func TestSessions_empty(t *testing.T) {
	db := setupDB(t, nil)
	sessions, err := Sessions(db)
	if err != nil {
		t.Fatalf("Sessions: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("want 0 sessions, got %d", len(sessions))
	}
}

func TestSessions_single(t *testing.T) {
	db := setupDB(t, []string{
		`INSERT INTO project (id, worktree, time_created, time_updated, sandboxes) VALUES ('p1', '/repo', 1, 1, '[]')`,
		`INSERT INTO session (id, project_id, slug, directory, title, version, time_created, time_updated, agent, model, tokens_input, tokens_output) VALUES ('s1', 'p1', 'test', '/repo', 'Test session', '1', 100, 200, 'opencode', 'gpt4', 500, 200)`,
	})
	sessions, err := Sessions(db)
	if err != nil {
		t.Fatalf("Sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("want 1 session, got %d", len(sessions))
	}
	if sessions[0].Title != "Test session" {
		t.Errorf("want title 'Test session', got %q", sessions[0].Title)
	}
	if sessions[0].TokensInput+sessions[0].TokensOutput != 700 {
		t.Errorf("want 700 tokens, got %d", sessions[0].TokensInput+sessions[0].TokensOutput)
	}
	if sessions[0].Agent != "opencode" {
		t.Errorf("want agent 'opencode', got %q", sessions[0].Agent)
	}
}

func TestSessions_order(t *testing.T) {
	db := setupDB(t, []string{
		`INSERT INTO project (id, worktree, time_created, time_updated, sandboxes) VALUES ('p1', '/repo', 1, 1, '[]')`,
		`INSERT INTO session (id, project_id, slug, directory, title, version, time_created, time_updated, tokens_input, tokens_output) VALUES ('s1', 'p1', 'a', '/repo', 'Older', '1', 100, 100, 0, 0)`,
		`INSERT INTO session (id, project_id, slug, directory, title, version, time_created, time_updated, tokens_input, tokens_output) VALUES ('s2', 'p1', 'b', '/repo', 'Newer', '1', 100, 200, 0, 0)`,
	})
	sessions, err := Sessions(db)
	if err != nil {
		t.Fatalf("Sessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("want 2 sessions, got %d", len(sessions))
	}
	if sessions[0].Title != "Newer" {
		t.Errorf("want first session 'Newer' (newest first), got %q", sessions[0].Title)
	}
	if sessions[1].Title != "Older" {
		t.Errorf("want second session 'Older', got %q", sessions[1].Title)
	}
}

func TestSessions_nullFields(t *testing.T) {
	db := setupDB(t, []string{
		`INSERT INTO project (id, worktree, time_created, time_updated, sandboxes) VALUES ('p1', '/repo', 1, 1, '[]')`,
		`INSERT INTO session (id, project_id, slug, directory, title, version, time_created, time_updated) VALUES ('s1', 'p1', 'test', '/repo', 'Null fields', '1', 100, 200)`,
	})
	sessions, err := Sessions(db)
	if err != nil {
		t.Fatalf("Sessions with null fields: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("want 1 session, got %d", len(sessions))
	}
}
