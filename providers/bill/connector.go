package bill

import (
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client & account
	common.RequireAuthenticatedClient
	common.PostAuthInfo

	components.SchemaProvider

	sessionId string // Session ID for Bill.com
}

// NewConnector.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	conn, err := components.Initialize(providers.Bill, params, constructor)
	if err != nil {
		return nil, err
	}

	authMetadata := NewAuthMetadataVars(params.Metadata)

	conn.sessionId = authMetadata.SessionId

	return conn, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	return connector, nil
}
