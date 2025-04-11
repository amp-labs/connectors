package atlassian

import (
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
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

	// TODO this is a temporary fix. Module based URL support will  resolve this.
	if err := conn.overwriteProviderInfo(); err != nil {
		return nil, err
	}

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
func (c *Connector) getJiraRestApiURL(arg string) (*urlbuilder.URL, error) {
	modulePath := supportedModules[c.Module()].Path()

	// In the case of JIRA / Atlassian Cloud, we use this path. In other cases, we fall back to the base path.
	if c.Module() == providers.ModuleAtlassianJira {
		cloudId, err := c.getCloudId()
		if err != nil {
			return nil, err
		}

		return urlbuilder.New(c.ProviderInfo().BaseURL, "ex/jira", cloudId, modulePath, arg)
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, modulePath, arg)
}

// URL allows to get list of sites associated with auth token.
// https://developer.atlassian.com/cloud/confluence/oauth-2-3lo-apps/#3-1-get-the-cloudid-for-your-site
func (c *Connector) getAccessibleSitesURL() (*urlbuilder.URL, error) {
	// TODO what URL to choose??? Root module?
	return urlbuilder.New(c.ProviderInfo().BaseURL, "oauth/token/accessible-resources")
}

func (c *Connector) getCloudId() (string, error) {
	if len(c.cloudId) == 0 {
		return "", ErrMissingCloudId
	}

	return c.cloudId, nil
}

func (c *Connector) overwriteProviderInfo() error {
	// When the module is Atlassian Connect, the base URL is different, so we need to override it.
	// TODO: Replace options with substitution map in the future to avoid having to know
	// which values need to be substituted.
	if c.Module() == providers.ModuleAtlassianJiraConnect {
		vars := []catalogreplacer.CatalogVariable{
			&paramsbuilder.Workspace{Name: c.workspace},
		}

		override := &providers.ProviderInfo{
			BaseURL: "https://{{.workspace}}.atlassian.net",
		}

		// Mutates the provider info with the overrides, and substitutes any variables.
		if err := c.ProviderInfo().Override(override).SubstituteWith(vars); err != nil {
			return err
		}
	}

	return nil
}
