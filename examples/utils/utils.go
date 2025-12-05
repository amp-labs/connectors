// nolint:revive
package utils

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
)

func Run(f func(ctx context.Context) error) {
	flag.Parse()

	setupLogging()

	// Catch Ctrl+C and handle it gracefully by shutting down the context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := runHandler(ctx, f); err != nil {
		slog.Error("error encountered", "err", err)

		os.Exit(1) // nolint:gocritic
	}
}

func runHandler(ctx context.Context, f func(ctx context.Context) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	err = f(ctx)

	return err
}
