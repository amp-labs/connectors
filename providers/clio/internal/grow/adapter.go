package grow

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

type Adapter struct {
	*components.Connector
	common.RequireAuthenticatedClient
	components.SchemaProvider
	components.Reader
}

func NewAdapter(params common.ConnectorParams) (*Adapter, error) {
	return components.Initialize(providers.Clio, params, constructor)
}

func constructor(base *components.Connector) (*Adapter, error) {
	adapter := &Adapter{
		Connector: base,
	}

	adapter.SchemaProvider = schema.NewOpenAPISchemaProvider(
		adapter.ProviderContext.Module(),
		Schemas,
	)

	registry := components.NewEmptyEndpointRegistry()
	adapter.Reader = reader.NewHTTPReader(
		adapter.HTTPClient().Client,
		registry,
		adapter.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  adapter.buildReadRequest,
			ParseResponse: adapter.parseReadResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	return adapter, nil
}
