package list

import (
	"database/sql"
	"fmt"

	"github.com/dagimg-dot/oc-sync/internal/types"
)

func Sessions(db *sql.DB) ([]types.Session, error) {
	rows, err := db.Query(`
		SELECT id, project_id, title, agent, model,
		       time_created, time_updated,
		       tokens_input, tokens_output
		FROM session
		ORDER BY time_updated DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []types.Session
	for rows.Next() {
		var s types.Session
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Title, &s.Agent, &s.Model,
			&s.TimeCreated, &s.TimeUpdated,
			&s.TokensInput, &s.TokensOutput); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sessions: %w", err)
	}
	return sessions, nil
}
