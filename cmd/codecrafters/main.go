package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/codecrafters-io/cli/internal/commands"
	"github.com/codecrafters-io/cli/internal/utils"
	"github.com/getsentry/sentry-go"
)

var defaultSentryDSN = "https://f96f875b76304994aed1827378054427@o294739.ingest.sentry.io/4504174762065920"

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
		fmt.Println(fmt.Sprintf("codecrafters %s", utils.VersionString()))
		os.Exit(0)
	}

	err := run()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func run() error {
	if dsn := envOr("SENTRY_DSN", defaultSentryDSN); dsn != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              dsn,
			Debug:            os.Getenv("SENTRY_DEBUG") == "1",
			Release:          utils.VersionString(),
			TracesSampleRate: 1.0,
			BeforeSend:       addRemoteURL,
		})
		_ = err // ignore

		defer sentry.Flush(time.Second)
	}

	cmd := flag.Arg(0)

	switch cmd {
	case "test":
		return commands.TestCommand()
	case "": // no argument
		flag.Usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: \"%s\". Did you mean to run \"codecrafters test\"?\n", cmd)
		fmt.Fprintf(os.Stderr, "Error: Run codecrafters help for a list of available commands.\n")

		return errors.New("bad usage")
	}

	return nil
}

func addRemoteURL(ev *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	repoDir, err := commands.GetRepositoryDir()
	if err != nil {
		return ev
	}

	codecraftersRemote, err := utils.IdentifyGitRemote(repoDir)
	if err == nil {
		ev.Extra["codecrafters_remote"] = codecraftersRemote
	} else {
		var e1 utils.NoCodecraftersRemoteFoundError
		var e2 utils.MultipleCodecraftersRemotesFoundError

		switch {
		case errors.Is(err, &e1):
			ev.Extra["all_remotes"] = e1.Remotes
		case errors.Is(err, &e2):
			ev.Extra["all_remotes"] = e2.Remotes
		}
	}

	return ev
}

func envOr(name, defaultVal string) string {
	v, ok := os.LookupEnv(name)
	if ok {
		return v
	}

	return defaultVal
}
