package workday

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/workday/metadata"
)

type Connector struct {
	*components.Connector
	components.SchemaProvider

	// Require authenticated client
	common.RequireAuthenticatedClient
	common.RequireWorkspace
	common.RequireMetadata
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Workday, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
		RequireMetadata: common.RequireMetadata{
			ExpectedMetadataKeys: []string{"tenantName"},
		},
	}

	connector.SchemaProvider = schema.NewCompositeSchemaProvider(
		schema.NewOpenAPISchemaProvider(connector.ProviderContext.Module(), metadata.Schemas),
	)

	return connector, nil
}
