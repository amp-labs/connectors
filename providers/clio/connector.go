package clio

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/clio/internal/grow"
	"github.com/amp-labs/connectors/providers/clio/internal/manage"
)

// Connector provides integration with the Clio provider.
//
// Module-specific behavior is delegated to internal adapters. Each adapter is
// built with components.Initialize (HubSpot/Salesforce style: the outer type
// does not embed *components.Connector).
type Connector struct {
	Grow   *grow.Adapter
	Manage *manage.Adapter
}

// NewConnector creates a new Clio connector.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	if params.Module == "" || params.Module == common.ModuleRoot {
		params.Module = providers.ModuleClioManage
	}

	conn := &Connector{}

	switch params.Module {
	case providers.ModuleClioGrow:
		adapter, err := grow.NewAdapter(params)
		if err != nil {
			return nil, err
		}

		conn.Grow = adapter
	case providers.ModuleClioManage:
		adapter, err := manage.NewAdapter(params)
		if err != nil {
			return nil, err
		}

		conn.Manage = adapter
	default:
		return nil, fmt.Errorf("%w: %s", common.ErrUnsupportedModule, params.Module)
	}

	return conn, nil
}

func (c *Connector) ListObjectMetadata(
	ctx context.Context,
	objectNames []string,
) (*connectors.ListObjectMetadataResult, error) {
	if c.Grow != nil {
		return c.Grow.ListObjectMetadata(ctx, objectNames)
	}

	if c.Manage != nil {
		return c.Manage.ListObjectMetadata(ctx, objectNames)
	}

	return nil, common.ErrNotImplemented
}
