package wordpress

import (
	"context"
	_ "embed"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/wordpress/internal/metadata"
)

var _ connectors.ObjectMetadataConnector = (*Connector)(nil)

type Connector struct {
	*components.Connector

	common.RequireAuthenticatedClient
	common.RequireWorkspace

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

	return connector, nil
}

func (c Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*connectors.ListObjectMetadataResult, error) {
	return metadata.Catalog.SelectResult(objectNames)
}
