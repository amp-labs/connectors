package heyreach

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const (
	apiVersion = "public"
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

	providerInfo, err := providers.ReadInfo(providers.HeyReach)
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client: params.Caller.Client,
			},
		},
	}

	conn.setBaseURL(providerInfo.BaseURL)

	return conn, nil
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}

func (c *Connector) getAPIURL(objectName string) (*urlbuilder.URL, error) {
	url, err := urlbuilder.New(c.BaseURL, apiVersion, objectName)
	if err != nil {
		return nil, err
	}

	return url, nil
}

// Provider returns the connector provider.
func (c *Connector) Provider() providers.Provider {
	return providers.HeyReach
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}
