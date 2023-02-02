package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/codecrafters-io/cli/internal/commands"
	"github.com/codecrafters-io/cli/internal/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
		fmt.Fprintf(os.Stderr, "%v\n", err)
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

	ctx := context.Background()

	logger := newLogger()

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

func newLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	var logWriter io.Writer

	switch logFmt := os.Getenv("CODECRAFTERS_LOG_FORMAT"); logFmt {
	default:
		fallthrough
	case "pretty":
		logWriter = zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.TimeFormat = "15:04:05.000"

			w.FormatMessage = func(x interface{}) string {
				return fmt.Sprintf("%-20s", x)
			}
		})
	case "json":
		logWriter = os.Stderr
	}

	logger := zerolog.New(logWriter).With().Timestamp().Logger()
	logger = logger.Level(zerolog.InfoLevel)

	if q := os.Getenv("CODECRAFTERS_LOG_LEVEL"); q != "" {
		lvl, err := zerolog.ParseLevel(q)
		if err == nil {
			logger = logger.Level(lvl)
		} else {
			logger.Warn().Err(err).Msg("parse log level")
		}
	}

	return logger
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
