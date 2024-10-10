package atlassian

import (
	"errors"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

var _ deep.URLResolver = customURLBuilder{}

func newURLBuilder(
	data *deep.ConnectorData[parameters, *AuthMetadataVars],
	clients *deep.Clients,
) *customURLBuilder {
	return &customURLBuilder{
		data:    data,
		clients: clients,
	}
}

type customURLBuilder struct {
	data    *deep.ConnectorData[parameters, *AuthMetadataVars]
	clients *deep.Clients
}

func (f customURLBuilder) FindURL(method deep.Method, baseURL, objectName string) (*urlbuilder.URL, error) {
	switch method {
	case deep.ReadMethod:
		return f.getJiraRestApiURL("search")
	case deep.CreateMethod:
		fallthrough
	case deep.UpdateMethod:
		return f.getJiraRestApiURL("issue")
	case deep.DeleteMethod:
		return f.getJiraRestApiURL("issue")
	}

	// TODO should be a general error handled by `deep` package
	return nil, errors.New("URL cannot be resolved")
}

func (f customURLBuilder) getJiraRestApiURL(arg string) (*urlbuilder.URL, error) {
	cloudId, err := getCloudId(f.data.Metadata)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(f.clients.BaseURL(), "ex/jira", cloudId, f.data.Module, arg)
}

func (f customURLBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "urlResolver",
		Constructor: newURLBuilder,
		Interface:   new(deep.URLResolver),
	}
}

func getCloudId(vars *AuthMetadataVars) (string, error) {
	if len(vars.CloudID) == 0 {
		return "", ErrMissingCloudId
	}

	return vars.CloudID, nil
}

// URL format follows structure applicable to Oauth2 Atlassian apps.
// https://developer.atlassian.com/cloud/jira/platform/rest/v2/intro/#other-integrations
func (c *Connector) getJiraRestApiURL(arg string) (*urlbuilder.URL, error) {
	cloudId, err := getCloudId(c.Data.Metadata)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.Clients.BaseURL(), "ex/jira", cloudId, c.Data.Module, arg)
}

// URL allows to get list of sites associated with auth token.
// https://developer.atlassian.com/cloud/confluence/oauth-2-3lo-apps/#3-1-get-the-cloudid-for-your-site
func (c *Connector) getAccessibleSitesURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(c.Clients.BaseURL(), "oauth/token/accessible-resources")
}
