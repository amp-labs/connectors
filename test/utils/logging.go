package utils

import (
	"log/slog"
	"os"
)

// SetupLogging sets up logging for the test suite.
func SetupLogging() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
