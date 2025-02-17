package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

func InitLogger(serviceName string, debug bool) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if debug {
		opts.Level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	Log = slog.New(handler).With(
		slog.String("service", serviceName),
	)
	slog.SetDefault(Log)
}
