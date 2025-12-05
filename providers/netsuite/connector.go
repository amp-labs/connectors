package netsuite

import (
	"context"
	_ "embed"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/netsuite/internal/restapi"
	"github.com/amp-labs/connectors/providers/netsuite/internal/suiteql"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client & account
	common.RequireAuthenticatedClient
	common.RequireWorkspace

	// TODO: Make REST API module's concurrent record fetching limit configurable
	// (currently hardcoded to 5)
	// https://github.com/amp-labs/connectors/pull/1920#discussion_r2248615641
	RESTAPI *restapi.Adapter
	SuiteQL *suiteql.Adapter
}

// NewConnector is a connector constructor.
// API Reference: https://td2972271.app.netsuite.com/app/help/helpcenter.nl?fid=section_158151234003.html
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	connector, err := components.Initialize(providers.Netsuite, params,
		func(base *components.Connector) (*Connector, error) {
			return &Connector{Connector: base}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	switch connector.Module() { //nolint:exhaustive
	case providers.ModuleNetsuiteRESTAPI:
		adapter, err := restapi.NewAdapter(params)
		if err != nil {
			return nil, err
		}

		connector.RESTAPI = adapter
	case providers.ModuleNetsuiteSuiteQL:
		adapter, err := suiteql.NewAdapter(params)
		if err != nil {
			return nil, err
		}

		connector.SuiteQL = adapter
	default:
		return nil, common.ErrUnsupportedModule
	}

	return connector, nil
}

func (c Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*connectors.ListObjectMetadataResult, error) {
	if c.RESTAPI != nil {
		return c.RESTAPI.ListObjectMetadata(ctx, objectNames)
	}

	if c.SuiteQL != nil {
		return c.SuiteQL.ListObjectMetadata(ctx, objectNames)
	}

	return nil, common.ErrNotImplemented
}

func (c Connector) Read(ctx context.Context, params connectors.ReadParams) (*connectors.ReadResult, error) {
	if c.RESTAPI != nil {
		return c.RESTAPI.Read(ctx, params)
	}

	if c.SuiteQL != nil {
		return c.SuiteQL.Read(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c Connector) Write(ctx context.Context, params connectors.WriteParams) (*connectors.WriteResult, error) {
	if c.RESTAPI != nil {
		return c.RESTAPI.Write(ctx, params)
	}

	// SuiteQL is read-only, so it doesn't support write operations
	return nil, common.ErrNotImplemented
}

func (c Connector) Delete(ctx context.Context, params connectors.DeleteParams) (*connectors.DeleteResult, error) {
	if c.RESTAPI != nil {
		return c.RESTAPI.Delete(ctx, params)
	}

	// SuiteQL is read-only, so it doesn't support delete operations
	return nil, common.ErrNotImplemented
}
