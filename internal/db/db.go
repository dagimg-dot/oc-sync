package db

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQLite connection to the OpenCode database.
// All reads are done on a read-only connection to prevent accidental writes.
type DB struct {
	path string
	ro   *sql.DB
}

// OpenRO opens a read-only connection to the OpenCode database.
func OpenRO(path string) (*DB, error) {
	ro, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro&_journal_mode=WAL&_query_only=true", path))
	if err != nil {
		return nil, fmt.Errorf("open read-only db: %w", err)
	}
	return &DB{path: path, ro: ro}, nil
}

// OpenRW opens a read-write connection (for import).
func OpenRW(path string) (*DB, error) {
	rw, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000", path))
	if err != nil {
		return nil, fmt.Errorf("open read-write db: %w", err)
	}
	return &DB{path: path, ro: rw}, nil
}

// RO returns the read-only connection.
func (d *DB) RO() *sql.DB { return d.ro }

func (d *DB) Close() error { return d.ro.Close() }

var stmtGetSessionOnce sync.Once
