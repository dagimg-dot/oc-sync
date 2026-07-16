package importer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dagimg-dot/oc-sync/internal/types"
)

func Session(db *sql.DB, src string, mappings []types.Mapping) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open export: %w", err)
	}
	defer f.Close()

	var exp types.SessionExport
	if err := json.NewDecoder(f).Decode(&exp); err != nil {
		return fmt.Errorf("decode export: %w", err)
	}

	if exp.Version > types.ExportVersion {
		return fmt.Errorf("export version %d is newer than supported version %d", exp.Version, types.ExportVersion)
	}

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	remoteProjectID := exp.Project.ID
	if mapping := lookupMapping(mappings, remoteProjectID); mapping != nil {
		exp.Session.ProjectID = mapping.LocalProjectID
		exp.Project.ID = mapping.LocalProjectID
		exp.Project.Worktree = mapping.LocalWorktree
	} else if err := insertProject(tx, &exp.Project); err != nil {
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

	var todoHasID bool
	todoHasID, err = hasTodoIDColumn(tx)
	if err != nil {
		return fmt.Errorf("check todo schema: %w", err)
	}
	for i := range exp.Todos {
		if err := insertTodo(tx, &exp.Todos[i], todoHasID); err != nil {
			return fmt.Errorf("todo: %w", err)
		}
	}

	if err := recalcTokens(tx, exp.Session.ID, &exp.Session); err != nil {
		return fmt.Errorf("recalc tokens: %w", err)
	}

	return tx.Commit()
}

func lookupMapping(mappings []types.Mapping, remoteProjectID string) *types.Mapping {
	for i := range mappings {
		if mappings[i].RemoteProjectID == remoteProjectID {
			return &mappings[i]
		}
	}
	return nil
}

func insertProject(tx *sql.Tx, p *types.Project) error {
	_, err := tx.ExecContext(context.Background(), `
		INSERT OR IGNORE INTO project (id, worktree, vcs, name, time_created, time_updated, sandboxes)
		VALUES (?, ?, ?, ?, ?, ?, '[]')
	`, p.ID, p.Worktree, p.VCS, p.Name, p.TimeCreated, p.TimeUpdated)
	return err
}

func insertSession(tx *sql.Tx, s *types.Session) error {
	_, err := tx.ExecContext(context.Background(), `
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
	_, err := tx.ExecContext(context.Background(), `
		INSERT OR IGNORE INTO message (id, session_id, time_created, time_updated, data)
		VALUES (?, ?, ?, ?, ?)
	`, m.ID, m.SessionID, m.TimeCreated, m.TimeUpdated, m.Data)
	return err
}

func insertPart(tx *sql.Tx, p *types.Part) error {
	_, err := tx.ExecContext(context.Background(), `
		INSERT OR IGNORE INTO part (id, message_id, session_id, time_created, time_updated, data)
		VALUES (?, ?, ?, ?, ?, ?)
	`, p.ID, p.MessageID, p.SessionID, p.TimeCreated, p.TimeUpdated, p.Data)
	return err
}

func hasTodoIDColumn(tx *sql.Tx) (bool, error) {
	rows, err := tx.QueryContext(context.Background(), "PRAGMA table_info(todo)")
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue *string
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return false, err
		}
		if name == "id" {
			return true, nil
		}
	}
	return false, rows.Err()
}

func insertTodo(tx *sql.Tx, t *types.Todo, hasID bool) error {
	if hasID {
		_, err := tx.ExecContext(context.Background(), `
			INSERT OR IGNORE INTO todo (id, session_id, content, status, priority, position, time_created, time_updated)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, t.ID, t.SessionID, t.Content, t.Status, t.Priority, t.Position, t.TimeCreated, t.TimeUpdated)
		return err
	}
	_, err := tx.ExecContext(context.Background(), `
		INSERT OR IGNORE INTO todo (session_id, content, status, priority, position, time_created, time_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, t.SessionID, t.Content, t.Status, t.Priority, t.Position, t.TimeCreated, t.TimeUpdated)
	return err
}

func recalcTokens(tx *sql.Tx, sessionID string, s *types.Session) error {
	_, err := tx.ExecContext(context.Background(), `
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
