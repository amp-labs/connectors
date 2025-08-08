package copper

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/copper/internal/metadata"
)

type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient
	common.RequireMetadata

	components.SchemaProvider

	userEmail string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	connector, err := components.Initialize(providers.Copper, params, constructor)
	if err != nil {
		return nil, err
	}

	connector.userEmail = params.Metadata["userEmail"]

	return connector, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
		RequireMetadata: common.RequireMetadata{
			ExpectedMetadataKeys: []string{"userEmail"},
		},
	}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(connector.ProviderContext.Module(), metadata.Schemas)

	return connector, nil
}
