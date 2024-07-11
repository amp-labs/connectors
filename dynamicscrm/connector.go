package dynamicscrm

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "v9.2"

type Connector struct {
	BaseURL   string
	Client    *common.JSONHTTPClient
	XMLClient *common.XMLHTTPClient
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

	httpClient := params.Client.Caller
	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: httpClient,
		},
		XMLClient: &common.XMLHTTPClient{
			HTTPClient: httpClient,
		},
	}

	providerInfo, err := providers.ReadInfo(conn.Provider(), &params.Workspace)
	if err != nil {
		return nil, err
	}

	// connector and its client must mirror base url and provide its own error parser
	conn.setBaseURL(providerInfo.BaseURL)
	conn.Client.HTTPClient.ErrorHandler = interpreter.ErrorHandler{
		JSON: conn.interpretJSONError,
	}.Handle

	return conn, nil
}

func (c *Connector) Provider() providers.Provider {
	return providers.DynamicsCRM
}

func (c *Connector) String() string {
	return fmt.Sprintf("%s.Connector", c.Provider())
}

func (c *Connector) getURL(arg string) (*urlbuilder.URL, error) {
	return constructURL(c.BaseURL, apiVersion, arg)
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}
