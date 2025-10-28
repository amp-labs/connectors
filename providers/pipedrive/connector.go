package pipedrive

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/pipedrive/internal/crm"
	"github.com/amp-labs/connectors/providers/pipedrive/metadata"
)

const (
	apiVersion string = "v1" // nolint:gochecknoglobals
)

// Connector represents the Pipedrive Connector.
type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
	Module  common.Module

	crmAdapter *crm.Adapter // embedded for v2 functionality
}

// NewConnector constructs the Pipedrive Connector and returns it, Fails
// if any of the required fields are not instantiated.
func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts,
		WithModule(providers.ModulePipedriveLegacy),
	)
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

	conn.Module = params.Selection

	if conn.Module.ID == providers.ModulePipedriveCRM {
		conn.crmAdapter = &crm.Adapter{
			Client:  conn.Client,
			BaseURL: providerInfo.BaseURL,
		}
	}

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

// getAPIURL constructs a specific object's resource URL in the format
// `{{baseURL}}/{{version}}/{{objectName}}`.
func (c *Connector) getAPIURL(arg string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, apiVersion, arg)
}

func (c *Connector) constructMetadataURL(obj string) (*urlbuilder.URL, error) {
	if metadataDiscoveryEndpoints.Has(obj) {
		obj = metadataDiscoveryEndpoints[obj]
	}

	return c.getAPIURL(obj)
}

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.LookupURLPath(c.Module.ID, objectName)
	if err != nil {
		return nil, err
	}

	return c.getAPIURL(path)
}
