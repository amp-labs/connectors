package asana

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "/1.0"

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
	Module  common.Module
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	httpClient := params.Client.Caller

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: httpClient,
		},
	}

	providerInfo, err := providers.ReadInfo(conn.Provider())
	if err != nil {
		return nil, err
	}

	conn.setBaseURL(providerInfo.BaseURL)

	return conn, nil
}

func (c *Connector) geAPIURL(object string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, apiVersion, object)
}

func (c *Connector) Provider() providers.Provider {
	return providers.Asana
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}
