package testroutines

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// Proxy is a test suite useful for testing connectors.ProxyConnector interface.
type Proxy struct {
	Name                string
	Builder             ConnectorBuilder[connectors.ProxyConnector]
	ExpectedProxy       *connectors.ProxyConfig
	ExpectedModuleProxy *connectors.ProxyConfig
}

// Run provides a procedure to test connectors.ReadConnector
func (r Proxy) Run(t *testing.T) {
	t.Helper()

	conn := r.Builder.Build(t, r.Name)
	defaultProxy, err1 := conn.ProxyConfig()
	testutils.CheckOutputWithError(t, r.Name+" ProxyConfig()",
		r.ExpectedProxy, nil,
		defaultProxy, err1)

	moduleProxy, err2 := conn.ProxyModuleConfig()
	testutils.CheckOutputWithError(t, r.Name+" ProxyModuleConfig()",
		r.ExpectedModuleProxy, nil,
		moduleProxy, err2)
}
