package logging

import (
	"context"
	"log/slog"
)

var nullLogger *slog.Logger

func init() {
	nullLogger = slog.New(&nullHandler{})
}

type nullHandler struct{}

func (n *nullHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}

func (n *nullHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

func (n *nullHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return n
}

func (n *nullHandler) WithGroup(_ string) slog.Handler {
	return n
}
