package main

import (
	"io"
	stdlog "log"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// appLogger is the process root logger. New/converted call sites use it (and its
// component/server children); un-migrated stdlib log.Printf calls are routed
// through it too (see initLogging), so output is uniform during the migration.
var appLogger zerolog.Logger

// initLogging builds the root zerolog logger and bridges the standard library
// logger into it. Configured via env:
//   - LOG_FORMAT=json → structured JSON lines (prod / log shipping); anything
//     else → a human-readable console writer (dev default).
//   - LOG_LEVEL=debug|info|warn|error|… → minimum level (default info).
func initLogging() {
	zerolog.SetGlobalLevel(parseLogLevel(os.Getenv("LOG_LEVEL")))

	var w io.Writer = os.Stderr
	if !strings.EqualFold(os.Getenv("LOG_FORMAT"), "json") {
		w = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"}
	}
	appLogger = zerolog.New(w).With().Timestamp().Logger()

	// Bridge stdlib log (every not-yet-migrated log.Printf) through zerolog so
	// all output shares one format. zerolog.Logger implements io.Writer.
	stdlog.SetFlags(0)
	stdlog.SetOutput(appLogger)
}

func parseLogLevel(s string) zerolog.Level {
	if s == "" {
		return zerolog.InfoLevel
	}
	if l, err := zerolog.ParseLevel(strings.ToLower(s)); err == nil {
		return l
	}
	return zerolog.InfoLevel
}
