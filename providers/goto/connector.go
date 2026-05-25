// Package gotoconn implements the GoTo connector.
//
// The package is named "gotoconn" instead of "goto" because "goto" is a
// reserved keyword in Go and cannot be used as a package identifier. The
// "conn" suffix is short for "connector".
package gotoconn

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/goto/internal/gotocore"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client & account
	common.RequireAuthenticatedClient
	common.PostAuthInfo

	// gotoCore handles api.getgo.com endpoints (Webinar, etc).
	gotoCore *gotocore.Adapter

	// TODO: We don't have sandbox access to api.goto.com,
	// so the gotoconnect is not implemented yet.
	// gotoConnect *gotocore.Adapter

	accountKey string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	if params.Module == "" {
		params.Module = providers.ModuleGoTo
	}

	conn, err := components.Initialize(providers.GoTo, params,
		func(base *components.Connector) (*Connector, error) {
			return &Connector{Connector: base}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	authMetadata := NewAuthMetadataVars(params.Metadata)
	conn.accountKey = authMetadata.AccountKey

	if err := initModuleAdapters(conn, params); err != nil {
		return nil, err
	}

	return conn, nil
}

func initModuleAdapters(conn *Connector, params common.ConnectorParams) error {
	switch conn.Module() { //nolint:exhaustive
	case providers.ModuleGoTo:
		adapter, err := gotocore.NewAdapter(params, conn.accountKey)
		if err != nil {
			return err
		}

		conn.gotoCore = adapter
	case providers.ModuleGoToConnect:
		return common.ErrUnsupportedModule
		// adapter, err := gotocore.NewAdapter(params, conn.accountKey)
		// if err != nil {
		// 	return err
		// }

		// conn.gotoConnect = adapter
	default:
		return common.ErrUnsupportedModule
	}

	return nil
}

// SetBaseURL fans the override out to any active module adapter so that
// unit tests pointing at a mock server reach the same host the top-level
// connector now uses.
func (c *Connector) SetBaseURL(newURL string) {
	c.Connector.SetBaseURL(newURL)

	if c.gotoCore != nil {
		c.gotoCore.SetBaseURL(newURL)
	}
}

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*connectors.ListObjectMetadataResult, error) {
	if c.gotoCore != nil {
		return c.gotoCore.ListObjectMetadata(ctx, objectNames)
	}

	return nil, common.ErrNotImplemented
}

func (c *Connector) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if c.gotoCore != nil {
		return c.gotoCore.Read(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c *Connector) Write(ctx context.Context, params common.WriteParams) (*common.WriteResult, error) {
	if c.gotoCore != nil {
		return c.gotoCore.Write(ctx, params)
	}

	return nil, common.ErrNotImplemented
}
