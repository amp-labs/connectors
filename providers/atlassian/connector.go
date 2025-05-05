package atlassian

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

// ErrMissingCloudId happens when cloud id was not provided via WithMetadata.
var ErrMissingCloudId = errors.New("connector missing cloud id")

type Connector struct {
	Client     *common.JSONHTTPClient
	moduleInfo providers.ModuleInfo
	moduleID   common.ModuleID

	// workspace is used to find cloud ID.
	workspace string
	cloudId   string

	*providers.ProviderInfo
	*components.URLManager
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts,
		WithModule(common.ModuleRoot), // The module is resolved on behalf of the user if the option is missing.
	)
	if err != nil {
		return nil, err
	}

	httpClient := params.Client.Caller
	httpClient.ErrorHandler = interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle

	// Convert metadata map to model.
	authMetadata := NewAuthMetadataVars(params.Metadata.Map)

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: httpClient,
		},
		workspace: params.Workspace.Name,
		moduleID:  params.Module.Selection.ID,
		cloudId:   authMetadata.CloudId,
	}

	conn.ProviderInfo, err = providers.ReadInfo(conn.Provider(), params.GetCatalogVars()...)
	if err != nil {
		return nil, err
	}

	conn.moduleInfo, err = conn.ProviderInfo.ReadModuleInfo(conn.moduleID)
	if err != nil {
		return nil, err
	}

	conn.URLManager = components.NewURLManager(conn.ProviderInfo, conn.moduleInfo)

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
	if c.moduleID == providers.ModuleAtlassianJira {
		cloudId, err := c.getCloudId()
		if err != nil {
			return nil, err
		}

		return c.ModuleAPI.DynamicURL(map[string]string{
			"cloudId": cloudId,
		}, arg)
	}

	return c.ModuleAPI.URL(arg)
}

// URL allows to get list of sites associated with auth token.
// https://developer.atlassian.com/cloud/confluence/oauth-2-3lo-apps/#3-1-get-the-cloudid-for-your-site
func (c *Connector) getAccessibleSitesURL() (*urlbuilder.URL, error) {
	return c.RootAPI.URL("oauth/token/accessible-resources")
}

func (c *Connector) getCloudId() (string, error) {
	if len(c.cloudId) == 0 {
		return "", ErrMissingCloudId
	}

	return c.cloudId, nil
}
