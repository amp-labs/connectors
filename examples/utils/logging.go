// nolint:revive,godoclint
package utils

import (
	"log"
	"log/slog"
	"os"
)

func setupLogging() {
	lvl := slog.LevelInfo
	if *debug {
		lvl = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     lvl,
	})

	slog.SetDefault(slog.New(handler))

	def := log.Default()
	*def = *slog.NewLogLogger(handler, slog.LevelInfo)
}
