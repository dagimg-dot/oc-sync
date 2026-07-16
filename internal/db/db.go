package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func Open(path string) (*sql.DB, error) {
	return sql.Open("sqlite3", fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_query_only=true", path))
}

func OpenRW(path string) (*sql.DB, error) {
	return sql.Open("sqlite3", fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000", path))
}
