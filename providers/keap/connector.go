package keap

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/keap/metadata"
)

const (
	ApiPathPrefix = "crm/rest"
	Version1      = "v1"
	Version2      = "v2"
)

// Connector implements API for V1 and V2:
// https://developer.keap.com/docs/rest/
// https://developer.keap.com/docs/restv2/
type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
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

	// Read provider info
	providerInfo, err := providers.ReadInfo(conn.Provider())
	if err != nil {
		return nil, err
	}

	// connector and its client must mirror base url and provide its own error parser
	conn.setBaseURL(providerInfo.BaseURL)
	conn.Client.HTTPClient.ErrorHandler = interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
		HTML: &interpreter.DirectFaultyResponder{Callback: conn.interpretHTMLError},
	}.Handle

	return conn, nil
}

func (c *Connector) Provider() providers.Provider {
	return providers.Keap
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.FindURLPath(common.ModuleRoot, objectName)
	if err != nil {
		return nil, err
	}

	// Path already includes Version from the schema.json.

	return c.getURL(path)
}

func (c *Connector) getWriteURL(objectName string) (*urlbuilder.URL, error) {
	return c.getURL(Version2, objectName)
}

func (c *Connector) getModelURL(objectName string) (*urlbuilder.URL, error) {
	return c.getURL(Version2, objectName, "model")
}

func (c *Connector) getURL(args ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, append([]string{
		ApiPathPrefix,
	}, args...)...)
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}
