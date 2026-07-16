package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dagimg-dot/oc-sync/internal/config"
	"github.com/dagimg-dot/oc-sync/internal/export"
	"github.com/dagimg-dot/oc-sync/internal/importer"
	"github.com/dagimg-dot/oc-sync/internal/list"
	"github.com/dagimg-dot/oc-sync/internal/sync"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Bidirectional sync with peer machines",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdSync()
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
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
	files, err := sync.PeerFiles(cfg.SyncDir, cfg.Hostname)
	if err == nil {
		for _, f := range files {
			if err := importer.Session(dbRW, f.Path, cfg.Mappings); err != nil {
				fmt.Fprintf(os.Stderr, "warning: import %s: %v\n", f.Path, err)
				continue
			}
			imported++
		}
	}

	fmt.Fprintf(os.Stderr, "exported %d, imported %d\n", exported, imported)
	return nil
}
