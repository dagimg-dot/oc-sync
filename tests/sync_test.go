package tests

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/dagimg-dot/oc-sync/internal/list"
)

// schema is the subset of OpenCode tables needed for sync.
const schema = `
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
	cost REAL DEFAULT 0 NOT NULL,
	tokens_input INTEGER DEFAULT 0 NOT NULL,
	tokens_output INTEGER DEFAULT 0 NOT NULL,
	FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS message (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL,
	time_created INTEGER NOT NULL,
	time_updated INTEGER NOT NULL,
	data TEXT NOT NULL,
	FOREIGN KEY (session_id) REFERENCES session(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS part (
	id TEXT PRIMARY KEY,
	message_id TEXT NOT NULL,
	session_id TEXT NOT NULL,
	time_created INTEGER NOT NULL,
	time_updated INTEGER NOT NULL,
	data TEXT NOT NULL,
	FOREIGN KEY (message_id) REFERENCES message(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS todo (
	session_id TEXT NOT NULL,
	content TEXT NOT NULL,
	status TEXT NOT NULL,
	priority TEXT NOT NULL,
	position INTEGER NOT NULL,
	time_created INTEGER NOT NULL,
	time_updated INTEGER NOT NULL,
	PRIMARY KEY (session_id, position),
	FOREIGN KEY (session_id) REFERENCES session(id) ON DELETE CASCADE
);
`

var seedA = []string{
	`INSERT INTO project (id, worktree, vcs, name, time_created, time_updated, sandboxes)
	 VALUES ('proj_aaa', '/home/jd/JDrive/Projects/GO/oc-sync', 'git', 'oc-sync', 1000, 1000, '[]')`,
	`INSERT INTO project (id, worktree, vcs, name, time_created, time_updated, sandboxes)
	 VALUES ('proj_global', '/', '', 'global', 1000, 1000, '[]')`,
	`INSERT INTO session (id, project_id, slug, directory, title, version, time_created, time_updated, agent, model, tokens_input, tokens_output)
	 VALUES ('ses_aaa', 'proj_aaa', 'fix-parser', '/home/jd/JDrive/Projects/GO/oc-sync', 'Fix parser bug', '1', 1000, 1500, 'opencode', 'gpt4', 500, 200)`,
	`INSERT INTO session (id, project_id, slug, directory, title, version, time_created, time_updated, agent, model, tokens_input, tokens_output)
	 VALUES ('ses_bbb', 'proj_global', 'notes', '/', 'Random notes', '1', 1100, 1200, 'opencode', 'gpt4', 100, 50)`,
	`INSERT INTO message (id, session_id, time_created, time_updated, data)
	 VALUES ('msg_a1', 'ses_aaa', 1000, 1000, '{"role":"user","summary":{"diffs":[]}}')`,
	`INSERT INTO message (id, session_id, time_created, time_updated, data)
	 VALUES ('msg_a2', 'ses_aaa', 1100, 1100, '{"role":"assistant","summary":{"diffs":[]}}')`,
	`INSERT INTO part (id, message_id, session_id, time_created, time_updated, data)
	 VALUES ('prt_a1', 'msg_a1', 'ses_aaa', 1000, 1000, '{"type":"text","text":"fix the parser please"}')`,
	`INSERT INTO part (id, message_id, session_id, time_created, time_updated, data)
	 VALUES ('prt_a2', 'msg_a2', 'ses_aaa', 1100, 1100, '{"type":"text","text":"done, fixed the tokenizer"}')`,
	`INSERT INTO part (id, message_id, session_id, time_created, time_updated, data)
	 VALUES ('prt_a3', 'msg_a2', 'ses_aaa', 1100, 1100, '{"type":"tool_use","text":"git diff"}')`,
	`INSERT INTO message (id, session_id, time_created, time_updated, data)
	 VALUES ('msg_b1', 'ses_bbb', 1150, 1150, '{"role":"user","summary":{"diffs":[]}}')`,
	`INSERT INTO part (id, message_id, session_id, time_created, time_updated, data)
	 VALUES ('prt_b1', 'msg_b1', 'ses_bbb', 1150, 1150, '{"type":"text","text":"note: remember to update the readme"}')`,
}

