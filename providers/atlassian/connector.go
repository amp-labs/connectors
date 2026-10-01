package atlassian

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/atlassian/internal/confluence"
	"github.com/amp-labs/connectors/providers/atlassian/internal/jira"
)

var (
	_ connectors.AuthMetadataConnector   = (*Connector)(nil)
	_ connectors.ProxyConnector          = (*Connector)(nil)
	_ connectors.ObjectMetadataConnector = (*Connector)(nil)
	_ connectors.ReadConnector           = (*Connector)(nil)
	_ connectors.WriteConnector          = (*Connector)(nil)
	_ connectors.DeleteConnector         = (*Connector)(nil)
)

type Connector struct {
	components.BaseConnector

	workspace  string
	jira       *jira.Adapter
	confluence *confluence.Adapter
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	if err := params.Validate(
		common.RequireAuthenticatedClient{},
		common.RequireWorkspace{},
	); err != nil {
		return nil, err
	}

	base, err := components.NewBaseConnector(providers.Atlassian, params)
	if err != nil {
		return nil, err
	}

	connector := &Connector{
		BaseConnector: *base,
		workspace:     params.Workspace,
	}

	switch base.Module() { // nolint:gocritic
	case providers.ModuleAtlassianJira:
		connector.jira, err = jira.NewAdapter(base)
		if err != nil {
			return nil, err
		}
	case providers.ModuleAtlassianConfluence:
		connector.confluence, err = confluence.NewAdapter(base)
		if err != nil {
			return nil, err
		}
	}

	return connector, nil
}

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*connectors.ListObjectMetadataResult, error) {
	if c.jira != nil {
		return c.jira.ListObjectMetadata(ctx, objectNames)
	}

	if c.confluence != nil {
		return c.confluence.ListObjectMetadata(ctx, objectNames)
	}

	return nil, common.ErrNotImplemented
}

func (c *Connector) Read(ctx context.Context, params connectors.ReadParams) (*connectors.ReadResult, error) {
	if c.jira != nil {
		return c.jira.Read(ctx, params)
	}

	if c.confluence != nil {
		return c.confluence.Read(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c *Connector) Write(ctx context.Context, params connectors.WriteParams) (*connectors.WriteResult, error) {
	if c.jira != nil {
		return c.jira.Write(ctx, params)
	}

	if c.confluence != nil {
		return c.confluence.Write(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c *Connector) Delete(ctx context.Context, params connectors.DeleteParams) (*connectors.DeleteResult, error) {
	if c.jira != nil {
		return c.jira.Delete(ctx, params)
	}

	if c.confluence != nil {
		return c.confluence.Delete(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

// URL allows to get list of sites associated with auth token.
// https://developer.atlassian.com/cloud/confluence/oauth-2-3lo-apps/#3-1-get-the-cloudid-for-your-site
func (c *Connector) getAccessibleSitesURL() (*urlbuilder.URL, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(url.Origin(), "oauth/token/accessible-resources")
}
