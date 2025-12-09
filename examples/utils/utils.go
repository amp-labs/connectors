// nolint:revive,godoclint
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

func runHandler(ctx context.Context, function func(ctx context.Context) error) (err error) {
	defer func() {
		if re := recover(); re != nil {
			var ok bool

			err, ok = re.(error)
			if !ok {
				panic(re)
			}
		}
	}()

	err = function(ctx)

	return err
}
