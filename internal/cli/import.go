package cli

import (
	"fmt"
	"os"

	"github.com/dagimg-dot/oc-sync/internal/config"
	"github.com/dagimg-dot/oc-sync/internal/importer"
	"github.com/dagimg-dot/oc-sync/internal/sync"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import sessions from peer machines",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdImport()
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
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

	files, err := sync.PeerFiles(cfg.SyncDir, cfg.Hostname)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sync directory does not exist")
		return nil
	}

	var imported int
	for _, f := range files {
		if err := importer.Session(db, f.Path, cfg.Mappings); err != nil {
			fmt.Fprintf(os.Stderr, "warning: import %s: %v\n", f.Path, err)
			continue
		}
		imported++
	}

	fmt.Fprintf(os.Stderr, "imported %d session(s)\n", imported)
	return nil
}
