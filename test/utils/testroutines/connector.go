// Package testroutines holds a collection of common test procedures.
// They provide a framework to write mock tests.
package testroutines

import (
	"github.com/amp-labs/connectors"
)

// ConnectorBuilder is a callback method to construct and configure connector for testing.
// This is a factory method called for every test suite.
type ConnectorBuilder[C connectors.Connector] func() (C, error)
