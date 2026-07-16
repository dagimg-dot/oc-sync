package export

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dagimg-dot/oc-sync/internal/types"
)

func Session(db *sql.DB, sessionID, syncDir string) error {
	exp, err := buildExport(db, sessionID)
	if err != nil {
		return err
	}
	return writeExport(syncDir, exp)
}

func buildExport(db *sql.DB, sessionID string) (*types.SessionExport, error) {
	session, err := querySession(db, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session %s: %w", sessionID, err)
	}

	project, err := queryProject(db, session.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("project %s: %w", session.ProjectID, err)
	}

	messages, err := queryMessages(db, sessionID)
	if err != nil {
		return nil, fmt.Errorf("messages: %w", err)
	}

	parts, err := queryParts(db, sessionID)
	if err != nil {
		return nil, fmt.Errorf("parts: %w", err)
	}

	todos, err := queryTodos(db, sessionID)
	if err != nil {
		return nil, fmt.Errorf("todos: %w", err)
	}

	hostname, _ := os.Hostname()

	return &types.SessionExport{
		Session:  *session,
		Project:  *project,
		Messages: messages,
		Parts:    parts,
		Todos:    todos,
		Source:   hostname,
	}, nil
}

func writeExport(syncDir string, exp *types.SessionExport) error {
	if err := os.MkdirAll(syncDir, 0755); err != nil {
		return fmt.Errorf("create sync dir: %w", err)
	}

	path := filepath.Join(syncDir, exp.Session.ID+".json")
	tmpPath := path + ".tmp"

	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(exp); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("encode json: %w", err)
	}
	f.Close()

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}

func querySession(db *sql.DB, id string) (*types.Session, error) {
	var s types.Session
	err := db.QueryRow(`
		SELECT id, project_id, COALESCE(parent_id,''), slug, directory,
		       COALESCE(path,''), title, COALESCE(agent,''), COALESCE(model,''),
		       COALESCE(cost,0), COALESCE(tokens_input,0), COALESCE(tokens_output,0),
		       time_created, time_updated
		FROM session WHERE id = ?
	`, id).Scan(
		&s.ID, &s.ProjectID, &s.ParentID, &s.Slug, &s.Directory,
		&s.Path, &s.Title, &s.Agent, &s.Model,
		&s.Cost, &s.TokensInput, &s.TokensOutput,
		&s.TimeCreated, &s.TimeUpdated,
	)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return &s, nil
}

func queryProject(db *sql.DB, id string) (*types.Project, error) {
	var p types.Project
	err := db.QueryRow(`
		SELECT id, worktree, COALESCE(vcs,''), COALESCE(name,''),
		       time_created, time_updated
		FROM project WHERE id = ?
	`, id).Scan(
		&p.ID, &p.Worktree, &p.VCS, &p.Name,
		&p.TimeCreated, &p.TimeUpdated,
	)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return &p, nil
}

func queryMessages(db *sql.DB, sessionID string) ([]types.Message, error) {
	rows, err := db.Query(`
		SELECT id, session_id, time_created, time_updated, data
		FROM message WHERE session_id = ?
		ORDER BY time_created, id
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []types.Message
	for rows.Next() {
		var m types.Message
		if err := rows.Scan(&m.ID, &m.SessionID, &m.TimeCreated, &m.TimeUpdated, &m.Data); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func queryParts(db *sql.DB, sessionID string) ([]types.Part, error) {
	rows, err := db.Query(`
		SELECT id, message_id, session_id, time_created, time_updated, data
		FROM part WHERE session_id = ?
		ORDER BY time_created, id
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []types.Part
	for rows.Next() {
		var p types.Part
		if err := rows.Scan(&p.ID, &p.MessageID, &p.SessionID, &p.TimeCreated, &p.TimeUpdated, &p.Data); err != nil {
			return nil, err
		}
		parts = append(parts, p)
	}
	return parts, rows.Err()
}

func queryTodos(db *sql.DB, sessionID string) ([]types.Todo, error) {
	rows, err := db.Query(`
		SELECT session_id, content, status, priority, position,
		       time_created, time_updated
		FROM todo WHERE session_id = ?
		ORDER BY position
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []types.Todo
	for rows.Next() {
		var t types.Todo
		if err := rows.Scan(&t.SessionID, &t.Content, &t.Status, &t.Priority, &t.Position,
			&t.TimeCreated, &t.TimeUpdated); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, rows.Err()
}
