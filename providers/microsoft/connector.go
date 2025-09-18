package microsoft

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/microsoft/metadata"
)

const apiVersion = "v1.0"

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient

	// Supported operations
	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Microsoft, params, constructor)
}

// nolint:funlen
func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
	}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(connector.ProviderContext.Module(), metadata.Schemas)

	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  errorHandler,
		},
	)

	connector.Writer = writer.NewHTTPWriter(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  connector.buildWriteRequest,
			ParseResponse: connector.parseWriteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	connector.Deleter = deleter.NewHTTPDeleter(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  connector.buildDeleteRequest,
			ParseResponse: connector.parseDeleteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	return connector, nil
}

func (c *Connector) getURL(parts ...string) (*urlbuilder.URL, error) {
	path := append([]string{apiVersion}, parts...)

	return urlbuilder.New(c.ProviderInfo().BaseURL, path...)
}
