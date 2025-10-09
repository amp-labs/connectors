package paddle

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/scripts/openapi/dixa/metadata"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient

	// Supported operations
	components.SchemaProvider
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	// Create base connector with provider info
	return components.Initialize(providers.Fathom, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewCompositeSchemaProvider(
		schema.NewOpenAPISchemaProvider(connector.ProviderContext.Module(), metadata.Schemas),
	)

	return connector, nil
}
