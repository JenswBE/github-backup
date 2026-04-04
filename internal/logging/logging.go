package logging

import (
	"log/slog"
	"os"
)

func Setup(verbose, console bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	var handler slog.Handler
	if console {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}
	slog.SetDefault(slog.New(handler))

	if verbose {
		slog.Debug("Verbose logging enabled")
	}
}
