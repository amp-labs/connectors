package restapi

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/netsuite/internal/shared"
)

const apiVersion = "v1"

type Adapter struct {
	*components.Connector
	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter
}

func NewAdapter(params common.ConnectorParams) (*Adapter, error) {
	return components.Initialize(providers.Netsuite, params, constructor)
}

func constructor(base *components.Connector) (*Adapter, error) {
	adapter := &Adapter{
		Connector: base,
	}

	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(shared.ErrorFormats, shared.StatusCodeMapping),
	}.Handle

	// Set the metadata provider for the connector
	adapter.SchemaProvider = schema.NewObjectSchemaProvider(
		adapter.HTTPClient().Client,
		schema.FetchModeParallel,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  adapter.buildObjectMetadataRequest,
			ParseResponse: adapter.parseObjectMetadataResponse,
		},
	)

	adapter.Reader = reader.NewHTTPReader(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  adapter.buildReadRequest,
			ParseResponse: adapter.parseReadResponse,
			ErrorHandler:  errorHandler,
		},
	)

	adapter.Writer = writer.NewHTTPWriter(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  adapter.buildWriteRequest,
			ParseResponse: adapter.parseWriteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	adapter.Deleter = deleter.NewHTTPDeleter(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  adapter.buildDeleteRequest,
			ParseResponse: adapter.parseDeleteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	return adapter, nil
}
