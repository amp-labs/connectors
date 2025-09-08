package nutshell

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/nutshell/internal/metadata"
)

type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient

	components.SchemaProvider
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Nutshell, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
	}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(connector.ProviderContext.Module(), metadata.Schemas)

	return connector, nil
}
