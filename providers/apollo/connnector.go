package apollo

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	defer common.PanicRecovery(func(cause error) {
		outErr = cause
		conn = nil
	})

	params, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	providerInfo, err := providers.ReadInfo(providers.Apollo)
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

// Provider returns the connector provider.
func (c *Connector) Provider() providers.Provider {
	return providers.Apollo
}

func (c *Connector) getAPIURL(objName string) (*urlbuilder.URL, error) {
	relativePath := strings.Join([]string{restAPIPrefix, objName}, "/")

	url, err := urlbuilder.New(c.BaseURL, relativePath)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("per_page", pageSize)

	return url, nil
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}