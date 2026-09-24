package atlassian

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "3"

type Connector struct {
	// Basic connector.
	*components.Connector

	// Require params.
	common.RequireAuthenticatedClient
	common.RequireWorkspace

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

	connector.SetErrorHandler(interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
		HTML: &interpreter.DirectFaultyResponder{Callback: connector.interpretHTMLError},
	}.Handle)

	return connector, nil
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

// URL format for providers.ModuleAtlassianJira follows structure applicable to Oauth2 Atlassian apps:
// https://developer.atlassian.com/cloud/jira/platform/rest/v2/intro/#other-integrations
func (c *Connector) getModuleURL(path ...string) (*urlbuilder.URL, error) {
	path = append([]string{apiVersion}, path...)

	return urlbuilder.New(c.ModuleInfo().BaseURL, path...)
}
