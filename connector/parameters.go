package connector

import "github.com/amp-labs/connectors/internal/parameters"

// Parameters is the unified input used to initialize any connector.
// Each connector may require a different subset of these fields.
type Parameters = parameters.Connector
