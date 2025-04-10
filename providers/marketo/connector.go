package marketo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
	Module  common.Module
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts,
		// The module is resolved on behalf of the user if the option is missing.
		WithModule(providers.ModuleMarketoLeads),
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
		Module: params.Module.Selection,
	}

	conn.setBaseURL(providerInfo.BaseURL)

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
	objName = common.AddSuffixIfNotExists(objName, ".json")

	return urlbuilder.New(c.BaseURL, restAPIPrefix, c.Module.Path(), objName)
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}
