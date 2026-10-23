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

const metadataKeyWorkspace = "workspace"

type Connector struct {
	*components.Connector

	common.RequireAuthenticatedClient
	common.RequireMetadata

	components.SchemaProvider
	components.Reader

	// Workspace is the slug every API path is scoped to: /v1/workspaces/{slug}/...
	Workspace string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.Clarify, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
		RequireMetadata: common.RequireMetadata{
			ExpectedMetadataKeys: []string{metadataKeyWorkspace},
		},
		Workspace: params.Metadata[metadataKeyWorkspace],
	}

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
