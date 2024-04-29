package utils

import (
	"errors"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
)

var defaultSentryDSN = "https://f96f875b76304994aed1827378054427@o294739.ingest.sentry.io/4504174762065920"

func InitSentry() {
	dsn, ok := os.LookupEnv("SENTRY_DSN")
	if !ok {
		dsn = defaultSentryDSN
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Debug:            os.Getenv("SENTRY_DEBUG") == "1",
		Release:          VersionString(),
		TracesSampleRate: 1.0,
		BeforeSend:       addRemoteURL,
	})
	_ = err // ignore
}

func TeardownSentry() {
	sentry.Flush(time.Second)
}

func addRemoteURL(ev *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	repoDir, err := GetRepositoryDir()
	if err != nil {
		return ev
	}

	codecraftersRemote, err := IdentifyGitRemote(repoDir)
	if err == nil {
		ev.Extra["codecrafters_remote"] = codecraftersRemote
	} else {
		var e1 NoCodecraftersRemoteFoundError
		var e2 MultipleCodecraftersRemotesFoundError

		switch {
		case errors.Is(err, &e1):
			ev.Extra["all_remotes"] = e1.Remotes
		case errors.Is(err, &e2):
			ev.Extra["all_remotes"] = e2.Remotes
		}
	}

	return ev
}
