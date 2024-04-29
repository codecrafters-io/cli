package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func NewLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	logWriter := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = "15:04:05.000"

		w.FormatMessage = func(message interface{}) string {
			return fmt.Sprintf("%-20s", message)
		}
	})

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
