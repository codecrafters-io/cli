package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/codecrafters-io/cli/internal/commands"
	"github.com/codecrafters-io/cli/internal/utils"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
)

// Usage: codecrafters test
func main() {
	utils.InitSentry()
	defer utils.TeardownSentry()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `CLI to interact with CodeCrafters

VERSION
  %s

USAGE
  $ codecrafters [COMMAND]

COMMANDS
  test:  run tests on project in current directory
`, utils.VersionString())

	}

	help := flag.Bool("help", false, "show usage instructions")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Println(utils.VersionString())
		os.Exit(0)
	}

	err := run()
	if err != nil {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Fprintf(os.Stderr, "%v\n", red(err))
		os.Exit(1)
	}

	os.Exit(0)
}

func run() error {
	ctx := context.Background()
	logger := utils.NewLogger()
	cmd := flag.Arg(0)

	logger.Debug().Str("command", cmd).Msg("command")

	ctx = logger.WithContext(ctx)

	switch cmd {
	case "test":
		return commands.TestCommand(ctx)
	case "help",
		"": // no argument
		flag.Usage()
	default:
		log.Error().Str("command", cmd).Msgf("Unknown command. Did you mean to run \"codecrafters test\"?")
		log.Info().Msg("Run codecrafters help for a list of available commands.")

		return errors.New("bad usage")
	}

	return nil
}

func envOr(name, defaultVal string) string {
	v, ok := os.LookupEnv(name)
	if ok {
		return v
	}

	return defaultVal
}
