package utils

import "io"

func Close(c io.Closer) {
	if err := c.Close(); err != nil {
		Fail("error closing: %w", err)
	}
}
