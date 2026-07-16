package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dagimg-dot/oc-sync/internal/config"
	"github.com/dagimg-dot/oc-sync/internal/sync"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show configuration and sync state",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdStatus()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
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

	db, err := openDB(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	myDir := filepath.Join(cfg.SyncDir, cfg.Hostname)
	pendingExports, err := sync.PendingExports(db, myDir)
	if err != nil {
		return fmt.Errorf("pending exports: %w", err)
	}
	fmt.Fprintf(os.Stderr, "pending:  %d session(s) to export\n", pendingExports)

	files, err := sync.PeerFiles(cfg.SyncDir, cfg.Hostname)
	if err != nil || len(files) == 0 {
		fmt.Fprintln(os.Stderr, "peers:    (none)")
	} else {
		peers := map[string]int{}
		for _, f := range files {
			peers[f.Machine]++
		}
		var names []string
		for name := range peers {
			names = append(names, name)
		}
		fmt.Fprintf(os.Stderr, "peers:    %s\n", strings.Join(names, ", "))
		fmt.Fprintf(os.Stderr, "pending:  %d session(s) to import\n", len(files))
	}

	if len(cfg.Mappings) > 0 {
		fmt.Fprintf(os.Stderr, "mappings: %d\n", len(cfg.Mappings))
	}
	return nil
}
