package main

import (
	"flag"
	"fmt"
	"os"
)

// Usage: codecrafters test
func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `CLI to interact with CodeCrafters:

VERSION
  %s

USAGE
  $ codecrafters [COMMAND]

COMMANDS
  test:  run tests on project in current directory
`, "0.0.1")

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
		fmt.Println("Tests!")
	case "": // no argument
		fmt.Fprintf(os.Stderr, "Unknown command. Did you mean to run \"codecrafters test\"?\n")
		fmt.Fprintf(os.Stderr, "Run codecrafters help for a list of available commands.\n")
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: \"%s\". Did you mean to run \"codecrafters test\"?\n", cmd)
		fmt.Fprintf(os.Stderr, "Error: Run codecrafters help for a list of available commands.\n")
		os.Exit(1)
	}
}
