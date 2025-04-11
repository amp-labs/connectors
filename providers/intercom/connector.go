package intercom

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "2.11"

var apiVersionHeader = common.Header{ // nolint:gochecknoglobals
	Key:   "Intercom-Version",
	Value: apiVersion,
}

type Connector struct {
	BaseURL string
	Module  common.Module
	Client  *common.JSONHTTPClient
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	providerInfo, err := providers.ReadInfo(providers.Intercom)
	if err != nil {
		return nil, err
	}

	httpClient := params.Client.Caller
	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: httpClient,
		},
	}
	// connector and its client must mirror base url and provide its own error parser
	conn.setBaseURL(providerInfo.BaseURL)
	conn.Client.HTTPClient.ErrorHandler = interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}.Handle

	return conn, nil
}

func (c *Connector) Provider() providers.Provider {
	return providers.Intercom
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

func (c *Connector) getURL(objectName string) (*urlbuilder.URL, error) {
	path := objectNameToURLPath.Get(objectName)

	return constructURL(c.BaseURL, path)
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}
