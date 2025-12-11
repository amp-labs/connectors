package pipeliner

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
	"github.com/amp-labs/connectors/providers/pipeliner/internal/metadata"
)

type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient

	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter

	workspace string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Pipeliner, params, makeConstructor(params))
}

func makeConstructor(params common.ConnectorParams) components.ConnectorConstructor[Connector] {
	return func(base *components.Connector) (*Connector, error) {
		connector := &Connector{
			Connector: base,
			workspace: params.Workspace,
		}

		connector.SchemaProvider = schema.NewOpenAPISchemaProvider(connector.ProviderContext.Module(), metadata.Schemas)

		errorHandler := interpreter.ErrorHandler{
			JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
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

		// Set the deleter for the connector
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
}

func (c *Connector) getURL(objectName, recordID string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.FindURLPath(common.ModuleRoot, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, "api/v100/rest/spaces/", c.workspace, path, recordID)
}

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	return c.getURL(objectName, "")
}

func (c *Connector) getWriteURL(objectName, recordID string) (*urlbuilder.URL, error) {
	return c.getURL(objectName, recordID)
}
