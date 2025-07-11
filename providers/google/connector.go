package google

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/google/internal/calendar"
)

// Connector for Google provider.
// Each adapter corresponds to Google Module implementation.
// Only one adapter can be non-nil and will be delegated to on reading/writing operations.
type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient

	Calendar *calendar.Adapter
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	connector, err := components.Initialize(providers.Google, params,
		func(base *components.Connector) (*Connector, error) {
			return &Connector{Connector: base}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	if connector.Module() == providers.ModuleGoogleCalendar {
		adapter, err := calendar.NewAdapter(params)
		if err != nil {
			return nil, err
		}

		connector.Calendar = adapter
	}

	return connector, nil
}

func (c Connector) setUnitTestBaseURL(url string) {
	if c.Calendar != nil {
		c.Calendar.SetUnitTestBaseURL(url)
	}
}

func (c Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*connectors.ListObjectMetadataResult, error) {
	if c.Calendar != nil {
		return c.Calendar.ListObjectMetadata(ctx, objectNames)
	}

	return nil, common.ErrNotImplemented
}

func (c Connector) Read(ctx context.Context, params connectors.ReadParams) (*connectors.ReadResult, error) {
	if c.Calendar != nil {
		return c.Calendar.Read(ctx, params)
	}

	return nil, common.ErrNotImplemented
}
