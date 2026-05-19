package meta

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/meta/internal/facebook"
)

// Connector for Facebook provider.
// Each adapter corresponds to Facebook Module implementation.
// Only one adapter can be non-nil and will be delegated to on reading/writing operations.
type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient

	Facebook *facebook.Adapter
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	connector, err := components.Initialize(providers.Meta, params,
		func(base *components.Connector) (*Connector, error) {
			return &Connector{Connector: base}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	if connector.Module() == providers.ModuleFacebook {
		adapter, err := facebook.NewAdapter(params)
		if err != nil {
			return nil, err
		}

		connector.Facebook = adapter
	}

	return connector, nil
}

func (c Connector) setUnitTestBaseURL(url string) {
	if c.Facebook != nil {
		c.Facebook.SetUnitTestBaseURL(url)
	}
}

func (c Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*connectors.ListObjectMetadataResult, error) {
	if c.Facebook != nil {
		return c.Facebook.ListObjectMetadata(ctx, objectNames)
	}

	return nil, common.ErrNotImplemented
}
