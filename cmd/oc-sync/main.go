package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	_ "github.com/mattn/go-sqlite3"

	"github.com/dagimg-dot/oc-sync/internal/config"
	"github.com/dagimg-dot/oc-sync/internal/export"
	"github.com/dagimg-dot/oc-sync/internal/importer"
	"github.com/dagimg-dot/oc-sync/internal/list"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `Usage: oc-sync <command> [options]

Commands:
  list                 List sessions from the local database
  export <id...>       Export one or more sessions by ID to sync directory
  import               Import sessions from sync directory into local database
  sync                 Export new + import foreign sessions (bidirectional)
  status               Show sync status

Run "oc-sync help <command>" for command-specific help.
`)
}

func run() error {
	if len(os.Args) < 2 {
		usage()
		return nil
	}

	switch os.Args[1] {
	case "help", "-h", "--help":
		usage()
		return nil
	case "list":
		return cmdList()
	case "export":
		return cmdExport(os.Args[2:])
	case "import":
		return cmdImport()
	case "sync":
		return cmdSync()
	case "status":
		return cmdStatus()
	default:
		return fmt.Errorf("unknown command: %s — run 'oc-sync help' for usage", os.Args[1])
	}
}

func openDB(path string) (*sql.DB, error) {
	return sql.Open("sqlite3", fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_query_only=true", path))
}

func openDBRW(path string) (*sql.DB, error) {
	return sql.Open("sqlite3", fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000", path))
}

func cmdList() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	db, err := openDB(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	sessions, err := list.Sessions(db)
	if err != nil {
		return fmt.Errorf("list: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tPROJECT\tTOKENS\tUPDATED")
	fmt.Fprintln(w, "--\t-----\t-------\t------\t-------")
	for _, s := range sessions {
		tokens := fmt.Sprintf("%dK", (s.TokensInput+s.TokensOutput)/1000)
		project := shorten(s.ProjectID, 12)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			shorten(s.ID, 16),
			shorten(s.Title, 40),
			project,
			tokens,
			formatTime(s.TimeUpdated),
		)
	}
	w.Flush()
	fmt.Fprintf(os.Stderr, "\n%d session(s)\n", len(sessions))
	return nil
}

func cmdExport(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: oc-sync export <session-id...>")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	db, err := openDB(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	syncDir := filepath.Join(cfg.SyncDir, cfg.Hostname)
	var failed int
	for _, id := range args {
		if err := export.Session(db, id, syncDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: export %s: %v\n", id, err)
			failed++
			continue
		}
		fmt.Fprintf(os.Stderr, "exported %s\n", id)
	}
	if failed > 0 {
		return fmt.Errorf("%d export(s) failed", failed)
	}
	return nil
}

func cmdImport() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	db, err := openDBRW(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	syncDir := cfg.SyncDir
	entries, err := os.ReadDir(syncDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "sync directory does not exist")
			return nil
		}
		return fmt.Errorf("read sync dir: %w", err)
	}

	var imported int
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == cfg.Hostname {
			continue
		}
		machineDir := filepath.Join(syncDir, entry.Name())
		files, err := os.ReadDir(machineDir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
				continue
			}
			p := filepath.Join(machineDir, f.Name())
			if err := importer.Session(db, p, cfg.Mappings); err != nil {
				return fmt.Errorf("import %s: %w", p, err)
			}
			imported++
		}
	}

	fmt.Fprintf(os.Stderr, "imported %d session(s)\n", imported)
	return nil
}

func cmdSync() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	dbRO, err := openDB(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer dbRO.Close()

	dbRW, err := openDBRW(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer dbRW.Close()

	myDir := filepath.Join(cfg.SyncDir, cfg.Hostname)
	exported := 0
	sessions, err := list.Sessions(dbRO)
	if err != nil {
		return fmt.Errorf("list sessions: %w", err)
	}
	for _, s := range sessions {
		p := filepath.Join(myDir, s.ID+".json")
		if _, err := os.Stat(p); os.IsNotExist(err) {
			if err := export.Session(dbRO, s.ID, myDir); err != nil {
				return fmt.Errorf("export %s: %w", s.ID, err)
			}
			exported++
		}
	}

	imported := 0
	entries, err := os.ReadDir(cfg.SyncDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read sync dir: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == cfg.Hostname {
			continue
		}
		machineDir := filepath.Join(cfg.SyncDir, entry.Name())
		files, err := os.ReadDir(machineDir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
				continue
			}
			p := filepath.Join(machineDir, f.Name())
			if err := importer.Session(dbRW, p, cfg.Mappings); err != nil {
				fmt.Fprintf(os.Stderr, "warning: import %s: %v\n", p, err)
				continue
			}
			imported++
		}
	}

	fmt.Fprintf(os.Stderr, "exported %d, imported %d\n", exported, imported)
	return nil
}

func cmdStatus() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "config:  %s\n", config.ConfigPath())
	fmt.Fprintf(os.Stderr, "db:      %s\n", cfg.DBPath)
	fmt.Fprintf(os.Stderr, "sync:    %s\n", cfg.SyncDir)
	fmt.Fprintf(os.Stderr, "host:    %s\n", cfg.Hostname)
	fmt.Fprintf(os.Stderr, "peers:   ")

	entries, err := os.ReadDir(cfg.SyncDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "(none)")
	} else {
		var peers []string
		for _, e := range entries {
			if e.IsDir() && e.Name() != cfg.Hostname {
				peers = append(peers, e.Name())
			}
		}
		if len(peers) == 0 {
			fmt.Fprintln(os.Stderr, "(none)")
		} else {
			fmt.Fprintln(os.Stderr, strings.Join(peers, ", "))
		}
	}

	if cfg.Mappings != nil {
		fmt.Fprintf(os.Stderr, "mappings: %d\n", len(cfg.Mappings))
	}
	return nil
}

func shorten(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func formatTime(ts int64) string {
	if ts == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", ts)
}
