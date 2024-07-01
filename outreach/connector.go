package outreach

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const (
	apiVersion = "api/v2"
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

	params, err := parameters{}.FromOptions(opts...)
	if err != nil {
		return nil, err
	}

	// Read provider info
	providerInfo, err := providers.ReadInfo(providers.Outreach)
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: params.Client.Caller,
		},
	}

	conn.setBaseURL(providerInfo.BaseURL)

	return conn, nil
}

func (c *Connector) getURL(arg string) (*urlbuilder.URL, error) {
	parts := []string{c.BaseURL, apiVersion, arg}
	filtered := make([]string, 0)

	for _, part := range parts {
		if len(part) != 0 {
			filtered = append(filtered, part)
		}
	}

	return constructURL(strings.Join(filtered, "/"))
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}
