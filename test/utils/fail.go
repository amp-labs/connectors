package utils

import (
	"log/slog"
	"os"
)

// Fail logs the message and exits with code 1.
func Fail(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
