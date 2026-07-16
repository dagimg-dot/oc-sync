package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "oc-sync",
	Short: "Offline OpenCode session sync tool",
	Long: `Export OpenCode sessions as JSON to a shared directory (Syncthing, SSHFS, USB)
and import them on another machine. Per-session files with INSERT OR IGNORE
import for safe bidirectional merging.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(rootCmd.ErrOrStderr(), "error: %v\n", err)
	}
}
