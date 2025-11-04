package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/codecrafters-io/cli/internal/commands"
	"github.com/codecrafters-io/cli/internal/utils"
	"github.com/fatih/color"
)

// Usage: codecrafters test
func main() {
	utils.InitLogger()

	utils.InitSentry()
	defer utils.TeardownSentry()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `CLI to interact with CodeCrafters

USAGE
  $ codecrafters [command]

EXAMPLES
  $ codecrafters test              # Run tests without committing changes
  $ codecrafters test --previous   # Run tests for all previous stages and the current stage without committing changes
  $ codecrafters submit            # Commit changes & submit to move to next step

COMMANDS
  test:             Run tests without committing changes
  submit:           Commit changes & submit to move to next step
  task:             View current stage instructions
  update-buildpack: Update language version
  help:             Show usage instructions

VERSION
  %s
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

		if err.Error() != "" {
			fmt.Fprintf(os.Stderr, "%v\n", red(err))
		}

		os.Exit(1)
	}

	os.Exit(0)
}

func run() error {
	cmd := flag.Arg(0)
	utils.Logger.Debug().Msgf("Running command: %s", cmd)

	switch cmd {
	case "test":
		testCmd := flag.NewFlagSet("test", flag.ExitOnError)
		shouldTestPrevious := testCmd.Bool("previous", false, "run tests for the current stage and all previous stages in ascending order")
		testCmd.Parse(flag.Args()[1:]) // parse the args after the test command

		return commands.TestCommand(*shouldTestPrevious)
	case "submit":
		return commands.SubmitCommand()
	case "task":
		taskCmd := flag.NewFlagSet("task", flag.ExitOnError)
		stageSlug := taskCmd.String("stage", "", "view instructions for a specific stage (slug, +N, or -N)")
		raw := taskCmd.Bool("raw", false, "print instructions without pretty-printing")
		taskCmd.Parse(flag.Args()[1:])

		return commands.TaskCommand(*stageSlug, *raw)
	case "update-buildpack":
		return commands.UpdateBuildpackCommand()
	case "help",
		"": // no argument
		flag.Usage()
	default:
		red := color.New(color.FgRed).SprintFunc()
		fmt.Printf(red("Unknown command '%s'. Did you mean to run `codecrafters test`?\n\n"), cmd)
		fmt.Printf("Run `codecrafters help` for a list of available commands.\n")

		return fmt.Errorf("")
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
