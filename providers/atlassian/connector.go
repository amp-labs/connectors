package atlassian

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/atlassian/internal/jira"
)

type Connector struct {
	// Basic connector.
	*components.Connector

	// Require params.
	common.RequireAuthenticatedClient
	common.RequireWorkspace

	Jira *jira.Adapter

	workspace string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.Atlassian, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
		workspace: params.Workspace,
	}

	if connector.Module() == providers.ModuleAtlassianJira {
		adapter, err := jira.NewAdapter(params)
		if err != nil {
			return nil, err
		}

		connector.Jira = adapter
	}

	return connector, nil
}

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*connectors.ListObjectMetadataResult, error) {
	if c.Jira != nil {
		return c.Jira.ListObjectMetadata(ctx, objectNames)
	}

	return nil, common.ErrNotImplemented
}

func (c *Connector) Read(ctx context.Context, params connectors.ReadParams) (*connectors.ReadResult, error) {
	if c.Jira != nil {
		return c.Jira.Read(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c *Connector) Write(ctx context.Context, params connectors.WriteParams) (*connectors.WriteResult, error) {
	if c.Jira != nil {
		return c.Jira.Write(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c *Connector) Delete(ctx context.Context, params connectors.DeleteParams) (*connectors.DeleteResult, error) {
	if c.Jira != nil {
		return c.Jira.Delete(ctx, params)
	}

	return nil, common.ErrNotImplemented
}

func (c *Connector) SetUnitTestBaseURL(url string) {
	// Module independent functionality.
	c.Connector.SetUnitTestBaseURL(url)

	// Replace base for respective module.
	if c.Jira != nil {
		c.Jira.SetUnitTestBaseURL(url)
	}
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
