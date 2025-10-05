package log

import (
	"log/slog"
	"os"
)

var Logger = newLogger()

func newLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}
