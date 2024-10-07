package marketo

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
	Module  string
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

	// Read marketo provider's details.
	providerInfo, err := providers.ReadInfo(providers.Marketo, &params.Workspace)
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client:          params.Caller.Client,
				ResponseHandler: responseHandler,
			},
		},
		Module: params.Module.Name,
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
	return providers.Marketo
}

func (c *Connector) getAPIURL(objName string) (*urlbuilder.URL, error) {
	objName = common.AddJSONSuffix(objName)
	bURL := strings.Join([]string{restAPIPrefix, c.Module, objName}, "/")

	return urlbuilder.New(c.BaseURL, bURL)
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}