var seedB = []string{
	`INSERT INTO project (id, worktree, vcs, name, time_created, time_updated, sandboxes)
	 VALUES ('proj_bbb', '/home/jd/Work/oc-sync', 'git', 'oc-sync', 2000, 2000, '[]')`,
	`INSERT INTO project (id, worktree, vcs, name, time_created, time_updated, sandboxes)
	 VALUES ('proj_global', '/', '', 'global', 2000, 2000, '[]')`,
}

func setupDB(t *testing.T, seeds []string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	for _, s := range seeds {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("seed: %s -> %v", s[:60], err)
		}
	}
	return db
}

func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "oc-sync-test-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestListSessions(t *testing.T) {
	db := setupDB(t, seedA)

	sessions, err := list.Sessions(db)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("want 2 sessions, got %d", len(sessions))
	}
	if sessions[0].Title != "Fix parser bug" {
		t.Errorf("want title 'Fix parser bug', got %q", sessions[0].Title)
	}
	if sessions[0].TokensInput+sessions[0].TokensOutput != 700 {
		t.Errorf("want 700 total tokens, got %d", sessions[0].TokensInput+sessions[0].TokensOutput)
	}
}

func TestExportSession(t *testing.T) {
	db := setupDB(t, seedA)
	syncDir := tempDir(t)

	err := exportSession(db, "ses_aaa", syncDir)
	if err != nil {
		t.Fatalf("export session: %v", err)
	}

	exportPath := filepath.Join(syncDir, "ses_aaa.json")
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Fatalf("export file not found: %s", exportPath)
	}

	// TODO: parse JSON and verify contents
}

func TestExportSession_selectsOnlyRequested(t *testing.T) {
	db := setupDB(t, seedA)
	syncDir := tempDir(t)

	err := exportSession(db, "ses_aaa", syncDir)
	if err != nil {
		t.Fatalf("export session: %v", err)
	}

	// ses_bbb should NOT be exported
	exportPath := filepath.Join(syncDir, "ses_bbb.json")
	if _, err := os.Stat(exportPath); !os.IsNotExist(err) {
		t.Errorf("unwanted export file exists: %s", exportPath)
	}
}

func TestImportSession(t *testing.T) {
	// Setup: machine A exports, machine B imports
	exportDir := tempDir(t)
	dbA := setupDB(t, seedA)
	dbB := setupDB(t, seedB) // B has different project ID

	if err := exportSession(dbA, "ses_aaa", exportDir); err != nil {
		t.Fatalf("export: %v", err)
	}

	importPath := filepath.Join(exportDir, "ses_aaa.json")
	err := importSession(dbB, importPath)
	if err != nil {
		t.Fatalf("import: %v", err)
	}

	// Verify session exists in B
	var count int
	if err := dbB.QueryRow("SELECT COUNT(*) FROM session WHERE id = 'ses_aaa'").Scan(&count); err != nil {
		t.Fatalf("query session: %v", err)
	}
	if count != 1 {
		t.Errorf("session ses_aaa not found in machine B")
	}

	// Verify messages and parts also transferred
	if err := dbB.QueryRow("SELECT COUNT(*) FROM message WHERE session_id = 'ses_aaa'").Scan(&count); err != nil {
		t.Fatalf("query messages: %v", err)
	}
	if count != 2 {
		t.Errorf("want 2 messages, got %d", count)
	}

	if err := dbB.QueryRow("SELECT COUNT(*) FROM part WHERE session_id = 'ses_aaa'").Scan(&count); err != nil {
		t.Fatalf("query parts: %v", err)
	}
	if count != 3 {
		t.Errorf("want 3 parts, got %d", count)
	}
}

func TestImport_idempotent(t *testing.T) {
	exportDir := tempDir(t)
	dbA := setupDB(t, seedA)
	dbB := setupDB(t, seedB)

	if err := exportSession(dbA, "ses_aaa", exportDir); err != nil {
		t.Fatalf("export: %v", err)
	}

	importPath := filepath.Join(exportDir, "ses_aaa.json")
	if err := importSession(dbB, importPath); err != nil {
		t.Fatalf("first import: %v", err)
	}
	if err := importSession(dbB, importPath); err != nil {
		t.Fatalf("second import: %v", err)
	}

	var msgCount int
	if err := dbB.QueryRow("SELECT COUNT(*) FROM message WHERE session_id = 'ses_aaa'").Scan(&msgCount); err != nil {
		t.Fatalf("query messages: %v", err)
	}
	if msgCount != 2 {
		t.Errorf("want 2 messages after double import, got %d", msgCount)
	}
}

