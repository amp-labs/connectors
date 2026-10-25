package twenty

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
)

// Connector reaches one Twenty instance, self-hosted or Cloud, with a workspace API key.
// The catalog's `workspace` input is the instance's domain, which every request is sent to.
type Connector struct {
	*components.Connector

	common.RequireAuthenticatedClient
	common.RequireWorkspace

	components.SchemaProvider
	components.Reader
	components.Writer
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.Twenty, params, constructor)
}

func constructor(_ common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	// Used by the metadata requests made through JSONHTTPClient.
	connector.HTTPClient().ErrorHandler = errorHandler

	connector.SchemaProvider = schema.NewDelegateSchemaProvider(connector.listObjectMetadata)

	// Every object is allowed through: which objects exist is the workspace's schema, custom objects
	// included. An object Twenty does not know is answered 400 ("object 'invoices' not found"), which
	// errorHandler maps to ErrBadRequest, so it is never retried.
	registry := components.NewEmptyEndpointRegistry()

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  errorHandler,
		},
	)

	connector.Writer = writer.NewHTTPWriter(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  connector.buildWriteRequest,
			ParseResponse: connector.parseWriteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	return connector, nil
}
