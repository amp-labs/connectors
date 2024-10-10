package atlassian

import (
	"errors"
	"fmt"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers"
	"github.com/spyzhov/ajson"
	"time"
)

// ErrMissingCloudId happens when cloud id was not provided via WithMetadata.
var ErrMissingCloudId = errors.New("connector missing cloud id")

type Connector struct {
	Data deep.ConnectorData[parameters, *AuthMetadataVars]
	deep.Clients
	deep.EmptyCloser
	deep.Reader
	deep.Remover
}

func NewConnector(opts ...Option) (*Connector, error) {
	constructor := func(
		clients *deep.Clients,
		closer *deep.EmptyCloser,
		data *deep.ConnectorData[parameters, *AuthMetadataVars],
		reader *deep.Reader,
		remover *deep.Remover) *Connector {
		return &Connector{
			Data:        *data,
			Clients:     *clients,
			EmptyCloser: *closer,
			Reader:      *reader,
			Remover:     *remover,
		}
	}
	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}
	firstPage := deep.FirstPageBuilder{
		Build: func(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error) {
			if !config.Since.IsZero() {
				// Read URL supports time scoping. common.ReadParams.Since is used to get relative time frame.
				// Here is an API example on how to request issues that were updated in the last 30 minutes.
				// search?jql=updated > "-30m"
				// The reason we use minutes is that it is the most granular API permits.
				diff := time.Since(config.Since)

				minutes := int64(diff.Minutes())
				if minutes > 0 {
					url.WithQueryParam("jql", fmt.Sprintf(`updated > "-%vm"`, minutes))
				}
			}

			return url, nil
		},
	}
	nextPage := deep.NextPageBuilder{
		Build: func(config common.ReadParams, previousPage *urlbuilder.URL, node *ajson.Node) (*urlbuilder.URL, error) {
			startAt, err := getNextRecords(node)
			if err != nil {
				return nil, err
			}

			if len(startAt) != 0 {
				previousPage.WithQueryParam("startAt", startAt)

				return previousPage, nil
			}

			return nil, nil
		},
	}
	readObjectLocator := deep.ReadObjectLocator{
		Locate: func(config common.ReadParams) string {
			return "issues"
		},
		FlattenRecords: flattenRecords,
	}

	return deep.ExtendedConnector[Connector, parameters, *AuthMetadataVars](
		constructor, providers.Atlassian, &AuthMetadataVars{}, opts,
		errorHandler,
		URLBuilder{},
		firstPage,
		nextPage,
		readObjectLocator,
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
