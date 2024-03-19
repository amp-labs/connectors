package salesforce

import (
	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/proxy"
)

// Connector is a Salesforce connector.
type Connector struct {
	*proxy.Connector
}

type Option func(conn *Connector)

// NewConnector returns a new Salesforce connector.
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
