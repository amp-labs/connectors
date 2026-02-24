package devrev

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/devrev/metadata"
)

type Connector struct {
	*components.Connector
	components.SchemaProvider
	common.RequireAuthenticatedClient
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.DevRev, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewCompositeSchemaProvider(
		schema.NewOpenAPISchemaProvider(connector.ProviderContext.Module(), metadata.Schemas),
	)

	return connector, nil
}
