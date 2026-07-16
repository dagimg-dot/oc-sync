package main

import (
	"fmt"
	"os"

	"github.com/dagimg-dot/oc-sync/internal/cli"
)

var helpText = map[string]string{
	"list": `Usage: oc-sync list

List sessions from the local database with title, project, token count, and
last updated time.`,
	"export": `Usage: oc-sync export <session-id...>

Export one or more sessions by ID to the sync directory. Each session is
written as a JSON file in <sync_dir>/<hostname>/<session-id>.json.`,
	"import": `Usage: oc-sync import

Import sessions from peer machine directories in the sync directory into the
local database. Sessions already present (same ID) are skipped.`,
	"sync": `Usage: oc-sync sync

Bidirectional sync: export new local sessions not yet in the sync directory,
then import foreign sessions from peer machines.`,
	"status": `Usage: oc-sync status

Show configuration, connected peers, and pending sync state.`,
}

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

func cmdHelp(args []string) {
	if len(args) == 0 {
		usage()
		return
	}
	if h, ok := helpText[args[0]]; ok {
		fmt.Fprintln(os.Stderr, h)
	} else {
		fmt.Fprintf(os.Stderr, "no help for %q\n", args[0])
	}
}

func run() error {
	if len(os.Args) < 2 {
		usage()
		return nil
	}

	switch os.Args[1] {
	case "help", "-h", "--help":
		cmdHelp(os.Args[2:])
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
	case "version":
		fmt.Printf("oc-sync %s (commit %s, built %s)\n", cli.Version(), cli.Commit(), cli.BuildDate())
		return nil
	default:
		return fmt.Errorf("unknown command: %s — run 'oc-sync help' for usage", os.Args[1])
	}
}
