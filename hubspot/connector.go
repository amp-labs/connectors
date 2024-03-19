package hubspot

import (
	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/proxy"
)

// Option is a function which mutates the hubspot connector configuration.
type Option func(params *Connector)

// Connector is a Hubspot connector that wraps around a proxy connector
// and extends it with Hubspot-specific functionality (read, write, etc).
type Connector struct {
	*proxy.Connector
}

// NewConnector returns a new hubspot connector.
func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	conn = &Connector{}

	for _, opt := range opts {
		opt(conn)
	}

	if conn.Connector == nil {
		conn = nil
		outErr = common.ErrNoProxyConnectorFound

		return
	}

	return
}

func WithProxyConnector(conn connectors.ProxyConnector) Option {
	return func(connector *Connector) {
		proxyConn, ok := conn.(*proxy.Connector)
		if !ok {
			return
		}

		connector.Connector = proxyConn
	}
}
