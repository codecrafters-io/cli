package main

import (
	"flag"
	"fmt"
	"github.com/codecrafters-io/cli/internal/commands"
	"os"
)

var version string = "0"
var commit string = "unknown"

// Usage: codecrafters test
func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `CLI to interact with CodeCrafters

VERSION
  %s

USAGE
  $ codecrafters [COMMAND]

COMMANDS
  test:  run tests on project in current directory
`, fmt.Sprintf("v%s-%s", version, commit[:7]))

	}

	help := flag.Bool("help", false, "show usage instructions")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	cmd := flag.Arg(0)

	switch cmd {
	case "test":
		os.Exit(commands.TestCommand())
	case "": // no argument
		flag.Usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: \"%s\". Did you mean to run \"codecrafters test\"?\n", cmd)
		fmt.Fprintf(os.Stderr, "Error: Run codecrafters help for a list of available commands.\n")
		os.Exit(1)
	}
}
