package wordpress

import (
	"context"
	_ "embed"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/wordpress/internal/metadata"
)

var (
	_ connectors.ObjectMetadataConnector = (*Connector)(nil)
	_ connectors.ReadConnector           = (*Connector)(nil)
	_ connectors.WriteConnector          = (*Connector)(nil)
	_ connectors.DeleteConnector         = (*Connector)(nil)
)

type Connector struct {
	*components.Connector

	common.RequireAuthenticatedClient
	common.RequireWorkspace

	components.Reader
	components.Writer
	components.Deleter

	workspace string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.WordPress, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
		workspace: params.Workspace,
	}

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

func (c Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*connectors.ListObjectMetadataResult, error) {
	return metadata.Catalog.SelectResult(objectNames)
}

// The URL structure of WordPress endpoints has the following format:
//
//	"https://public-api.wordpress.com/{namespace}/{version}/{endpoint}"
//
// * {namespace}: The API namespace (e.g., ‘rest’, ‘wp’, ‘wpcom’)
// * {version}: The API version (e.g., ‘v1’, ‘v1.4’, ‘v2’, ‘v4’)
//
// Reference: https://developer.wordpress.com/docs/api/getting-started/#2-general-url-structure
func (c Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Catalog.ReadPathWithIds(objectName,
		httpkit.PathIdentifier{Name: "workspace_id", Value: c.workspace},
	)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, path)
}

func (c Connector) getWriteDeleteURL(objectName string, recordId string) (*urlbuilder.URL, error) {
	path, err := metadata.Catalog.ReadPathWithIds(objectName,
		httpkit.PathIdentifier{Name: "workspace_id", Value: c.workspace},
	)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, path, recordId)
}
