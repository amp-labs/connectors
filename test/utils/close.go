package utils

import "io"

// Close closes the given Closer and fails if there is an error.
func Close(c io.Closer) {
	if err := c.Close(); err != nil {
		Fail("error closing: %w", err)
	}
}
