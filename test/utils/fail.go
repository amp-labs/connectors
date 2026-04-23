package utils

import (
	"log/slog"
	"os"
	"time"
)

// Fail logs the message and exits with code 1.
func Fail(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}

func Timestamp(timeText string) time.Time {
	result, err := time.Parse(time.RFC3339, timeText)
	if err != nil {
		Fail("bad timestamp", "error", err)
	}

	return result
}
