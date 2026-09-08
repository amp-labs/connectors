package wealthbox

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.Wealthbox, params, constructor)
}

func constructor(_ common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	return connector, nil
}
