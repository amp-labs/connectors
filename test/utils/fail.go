package utils

import (
	"log/slog"
	"os"
)

func Fail(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
