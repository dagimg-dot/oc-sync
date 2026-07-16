package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dagimg-dot/oc-sync/internal/config"
	"github.com/dagimg-dot/oc-sync/internal/db"
	"github.com/dagimg-dot/oc-sync/internal/export"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export <session-id...>",
	Short: "Export sessions to sync directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdExport(args)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}

func cmdExport(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: oc-sync export <session-id...>")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	db, err := db.Open(cfg.DBPath)
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
