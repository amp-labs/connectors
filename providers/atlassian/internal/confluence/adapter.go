package confluence

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

// nolint:unused
const apiVersion = "/wiki/api/v2"

type Adapter struct {
	*components.Connector
	common.RequireAuthenticatedClient

	components.SchemaProvider
}

func NewAdapter(params common.ConnectorParams) (*Adapter, error) {
	return components.Initialize(providers.Atlassian, params, constructor)
}

func constructor(base *components.Connector) (*Adapter, error) {
	adapter := &Adapter{
		Connector: base,
	}

	adapter.SchemaProvider = schema.NewOpenAPISchemaProvider(adapter.ProviderContext.Module(), Schemas.Metadata)

	return adapter, nil
}
