package atlassian

import (
	"errors"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

var _ deep.URLResolver = URLBuilder{}

func newURLBuilder(
	data *deep.ConnectorData[parameters, *AuthMetadataVars],
	clients *deep.Clients,
) *URLBuilder {
	return &URLBuilder{
		data:    data,
		clients: clients,
	}
}

type URLBuilder struct {
	data    *deep.ConnectorData[parameters, *AuthMetadataVars]
	clients *deep.Clients
}

func (f URLBuilder) FindURL(method deep.Method, baseURL, objectName string) (*urlbuilder.URL, error) {
	switch method {
	case deep.ReadMethod:
		return f.getJiraRestApiURL("search")
	case deep.DeleteMethod:
		return f.getJiraRestApiURL("issue")
	}

	// TODO should be a general error handled by `deep` package
	return nil, errors.New("URL cannot be resolved")
}

func (f URLBuilder) getJiraRestApiURL(arg string) (*urlbuilder.URL, error) {
	cloudId, err := getCloudId(f.data.Metadata)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(f.clients.BaseURL(), "ex/jira", cloudId, f.data.Module, arg)
}

func (f URLBuilder) Satisfies() requirements.Dependency {
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
