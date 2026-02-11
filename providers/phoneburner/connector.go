package phoneburner

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/phoneburner/metadata"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient

	// supported operations
	components.SchemaProvider
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.PhoneBurner, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(
		connector.ProviderContext.Module(),
		metadata.Schemas,
	)

	return connector, nil
}
