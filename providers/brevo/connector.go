package brevo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
)

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

	providerInfo, err := providers.ReadInfo(providers.Brevo)

	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: params.Caller,
		},
		Module: params.Selection,
	}

	conn.setBaseURL(providerInfo.BaseURL)

	return conn, nil

}

func (c *Connector) Provider() providers.Provider {
	return providers.Brevo
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}
