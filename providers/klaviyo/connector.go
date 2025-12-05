package klaviyo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/klaviyo/metadata"
)

// headerVersion2024Oct15 is the latest stable version of API as of the date of writing.
// https://developers.klaviyo.com/en/reference/api_overview
const headerVersion2024Oct15 = "2024-10-15"

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
		Custom: map[string]interpreter.FaultyResponseHandler{
			"application/vnd.api+json": interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
		},
	}.Handle

	return conn, nil
}

func (c *Connector) Provider() providers.Provider {
	return providers.Klaviyo
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.LookupURLPath(common.ModuleRoot, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.BaseURL, path)
}

func (c *Connector) getWriteURL(objectName string) (*urlbuilder.URL, error) {
	path := objectNameToWritePath.Get(objectName)

	return urlbuilder.New(c.BaseURL, path)
}

func (c *Connector) getDeleteURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, "api", objectName)
}

func (c *Connector) revisionHeader() common.Header {
	return common.Header{
		Key:   "revision",
		Value: headerVersion2024Oct15,
	}
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}
