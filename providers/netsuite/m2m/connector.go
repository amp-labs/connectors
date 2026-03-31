package m2m

import (
	"context"
	"fmt"
	"time"
	_ "time/tzdata"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/netsuite/internal/restapi"
	"github.com/amp-labs/connectors/providers/netsuite/internal/restlet"
	"github.com/amp-labs/connectors/providers/netsuite/internal/suiteql"
)

// Connector is the NetSuite M2M connector. It uses OAuth 2.0 Client Credentials
// with JWT bearer assertion for authentication, and delegates all operations
// to the same adapters as the standard NetSuite connector.
type Connector struct {
	*components.Connector

	common.RequireAuthenticatedClient
	common.RequireWorkspace

	RESTAPI *restapi.Adapter
	SuiteQL *suiteql.Adapter
	RESTlet *restlet.Adapter

	instanceTimezone *time.Location
}

// NewConnector creates a new NetSuite M2M connector.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	connector, err := components.Initialize(providers.NetsuiteM2M, params,
		func(base *components.Connector) (*Connector, error) {
			return &Connector{Connector: base}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	if tz, ok := params.Metadata["sessionTimezone"]; ok && tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			connector.instanceTimezone = loc
		}
	}

	if err := initModuleAdapters(connector, params); err != nil {
		return nil, err
	}

	return connector, nil
}

// NewM2MAuthenticatedClient builds an authenticated HTTP client using M2M credentials.
// This is called by the server when creating a connector for an M2M connection.
// The credentials (clientId, certificateId, privateKey) are read from the custom auth
// inputs, and accountId comes from the workspace metadata.
func NewM2MAuthenticatedClient(
	ctx context.Context,
	accountID, clientID, certificateID, privateKeyPEM string,
) (common.AuthenticatedHTTPClient, error) {
	privKey, err := ParseECPrivateKey([]byte(privateKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("parsing M2M private key: %w", err)
	}

	return common.NewCustomAuthHTTPClient(ctx,
		common.WithCustomDynamicHeaders(
			NewHeadersGenerator(accountID, clientID, certificateID, privKey, DefaultScopes),
		),
	)
}

func initModuleAdapters(connector *Connector, params common.ConnectorParams) error {
	switch connector.Module() { //nolint:exhaustive
	case providers.ModuleNetsuiteRESTAPI:
		adapter, err := restapi.NewAdapterForProvider(providers.NetsuiteM2M, params)
		if err != nil {
			return err
		}

		connector.RESTAPI = adapter
	case providers.ModuleNetsuiteSuiteQL:
		adapter, err := suiteql.NewAdapterForProvider(providers.NetsuiteM2M, params)
		if err != nil {
			return err
		}

		connector.SuiteQL = adapter
	case providers.ModuleNetsuiteRESTlet:
		adapter, err := restlet.NewAdapterForProvider(providers.NetsuiteM2M, params)
		if err != nil {
			return err
		}

		connector.RESTlet = adapter
	default:
		return common.ErrUnsupportedModule
	}

	return nil
}

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*connectors.ListObjectMetadataResult, error) {
	if c.RESTAPI != nil {
		return c.RESTAPI.ListObjectMetadata(ctx, objectNames)
	}

	if c.SuiteQL != nil {
		return c.SuiteQL.ListObjectMetadata(ctx, objectNames)
	}

	if c.RESTlet != nil {
		return c.RESTlet.ListObjectMetadata(ctx, objectNames)
	}

	return nil, common.ErrNotImplemented
}

func (c *Connector) Read(ctx context.Context, params connectors.ReadParams) (*connectors.ReadResult, error) {
	params = c.convertTimestampsToInstanceTimezone(params)

	if c.RESTAPI != nil {
		return c.RESTAPI.Read(ctx, params)
	}

	if c.SuiteQL != nil {
		return c.SuiteQL.Read(ctx, params)
	}

	if c.RESTlet != nil {
		return c.RESTlet.Read(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c *Connector) Write(ctx context.Context, params connectors.WriteParams) (*connectors.WriteResult, error) {
	if c.RESTAPI != nil {
		return c.RESTAPI.Write(ctx, params)
	}

	if c.RESTlet != nil {
		return c.RESTlet.Write(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c *Connector) Search(ctx context.Context, params *common.SearchParams) (*common.SearchResult, error) {
	if c.RESTlet != nil {
		return c.RESTlet.Search(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c *Connector) Delete(ctx context.Context, params connectors.DeleteParams) (*connectors.DeleteResult, error) {
	if c.RESTAPI != nil {
		return c.RESTAPI.Delete(ctx, params)
	}

	if c.RESTlet != nil {
		return c.RESTlet.Delete(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c *Connector) convertTimestampsToInstanceTimezone(params connectors.ReadParams) connectors.ReadParams {
	if c.instanceTimezone == nil {
		return params
	}

	if !params.Since.IsZero() {
		params.Since = params.Since.In(c.instanceTimezone)
	}

	if !params.Until.IsZero() {
		params.Until = params.Until.In(c.instanceTimezone)
	}

	return params
}
