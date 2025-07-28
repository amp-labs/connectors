package netsuite

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/netsuite/internal/restapi"
)

const apiVersion = "v1"

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client & account
	common.RequireAuthenticatedClient
	common.RequireWorkspace

	// TODO: Expose concurrency knobs to the server.
	RESTAPI *restapi.Adapter
}

// API Reference: https://td2972271.app.netsuite.com/app/help/helpcenter.nl?fid=section_158151234003.html
// NewConnector.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	connector, err := components.Initialize(providers.Netsuite, params,
		func(base *components.Connector) (*Connector, error) {
			return &Connector{Connector: base}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	switch connector.Module() {
	case providers.NetsuiteModuleRESTAPI:
		adapter, err := restapi.NewAdapter(params)
		if err != nil {
			return nil, err
		}

		connector.RESTAPI = adapter
	default:
		return nil, fmt.Errorf("module %s not supported", connector.Module())
	}

	return connector, nil
}

func (c Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*connectors.ListObjectMetadataResult, error) {
	if c.RESTAPI != nil {
		return c.RESTAPI.ListObjectMetadata(ctx, objectNames)
	}

	return nil, common.ErrNotImplemented
}

func (c Connector) Read(ctx context.Context, params connectors.ReadParams) (*connectors.ReadResult, error) {
	if c.RESTAPI != nil {
		return c.RESTAPI.Read(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c Connector) Write(ctx context.Context, params connectors.WriteParams) (*connectors.WriteResult, error) {
	if c.RESTAPI != nil {
		return c.RESTAPI.Write(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c Connector) Delete(ctx context.Context, params connectors.DeleteParams) (*connectors.DeleteResult, error) {
	if c.RESTAPI != nil {
		return c.RESTAPI.Delete(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c Connector) setUnitTestBaseURL(url string) {
	if c.RESTAPI != nil {
		c.RESTAPI.SetUnitTestBaseURL(url)
	}
}
