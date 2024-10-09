package atlassian

import (
	"errors"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers"
)

// ErrMissingCloudId happens when cloud id was not provided via WithMetadata.
var ErrMissingCloudId = errors.New("connector missing cloud id")

type Connector struct {
	Data deep.ConnectorData[parameters, *AuthMetadataVars]
	deep.Clients
	deep.EmptyCloser
}

func NewConnector(opts ...Option) (*Connector, error) {
	constructor := func(
		clients *deep.Clients,
		closer *deep.EmptyCloser,
		data *deep.ConnectorData[parameters, *AuthMetadataVars]) *Connector {
		return &Connector{
			Clients:     *clients,
			EmptyCloser: *closer,
			Data:        *data,
		}
	}
	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}

	return deep.ExtendedConnector[Connector, parameters, *AuthMetadataVars](
		constructor, providers.Atlassian, &AuthMetadataVars{}, opts,
		errorHandler,
	)
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

func getCloudId(vars *AuthMetadataVars) (string, error) {
	if len(vars.CloudID) == 0 {
		return "", ErrMissingCloudId
	}

	return vars.CloudID, nil
}
