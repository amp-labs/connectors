package auphan

import (
	"errors"
	"fmt"
	"time"
)

var errInvalidTimezone = errors.New("invalid timezone")

// loadLocation reads the "timezone" metadata, an IANA name like "America/Los_Angeles".
// An empty name is UTC.
func loadLocation(name string) (*time.Location, error) {
	location, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("%w %q: %w", errInvalidTimezone, name, err)
	}

	return location, nil
}

// storeTime converts t to the store's timezone, so "Nov 5" means Nov 5 at the store.
func (c *Connector) storeTime(t time.Time) time.Time {
	return t.In(c.location)
}
