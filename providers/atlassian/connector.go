package atlassian

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "3"

type Connector struct {
	Client *common.JSONHTTPClient

	// workspace is used to find cloud ID.
	workspace string

	providerInfo *providers.ProviderInfo
	moduleInfo   *providers.ModuleInfo
	moduleID     common.ModuleID
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts,
		WithModule(common.ModuleRoot), // The module is resolved on behalf of the user if the option is missing.
	)
	if err != nil {
		return nil, err
	}

	httpClient := params.Client.Caller
	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: httpClient,
		},
		workspace: params.Workspace.Name,
		moduleID:  params.Module.Selection.ID,
	}

	// Convert metadata map to model.
	authMetadata := NewAuthMetadataVars(params.Metadata.Map)

	conn.providerInfo, err = providers.ReadInfo(conn.Provider(), &params.Workspace, authMetadata)
	if err != nil {
		return nil, err
	}

	conn.moduleInfo = conn.providerInfo.ReadModuleInfo(conn.moduleID)

	// HTTPClient will soon not store Base URL.
	conn.Client.HTTPClient.Base = conn.providerInfo.BaseURL
	conn.Client.HTTPClient.ErrorHandler = interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
		HTML: &interpreter.DirectFaultyResponder{Callback: conn.interpretHTMLError},
	}.Handle

	return conn, nil
}

func (c *Connector) Provider() providers.Provider {
	return providers.Atlassian
}

func (c *Connector) String() string {
	return fmt.Sprintf("%s.Connector[%s]", c.Provider(), c.moduleID)
}

// This method must be used only by the unit tests.
func (c *Connector) setBaseURL(rootURL, moduleURL string) {
	c.providerInfo.BaseURL = rootURL
	c.moduleInfo.BaseURL = moduleURL
}

// URL allows to get list of sites associated with auth token.
// https://developer.atlassian.com/cloud/confluence/oauth-2-3lo-apps/#3-1-get-the-cloudid-for-your-site
func (c *Connector) getAccessibleSitesURL() (*urlbuilder.URL, error) {
	url, err := urlbuilder.New(c.providerInfo.BaseURL)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(url.Origin(), "oauth/token/accessible-resources")
}

// URL format for providers.ModuleAtlassianJira follows structure applicable to Oauth2 Atlassian apps:
// https://developer.atlassian.com/cloud/jira/platform/rest/v2/intro/#other-integrations
func (c *Connector) getModuleURL(path ...string) (*urlbuilder.URL, error) {
	path = append([]string{apiVersion}, path...)

	return urlbuilder.New(c.moduleInfo.BaseURL, path...)
}
