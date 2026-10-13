package clickup

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/clickup/internal/metadata"
)

var (
	_ connectors.ObjectMetadataConnector = (*Connector)(nil)
	_ connectors.ReadConnector           = (*Connector)(nil)
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient

	// supported operations
	components.SchemaProvider
	components.Reader

	// Workspace is used to resolve the majority of Objects.
	// ClickUp API v3 uses term workspace_id while v2 uses team_id
	workspace string
}

// NewConnector creates ClickUp connector.
//
// Developer docs: https://developer.clickup.com
// Ampersand docs: https://docs.ampersand.ai/provider-guides/clickup
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.ClickUp, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
		workspace: params.Workspace,
	}

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

	return connector, nil
}

func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	return metadata.Catalog.SelectResult(objectNames)
}

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Catalog.ReadPathWithIds(objectName,
		httpkit.PathIdentifier{Name: "workspace_id", Value: c.workspace},
	)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, path)
}
