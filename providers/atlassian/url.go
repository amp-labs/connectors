package atlassian

import (
	"errors"

	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/dpobjects"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/internal/deep/dpvars"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

var _ dpobjects.URLResolver = customURLBuilder{}

func newURLBuilder(
	data *dpvars.ConnectorData[parameters, *AuthMetadataVars],
	clients *dprequests.Clients,
) *customURLBuilder {
	return &customURLBuilder{
		data:    data,
		clients: clients,
	}
}

type customURLBuilder struct {
	data    *dpvars.ConnectorData[parameters, *AuthMetadataVars]
	clients *dprequests.Clients
}

func (f customURLBuilder) FindURL(method dpobjects.Method, baseURL, objectName string) (*urlbuilder.URL, error) {
	switch method {
	case dpobjects.ReadMethod:
		return f.getJiraRestApiURL("search")
	case dpobjects.CreateMethod:
		fallthrough
	case dpobjects.UpdateMethod:
		return f.getJiraRestApiURL("issue")
	case dpobjects.DeleteMethod:
		return f.getJiraRestApiURL("issue")
	}

	// TODO should be a general error handled by `deep` package
	return nil, errors.New("URL cannot be resolved")
}

// URL format follows structure applicable to Oauth2 Atlassian apps.
// https://developer.atlassian.com/cloud/jira/platform/rest/v2/intro/#other-integrations
func (f customURLBuilder) getJiraRestApiURL(arg string) (*urlbuilder.URL, error) {
	cloudId, err := getCloudId(f.data.Metadata)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(f.clients.BaseURL(), "ex/jira", cloudId, f.data.Module, arg)
}

// URL allows to get list of sites associated with auth token.
// https://developer.atlassian.com/cloud/confluence/oauth-2-3lo-apps/#3-1-get-the-cloudid-for-your-site
func (f customURLBuilder) getAccessibleSitesURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(f.clients.BaseURL(), "oauth/token/accessible-resources")
}

func (f customURLBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.ObjectURLResolver,
		Constructor: newURLBuilder,
		Interface:   new(dpobjects.URLResolver),
	}
}

func getCloudId(vars *AuthMetadataVars) (string, error) {
	if len(vars.CloudID) == 0 {
		return "", ErrMissingCloudId
	}

	return vars.CloudID, nil
}
