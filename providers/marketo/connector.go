package marketo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	BaseURL    string
	Client     *common.JSONHTTPClient
	moduleInfo *providers.ModuleInfo
	moduleID   common.ModuleID
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts,
		// The module is resolved on behalf of the user if the option is missing.
		WithModule(common.ModuleRoot),
	)
	if err != nil {
		return nil, err
	}

	// Read marketo provider's details.
	providerInfo, err := providers.ReadInfo(providers.Marketo, &params.Workspace)
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client:          params.Caller.Client,
				ResponseHandler: responseHandler,
			},
		},
		moduleID: params.Module.Selection.ID,
	}

	conn.setBaseURL(providerInfo.BaseURL)

	conn.moduleInfo = providerInfo.ReadModuleInfo(conn.moduleID)

	return conn, nil
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}

// Provider returns the connector provider.
func (c *Connector) Provider() providers.Provider {
	return providers.Marketo
}

func (c *Connector) getAPIURL(objName string) (*urlbuilder.URL, error) {
	modulePath := supportedModules[c.moduleID].Path()
	objName = common.AddSuffixIfNotExists(objName, ".json")

	return urlbuilder.New(c.BaseURL, restAPIPrefix, modulePath, objName)
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}
