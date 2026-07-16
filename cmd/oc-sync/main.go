package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: oc-sync <command> [options]

Commands:
  list      List sessions with details
  export    Export selected sessions to sync directory
  import    Import sessions from sync directory
  sync      Export new + import foreign sessions
  status    Show sync status

Flags:
  -h, --help    Show this help
  -v, --verbose Verbose output

Use "oc-sync <command> -h" for command-specific help.
`)
	}

	if len(os.Args) < 2 {
		flag.Usage()
		return nil
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "list":
		return listSessions(args)
	case "export":
		return exportSessions(args)
	case "import":
		return importSessions(args)
	case "sync":
		return syncCmd(args)
	case "status":
		return showStatus(args)
	case "-h", "--help":
		flag.Usage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func listSessions(args []string) error {
	fmt.Println("list: not yet implemented")
	return nil
}

func exportSessions(args []string) error {
	fmt.Println("export: not yet implemented")
	return nil
}

func importSessions(args []string) error {
	fmt.Println("import: not yet implemented")
	return nil
}

func syncCmd(args []string) error {
	fmt.Println("sync: not yet implemented")
	return nil
}

func showStatus(args []string) error {
	fmt.Println("status: not yet implemented")
	return nil
}
