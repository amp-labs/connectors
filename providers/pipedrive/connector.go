package pipedrive

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const (
	apiVersion string = "v1"    // nolint:gochecknoglobals
	limitQuery string = "limit" // nolint:gochecknoglobals
)

// Connector represents the Pipedrive Connector.
type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
}

// NewConnector constructs the Pipedriver Connector and returns it, Fails
// if any of the require fields are not instantiated.
func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	defer common.PanicRecovery(func(cause error) {
		outErr = cause
		conn = nil
	})

	params, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: params.Client.Caller,
		},
	}

	providerInfo, err := providers.ReadInfo(conn.Provider())
	if err != nil {
		return nil, err
	}

	conn.setBaseURL(providerInfo.BaseURL)

	return conn, nil
}

// Provider returns the pipedrive provider instance.
func (c *Connector) Provider() providers.Provider {
	return providers.Pipedrive
}

// String implements the fmt.Stringer interface.
func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}

func (c *Connector) getApiURL(arg string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, apiVersion, arg)
}
