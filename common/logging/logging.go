package logging

import (
	"context"
	"log/slog"
)

// WithLoggerEnabled returns a new context with the logger
// explicitly enabled or disabled. If the key is not set, the
// logger will be enabled by default.
func WithLoggerEnabled(ctx context.Context, enabled bool) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, contextKey("loggerEnabled"), enabled)
}

// WithVerboseLogging returns a new context with the logger
// configured to be verbose or not. If the key is not set, the
// logger will *NOT* be verbose by default.
func WithVerboseLogging(ctx context.Context, verbose bool) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, contextKey("loggerVerbose"), verbose)
}

// With returns a new context with the given values added.
// The values are added to the logger automatically.
func With(ctx context.Context, values ...any) context.Context {
	if len(values) == 0 && ctx != nil {
		// Corner case, don't bother creating a new context.
		return ctx
	}

	vals := append(getValues(ctx), values...)

	return context.WithValue(ctx, contextKey("loggerValues"), vals)
}

// It's considered good practice to use unexported custom types for context keys.
// This avoids collisions with other packages that might be using the same string
// values for their own keys.
type contextKey string

func getValues(ctx context.Context) []any { //nolint:contextcheck
	if ctx == nil {
		ctx = context.Background()
	}

	// Check for a subsystem override.
	sub := ctx.Value(contextKey("loggerValues"))
	if sub != nil {
		val, ok := sub.([]any)
		if ok {
			return val
		} else {
			return nil
		}
	} else {
		return nil
	}
}

// IsLoggerEnabled returns true if the logger is enabled in the context.
func IsLoggerEnabled(ctx context.Context) bool {
	if ctx == nil {
		return true
	}

	sub := ctx.Value(contextKey("loggerEnabled"))
	if sub != nil {
		val, ok := sub.(bool)
		if ok {
			return val
		}
	}

	return true
}

// IsVerboseLogging returns true if the logger is configured to allow verbose messages.
func IsVerboseLogging(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	verb := ctx.Value(contextKey("loggerVerbose"))
	if verb != nil {
		val, ok := verb.(bool)
		if ok {
			return val
		}
	}

	return false
}

// Logger returns a logger.
//
//nolint:contextcheck,cyclop
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

	if !IsLoggerEnabled(realCtx) {
		// The logger has been explicitly disabled.
		//
		// It's much, much simpler to just return a logger which
		// throws everything away, than to add a check everywhere
		// we might want to log something.
		return nullLogger
	}

	// Get the default logger
	logger := slog.Default()

	// Check for key-values to add to the logger.
	vals := getValues(realCtx)
	if vals != nil {
		logger = logger.With(vals...)
	}

	// Return the logger
	return logger
}

// VerboseLogger returns a logger used for more verbose logging.
//
//nolint:contextcheck,cyclop
func VerboseLogger(ctx ...context.Context) *slog.Logger {
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

	if !IsLoggerEnabled(realCtx) {
		// The logger has been explicitly disabled.
		//
		// It's much, much simpler to just return a logger which
		// throws everything away, than to add a check everywhere
		// we might want to log something.
		return nullLogger
	}

	if !IsVerboseLogging(realCtx) {
		// The caller wants to log something verbose,
		// but the logger is not configured to be verbose.
		// So we'll just return a logger which throws everything away.
		return nullLogger
	}

	// Get the default logger
	logger := slog.Default()

	// Check for key-values to add to the logger.
	vals := getValues(realCtx)
	if vals != nil {
		logger = logger.With(vals...)
	}

	// Return the logger
	return logger
}
