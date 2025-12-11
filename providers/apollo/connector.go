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

type operation string

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
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

// Provider returns the connector provider.
func (c *Connector) Provider() providers.Provider {
	return providers.Apollo
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}

// getAPIURL builds the url we can write/read data from
// Depending on the operation(read or write), some objects will need different endpoints.
// That's the sole purpose of the variable ops.
func (c *Connector) getAPIURL(objectName string, ops operation) (*urlbuilder.URL, error) {
	objectName = constructSupportedObjectName(objectName)

	relativePath := strings.Join([]string{restAPIPrefix, objectName}, "/")

	url, err := urlbuilder.New(c.BaseURL, relativePath)
	if err != nil {
		return nil, err
	}

	// If the given object uses search endpoint for Reading,
	// checks for the  method and makes the call.
	// currently we do not support routing to Search method.
	//
	if usesSearching(objectName) && ops == readOp {
		url.AddPath(searchingPath)
	}

	return url, nil
}
