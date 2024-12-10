package atlassian

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

// ErrMissingCloudId happens when cloud id was not provided via WithMetadata.
var ErrMissingCloudId = errors.New("connector missing cloud id")

type Connector struct {
	Client *common.JSONHTTPClient
	Module common.Module

	// workspace is used to find cloud ID.
	workspace string
	cloudId   string

	*providers.ProviderInfo
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts,
		WithModule(ModuleEmpty), // The module is resolved on behalf of the user if the option is missing.
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
		Module:    params.Module.Selection,
	}

	// Convert metadata map to model.
	authMetadata := NewAuthMetadataVars(params.Metadata.Map)
	conn.cloudId = authMetadata.CloudId

	if err := conn.setProviderInfo(); err != nil {
		return nil, err
	}

	// connector and its client must mirror base url and provide its own error parser
	conn.setBaseURL(conn.BaseURL)
	conn.Client.HTTPClient.ErrorHandler = interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle

	return conn, nil
}

func (c *Connector) Provider() providers.Provider {
	return providers.Atlassian
}

func (c *Connector) String() string {
	return fmt.Sprintf("%s.Connector[%s]", c.Provider(), c.Module)
}

// URL format follows structure applicable to Oauth2 Atlassian apps.
// https://developer.atlassian.com/cloud/jira/platform/rest/v2/intro/#other-integrations
func (c *Connector) getJiraRestApiURL(arg string) (*urlbuilder.URL, error) {
	// In the case of JIRA / Atlassian Cloud, we use this path. In other cases, we fall back to the base path.
	if c.Module.ID == ModuleJira {
		cloudId, err := c.getCloudId()
		if err != nil {
			return nil, err
		}

		return urlbuilder.New(c.BaseURL, "ex/jira", cloudId, c.Module.Path(), arg)
	}

	return urlbuilder.New(c.BaseURL, c.Module.Path(), arg)
}

// URL allows to get list of sites associated with auth token.
// https://developer.atlassian.com/cloud/confluence/oauth-2-3lo-apps/#3-1-get-the-cloudid-for-your-site
func (c *Connector) getAccessibleSitesURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, "oauth/token/accessible-resources")
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}

func (c *Connector) getCloudId() (string, error) {
	if len(c.cloudId) == 0 {
		return "", ErrMissingCloudId
	}

	return c.cloudId, nil
}

func (c *Connector) setProviderInfo() error {
	// Read provider info
	providerInfo, err := providers.ReadInfo(c.Provider())
	if err != nil {
		return err
	}

	c.ProviderInfo = providerInfo

	// When the module is Atlassian Connect, the base URL is different, so we need to override it.
	// TODO: Replace options with substitution map in the future to avoid having to know
	// which values need to be substituted.
	if c.Module.ID == ModuleAtlassianJiraConnect {
		vars := []catalogreplacer.CatalogVariable{
			&paramsbuilder.Workspace{Name: c.workspace},
		}

		override := &providers.ProviderInfo{
			BaseURL: "https://{{.workspace}}.atlassian.net",
		}

		// Mutates the provider info with the overrides, and substitutes any variables.
		if err := c.ProviderInfo.Override(override).SubstituteWith(vars); err != nil {
			return err
		}
	}

	return nil
}
