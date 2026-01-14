package linkedin

import (
	"context"
	_ "embed"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/linkedin/internal/ads"
	"github.com/amp-labs/connectors/providers/linkedin/internal/platform"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client & account
	common.RequireAuthenticatedClient

	Platform *platform.Adapter
	Ads      *ads.Adapter
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	connector, err := components.Initialize(providers.LinkedIn, params,
		func(base *components.Connector) (*Connector, error) {
			return &Connector{Connector: base}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	switch connector.Module() { //nolint:exhaustive
	case providers.ModuleLinkedInPlatform:
		adapter, err := platform.NewAdapter(params)
		if err != nil {
			return nil, err
		}

		connector.Platform = adapter
	case providers.ModuleLinkedInAds:
		adapter, err := ads.NewAdapter(params)
		if err != nil {
			return nil, err
		}

		connector.Ads = adapter
	default:
		return nil, common.ErrUnsupportedModule
	}

	return connector, nil
}

func (c Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*connectors.ListObjectMetadataResult, error) {
	// Platform is write-only, so it doesn't support metadata operations.
	if c.Ads != nil {
		return c.Ads.ListObjectMetadata(ctx, objectNames)
	}

	return nil, common.ErrNotImplemented
}

func (c Connector) Read(ctx context.Context, params connectors.ReadParams) (*connectors.ReadResult, error) {
	// Platform is write-only, so it doesn't support read operations.
	if c.Ads != nil {
		return c.Ads.Read(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c Connector) Write(ctx context.Context, params connectors.WriteParams) (*connectors.WriteResult, error) {
	if c.Platform != nil {
		return c.Platform.Write(ctx, params)
	}

	if c.Ads != nil {
		return c.Ads.Write(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c Connector) Delete(ctx context.Context, params connectors.DeleteParams) (*connectors.DeleteResult, error) {
	if c.Platform != nil {
		return c.Platform.Delete(ctx, params)
	}

	if c.Ads != nil {
		return c.Ads.Delete(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c Connector) setUnitTestBaseURL(url string) {
	if c.Platform != nil {
		c.Platform.SetUnitTestBaseURL(url)
	}

	if c.Ads != nil {
		c.Ads.SetUnitTestBaseURL(url)
	}
}
