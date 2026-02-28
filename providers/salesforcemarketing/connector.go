package salesforcemarketing

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/salesforcemarketing/metadata"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client and workspace
	common.RequireAuthenticatedClient
	common.RequireWorkspace

	// Supported operations
	components.SchemaProvider
}

func NewConnector(params common.Parameters) (*Connector, error) {
	// Create base connector with provider info
	return components.Initialize(providers.SalesforceMarketing, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(connector.ProviderContext.Module(), metadata.Schemas)

	return connector, nil
}
