package salesflare

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
	"github.com/amp-labs/connectors/providers/salesflare/internal/metadata"
)

type Connector struct {
	*components.Connector

	common.RequireAuthenticatedClient

	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Salesflare, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
	}

	fallbackSchema := schema.NewObjectSchemaProvider(
		base.HTTPClient().Client,
		schema.FetchModeSerial,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  connector.buildSingleHandlerRequest,
			ParseResponse: connector.parseSingleHandlerResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	connector.SchemaProvider = schema.NewCompositeSchemaProvider(
		schema.NewOpenAPISchemaProvider(connector.ProviderContext.Module(), metadata.Schemas),
		fallbackSchema,
	)

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

func (c Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, objectName)
}

func (c Connector) getWriteURL(objectName string, identifier string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, objectName, identifier)
}

func (c Connector) getDeleteURL(objectName string, identifier string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, objectName, identifier)
}
