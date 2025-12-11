package contexts

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsContextAlive(t *testing.T) {
	t.Parallel()

	t.Run("returns false for nil context", func(t *testing.T) {
		t.Parallel()
		// Note: Testing with nil context directly to verify nil handling
		assert.False(t, IsContextAlive(nil)) //nolint:staticcheck // Testing nil context behavior
	})

	t.Run("returns true for active context", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		assert.True(t, IsContextAlive(ctx))
	})

	t.Run("returns false for cancelled context", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		assert.False(t, IsContextAlive(ctx))
	})

	t.Run("returns false for expired context", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(t.Context(), 1*time.Millisecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond)
		assert.False(t, IsContextAlive(ctx))
	})

	t.Run("returns true for context with future deadline", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(t.Context(), 1*time.Hour)
		defer cancel()

		assert.True(t, IsContextAlive(ctx))
	})
}