func TestImport_globalSession(t *testing.T) {
	// Global sessions (project_id = proj_global) don't need path mapping
	exportDir := tempDir(t)
	dbA := setupDB(t, seedA)
	dbB := setupDB(t, seedB)

	if err := exportSession(dbA, "ses_bbb", exportDir); err != nil {
		t.Fatalf("export: %v", err)
	}

	importPath := filepath.Join(exportDir, "ses_bbb.json")
	if err := importSession(dbB, importPath); err != nil {
		t.Fatalf("import global session: %v", err)
	}

	var title string
	if err := dbB.QueryRow("SELECT title FROM session WHERE id = 'ses_bbb'").Scan(&title); err != nil {
		t.Fatalf("query global session: %v", err)
	}
	if title != "Random notes" {
		t.Errorf("want title 'Random notes', got %q", title)
	}
}

func TestImport_divergentMerge(t *testing.T) {
	// Both machines have the same session but diverged
	exportDir := tempDir(t)
	dbA := setupDB(t, seedA)

	// B starts with A's session then adds its own messages
	dbB := setupDB(t, append(seedB,
		`INSERT INTO session (id, project_id, slug, directory, title, version, time_created, time_updated)
		 VALUES ('ses_aaa', 'proj_bbb', 'fix-parser', '/home/jd/Work/oc-sync', 'Fix parser bug', '1', 1000, 1600)`,
		`INSERT INTO message (id, session_id, time_created, time_updated, data)
		 VALUES ('msg_b2', 'ses_aaa', 1550, 1550, '{"role":"user","summary":{"diffs":[]}}')`,
		`INSERT INTO part (id, message_id, session_id, time_created, time_updated, data)
		 VALUES ('prt_b2', 'msg_b2', 'ses_aaa', 1550, 1550, '{"type":"text","text":"verify the fix handles edge cases"}')`,
	))

	// Export A's version (has msg_a1, msg_a2 — no msg_b2)
	if err := exportSession(dbA, "ses_aaa", exportDir); err != nil {
		t.Fatalf("export: %v", err)
	}

	// Import into B — B should end up with both A's and B's messages
	importPath := filepath.Join(exportDir, "ses_aaa.json")
	if err := importSession(dbB, importPath); err != nil {
		t.Fatalf("import: %v", err)
	}

	var msgCount int
	if err := dbB.QueryRow("SELECT COUNT(*) FROM message WHERE session_id = 'ses_aaa'").Scan(&msgCount); err != nil {
		t.Fatalf("query messages: %v", err)
	}
	if msgCount != 3 {
		t.Errorf("want 3 messages after merge (2 from A + 1 from B), got %d", msgCount)
	}

	// Session token totals should reflect all messages
	var tokensInput, tokensOutput int64
	err := dbB.QueryRow("SELECT tokens_input, tokens_output FROM session WHERE id = 'ses_aaa'").Scan(&tokensInput, &tokensOutput)
	if err != nil {
		t.Fatalf("query session tokens: %v", err)
	}
	if tokensInput+tokensOutput == 0 {
		t.Error("token totals should be > 0 after merge")
	}
}

func TestImport_withTodos(t *testing.T) {
	exportDir := tempDir(t)
	dbA := setupDB(t, append(seedA,
		`INSERT INTO todo (session_id, content, status, priority, position, time_created, time_updated)
		 VALUES ('ses_aaa', 'review the fix', 'pending', 'high', 0, 1000, 1000)`,
	))
	dbB := setupDB(t, seedB)

	if err := exportSession(dbA, "ses_aaa", exportDir); err != nil {
		t.Fatalf("export: %v", err)
	}

	importPath := filepath.Join(exportDir, "ses_aaa.json")
	if err := importSession(dbB, importPath); err != nil {
		t.Fatalf("import: %v", err)
	}

	var todoCount int
	if err := dbB.QueryRow("SELECT COUNT(*) FROM todo WHERE session_id = 'ses_aaa'").Scan(&todoCount); err != nil {
		t.Fatalf("query todos: %v", err)
	}
	if todoCount != 1 {
		t.Errorf("want 1 todo, got %d", todoCount)
	}
}

func exportSession(db *sql.DB, sessionID, syncDir string) error {
	return nil
}

func importSession(db *sql.DB, src string) error {
	return nil
}
