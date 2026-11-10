package auphan

import (
	"errors"
	"fmt"
	"time"
)

var errInvalidTimezone = errors.New("invalid timezone")

// loadLocation reads the "timezone" metadata, an IANA name like "America/Los_Angeles".
func loadLocation(name string) (*time.Location, error) {
	if name == "" {
		return nil, nil //nolint:nilnil // no store timezone configured
	}

	location, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("%w %q: %w", errInvalidTimezone, name, err)
	}

	return location, nil
}

// storeTime converts t to the store's timezone, so "Nov 5" means Nov 5 at the store.
// Without a configured timezone, t keeps its own: Auphan does not report a store's timezone.
func (c *Connector) storeTime(t time.Time) time.Time {
	if c.location == nil {
		return t
	}

	return t.In(c.location)
}
