package importer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dagimg-dot/oc-sync/internal/types"
)

var bg = context.Background()

func Session(db *sql.DB, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open export: %w", err)
	}
	defer f.Close()

	var exp types.SessionExport
	if err := json.NewDecoder(f).Decode(&exp); err != nil {
		return fmt.Errorf("decode export: %w", err)
	}

	tx, err := db.BeginTx(bg, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := insertProject(tx, &exp.Project); err != nil {
		return fmt.Errorf("project: %w", err)
	}

	if err := insertSession(tx, &exp.Session); err != nil {
		return fmt.Errorf("session: %w", err)
	}

	for i := range exp.Messages {
		if err := insertMessage(tx, &exp.Messages[i]); err != nil {
			return fmt.Errorf("message %s: %w", exp.Messages[i].ID, err)
		}
	}

	for i := range exp.Parts {
		if err := insertPart(tx, &exp.Parts[i]); err != nil {
			return fmt.Errorf("part %s: %w", exp.Parts[i].ID, err)
		}
	}

	for i := range exp.Todos {
		if err := insertTodo(tx, &exp.Todos[i]); err != nil {
			return fmt.Errorf("todo: %w", err)
		}
	}

	if err := recalcTokens(tx, exp.Session.ID, &exp.Session); err != nil {
		return fmt.Errorf("recalc tokens: %w", err)
	}

	return tx.Commit()
}

func insertProject(tx *sql.Tx, p *types.Project) error {
	_, err := tx.ExecContext(bg, `
		INSERT OR IGNORE INTO project (id, worktree, vcs, name, time_created, time_updated, sandboxes)
		VALUES (?, ?, ?, ?, ?, ?, '[]')
	`, p.ID, p.Worktree, p.VCS, p.Name, p.TimeCreated, p.TimeUpdated)
	return err
}

func insertSession(tx *sql.Tx, s *types.Session) error {
	_, err := tx.ExecContext(bg, `
		INSERT OR IGNORE INTO session
			(id, project_id, parent_id, slug, directory, path, title, version,
			 agent, model, cost, tokens_input, tokens_output,
			 time_created, time_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?, '',
			?, ?, ?, ?, ?,
			?, ?)
	`, s.ID, s.ProjectID, nullStr(s.ParentID), s.Slug, s.Directory, nullStr(s.Path),
		s.Title, nullStr(s.Agent), nullStr(s.Model),
		s.Cost, s.TokensInput, s.TokensOutput,
		s.TimeCreated, s.TimeUpdated)
	return err
}

func insertMessage(tx *sql.Tx, m *types.Message) error {
	_, err := tx.ExecContext(bg, `
		INSERT OR IGNORE INTO message (id, session_id, time_created, time_updated, data)
		VALUES (?, ?, ?, ?, ?)
	`, m.ID, m.SessionID, m.TimeCreated, m.TimeUpdated, m.Data)
	return err
}

func insertPart(tx *sql.Tx, p *types.Part) error {
	_, err := tx.ExecContext(bg, `
		INSERT OR IGNORE INTO part (id, message_id, session_id, time_created, time_updated, data)
		VALUES (?, ?, ?, ?, ?, ?)
	`, p.ID, p.MessageID, p.SessionID, p.TimeCreated, p.TimeUpdated, p.Data)
	return err
}

func insertTodo(tx *sql.Tx, t *types.Todo) error {
	_, err := tx.ExecContext(bg, `
		INSERT OR IGNORE INTO todo (session_id, content, status, priority, position, time_created, time_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, t.SessionID, t.Content, t.Status, t.Priority, t.Position, t.TimeCreated, t.TimeUpdated)
	return err
}

func recalcTokens(tx *sql.Tx, sessionID string, s *types.Session) error {
	_, err := tx.ExecContext(bg, `
		UPDATE session SET
			tokens_input = MAX(tokens_input, ?),
			tokens_output = MAX(tokens_output, ?),
			cost = MAX(cost, ?)
		WHERE id = ?
	`, s.TokensInput, s.TokensOutput, s.Cost, sessionID)
	return err
}

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
