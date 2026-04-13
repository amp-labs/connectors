package paypalSandbox

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/paypal"
)

// Connector reuses the PayPal connector implementation. Sandbox and production
// share the same API Key;
type Connector = paypal.Connector

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return paypal.NewConnectorForProvider(providers.PayPalSandBox, params)
}
