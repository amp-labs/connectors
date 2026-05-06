package manage

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
)

type Adapter struct {
	*components.Connector
	common.RequireAuthenticatedClient
	components.SchemaProvider
	components.Writer
	components.Deleter
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
	adapter.Writer = writer.NewHTTPWriter(
		adapter.HTTPClient().Client,
		registry,
		adapter.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  adapter.buildWriteRequest,
			ParseResponse: adapter.parseWriteResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	adapter.Deleter = deleter.NewHTTPDeleter(
		adapter.HTTPClient().Client,
		registry,
		adapter.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  adapter.buildDeleteRequest,
			ParseResponse: adapter.parseDeleteResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	return adapter, nil
}
