package search

import (
	"github.com/amp-labs/connectors/providers/connectwise/internal/core"
)

// Strategy handles search operations.
type Strategy struct {
	base *core.Base
}

// NewStrategy creates a new batch Strategy configured to Search records using conditions.
func NewStrategy(base *core.Base) *Strategy {
	return &Strategy{base: base}
}
