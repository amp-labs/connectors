package components

import "time"

// Clock provides the current time, enabling deterministic testing of time-dependent logic.
// Implementations include production (real time) and test (fixed time) variants.
type Clock interface {
	Now() time.Time
}

// RealClock returns the current wall-clock time using time.Now().
type RealClock struct{}

func NewRealClock() *RealClock {
	return new(RealClock)
}

func (RealClock) Now() time.Time {
	return time.Now()
}

// FixedClock returns a fixed timestamp for reproducible tests.
// Use time.Date(...) or time.Now().Add(...) to set specific values.
type FixedClock struct {
	t time.Time
}

// NewFixedClock creates a FixedClock at the given time.
func NewFixedClock(t time.Time) *FixedClock {
	return &FixedClock{t: t}
}

func (c *FixedClock) Now() time.Time {
	return c.t
}
