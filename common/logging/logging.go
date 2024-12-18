package logging

import (
	"context"
	"log/slog"
)

type keyValues []keyValue

type keyValue struct {
	Key   string
	Value any
}

// WithKeyValue returns a new context with the given key-value pair which
// will be picked up by the logger and used in the structured messages.
func WithKeyValue(ctx context.Context, key string, value any) context.Context {
	kv := getKeysAndValues(ctx)

	entry := keyValue{
		Key:   key,
		Value: value,
	}

	kv = append(kv, entry)

	return context.WithValue(ctx, contextKey("keyValues"), kv)
}

func (kv *keyValues) Slice() []any {
	if kv == nil {
		return nil
	}

	if len(*kv) == 0 {
		return nil
	}

	out := make([]any, 0, len(*kv)*2) //nolint:mnd

	for _, v := range *kv {
		out = append(out, v.Slice()...)
	}

	return out
}

func (kv *keyValue) Slice() []any {
	return []any{kv.Key, kv.Value}
}

// It's considered good practice to use unexported custom types for context keys.
// This avoids collisions with other packages that might be using the same string
// values for their own keys.
type contextKey string

// getKeysAndValues returns the key-values from the context.
// Used to build the logger.
func getKeysAndValues(ctx context.Context) keyValues { //nolint:contextcheck
	if ctx == nil {
		ctx = context.Background()
	}

	// Check for a subsystem override.
	sub := ctx.Value(contextKey("keyValues"))
	if sub != nil {
		val, ok := sub.(keyValues)
		if ok {
			return val
		} else {
			return nil
		}
	} else {
		return nil
	}
}

// Logger returns a logger.
//
//nolint:contextcheck
func Logger(ctx ...context.Context) *slog.Logger {
	if len(ctx) == 0 {
		return slog.Default()
	}

	var realCtx context.Context

	// Honestly we only care if there's zero or one contexts.
	// If there's more than one, we'll just use the first one.
	for _, c := range ctx {
		if c != nil {
			realCtx = c //nolint:fatcontext

			break
		}
	}

	if realCtx == nil {
		// No context provided, so we'll just use a sane default
		realCtx = context.Background()
	}

	// Get the default logger
	logger := slog.Default()

	kv := getKeysAndValues(realCtx)
	if kv != nil {
		logger = logger.With(kv.Slice()...)
	}

	return logger
}
