package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/dagimg-dot/oc-sync/internal/config"
	"github.com/dagimg-dot/oc-sync/internal/list"
	"github.com/spf13/cobra"
)

func formatTokens(total int64) string {
	if total < 1000 {
		return fmt.Sprintf("%d", total)
	}
	return fmt.Sprintf("%.1fK", float64(total)/1000)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List sessions from the local database",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdList()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
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
		tokens := formatTokens(s.TokensInput + s.TokensOutput)
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
	t := time.UnixMilli(ts)
	if time.Since(t) < 24*time.Hour {
		return t.Format("15:04")
	}
	return t.Format("Jan 02")
}
