package confluence

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

type Adapter struct {
	*components.Connector
	common.RequireAuthenticatedClient

	components.SchemaProvider
}

func NewAdapter(params common.ConnectorParams) (*Adapter, error) {
	return components.Init(providers.Atlassian, params, constructor)
}

func constructor(_ common.ConnectorParams, base *components.Connector) (*Adapter, error) {
	adapter := &Adapter{
		Connector: base,
	}

	adapter.SchemaProvider = schema.NewOpenAPISchemaProvider(adapter.ProviderContext.Module(), Schemas)

	return adapter, nil
}
