package salesfinity

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Salesfinity, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
	}

	return connector, nil
}
