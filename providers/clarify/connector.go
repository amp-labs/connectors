package clarify

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	*components.Connector

	common.RequireAuthenticatedClient
	// The workspace slug templates the catalog BaseURL (/v1/workspaces/{{.workspace}}). The catalog's
	// `workspace` metadata input reaches connectors as ConnectorParams.Workspace.
	common.RequireWorkspace

	components.SchemaProvider
	components.Reader
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.Clarify, params, constructor)
}

func constructor(_ common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	registry, err := components.NewEndpointRegistry(supportedOperations())
	if err != nil {
		return nil, err
	}

	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle

	// Used by the schemas requests made through JSONHTTPClient in ListObjectMetadata.
	connector.HTTPClient().ErrorHandler = errorHandler

	connector.SchemaProvider = schema.NewDelegateSchemaProvider(connector.listObjectMetadata)

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

	return connector, nil
}
