package atlassian

import (
	"errors"
	"fmt"
	"strings"

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
	Client     *common.JSONHTTPClient
	moduleInfo *providers.ModuleInfo
	moduleID   common.ModuleID

	// workspace is used to find cloud ID.
	workspace string
	cloudId   string

	*providers.ProviderInfo
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
	}

	// Convert metadata map to model.
	authMetadata := NewAuthMetadataVars(params.Metadata.Map)
	conn.cloudId = authMetadata.CloudId

	if err := conn.setProviderInfo(); err != nil {
		return nil, err
	}

	module := conn.ProviderInfo.ReadModule(params.Module.Selection.ID)
	conn.moduleID, conn.moduleInfo = module.ID, &module.ModuleInfo

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
	return fmt.Sprintf("%s.Connector[%s]", c.Provider(), c.moduleID)
}

// URL format follows structure applicable to Oauth2 Atlassian apps.
// https://developer.atlassian.com/cloud/jira/platform/rest/v2/intro/#other-integrations
func (c *Connector) getJiraRestApiURL(arg string) (*urlbuilder.URL, error) {
	// In the case of JIRA / Atlassian Cloud, we use this path. In other cases, we fall back to the base path.
	modulePath := supportedModules[c.moduleID].Path()

	if c.moduleID == providers.ModuleAtlassianJira {
		cloudId, err := c.getCloudId()
		if err != nil {
			return nil, err
		}

		return urlbuilder.New(c.BaseURL, "ex/jira", cloudId, modulePath, arg)
	}

	return urlbuilder.New(c.BaseURL, modulePath, arg)
}

func (c *Connector) setBaseURL(newURL string) {
	// This is a temporary fix. And will be addressed when URLs are properly loaded from ProviderInfo.
	if strings.Contains(newURL, "ex/jira") {
		newURL = "https://api.atlassian.com"
	}

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
	if c.moduleID == providers.ModuleAtlassianJiraConnect {
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
