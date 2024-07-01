package salesforce

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const (
	apiVersion    = "59.0"
	versionPrefix = "v"
	restAPISuffix = "/services/data/v59.0"
)

// Connector is a Salesforce connector.
type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
}

func APIVersion() string {
	return versionPrefix + apiVersion
}

func APIVersionSOAP() string {
	return apiVersion
}

// NewConnector returns a new Salesforce connector.
func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	defer common.PanicRecovery(func(cause error) {
		outErr = cause
		conn = nil
	})

	params, err := parameters{}.FromOptions(opts...)
	if err != nil {
		return nil, err
	}

	// Read provider info & replace catalog variables with given substitutions, if any
	providerInfo, err := providers.ReadInfo(providers.Salesforce, &params.Workspace)
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: params.Client.Caller,
		},
	}

	conn.setBaseURL(providerInfo.BaseURL)
	conn.Client.HTTPClient.ErrorHandler = conn.interpretError
	conn.Client.ErrorPostProcessor.Process = handleError

	return conn, nil
}

// Provider returns the connector provider.
func (c *Connector) Provider() providers.Provider {
	return providers.Salesforce
}

// String returns a string representation of the connector, which is useful for logging / debugging.
func (c *Connector) String() string {
	return fmt.Sprintf("%s.Connector", c.Provider())
}

func (c *Connector) getURL(paths ...string) (*urlbuilder.URL, error) {
	parts := append([]string{
		c.BaseURL,
		restAPISuffix, // scope URLs to API version
	}, paths...)

	return constructURL(strings.Join(parts, "/"))
}

func (c *Connector) getDomainURL(paths ...string) (*urlbuilder.URL, error) {
	parts := append([]string{
		c.BaseURL,
	}, paths...)

	return constructURL(strings.Join(parts, "/"))
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}
