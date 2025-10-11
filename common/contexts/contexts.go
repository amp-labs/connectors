package contexts

import "context"

// IsContextAlive returns true if the context is not done.
func IsContextAlive(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	// This is non-blocking, so it will return immediately.
	select {
	case <-ctx.Done():
		return false
	default:
		return true
	}
}
