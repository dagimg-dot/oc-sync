package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("oc-sync %s (commit %s, built %s)\n", version, commit, buildDate)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func Version() string   { return version }
func Commit() string    { return commit }
func BuildDate() string { return buildDate }
