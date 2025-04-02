package atlassian

import (
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

// ErrMissingCloudId happens when cloud id was not provided via WithMetadata.
var ErrMissingCloudId = errors.New("connector missing cloud id")

type Connector struct {
	// Basic connector
	*components.Connector

	// workspace is used to find cloud ID.
	workspace string
	cloudId   string
}

// NewConnector is an old constructor, use NewConnectorV2.
// Deprecated.
func NewConnector(opts ...Option) (*Connector, error) {
	params, err := newParams(opts)
	if err != nil {
		return nil, err
	}

	return NewConnectorV2(*params)
}

func NewConnectorV2(params common.Parameters) (*Connector, error) {
	conn, err := components.Initialize(providers.Atlassian, params, constructor)
	if err != nil {
		return nil, err
	}

	conn.workspace = params.Workspace

	// Convert metadata map to model.
	authMetadata := NewAuthMetadataVars(params.Metadata)
	conn.cloudId = authMetadata.CloudId

	return conn, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	base.SetErrorHandler(interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle)

	return &Connector{Connector: base}, nil
}

// URL format follows structure applicable to Oauth2 Atlassian apps.
// https://developer.atlassian.com/cloud/jira/platform/rest/v2/intro/#other-integrations
func (c *Connector) getJiraRestApiURL(objectName string) (*urlbuilder.URL, error) {
	cloudId, err := c.getCloudId()
	if err != nil {
		return nil, err
	}

	return c.ModuleClient.TemplateURL(map[string]string{
		cloudIdKey: cloudId,
	}, objectName)
}

// URL allows to get list of sites associated with auth token.
// https://developer.atlassian.com/cloud/confluence/oauth-2-3lo-apps/#3-1-get-the-cloudid-for-your-site
func (c *Connector) getAccessibleSitesURL() (*urlbuilder.URL, error) {
	return c.RootClient.URL("oauth/token/accessible-resources")
}

func (c *Connector) getCloudId() (string, error) {
	if len(c.cloudId) == 0 {
		return "", ErrMissingCloudId
	}

	return c.cloudId, nil
}
