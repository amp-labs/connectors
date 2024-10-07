package deep

import (
	"fmt"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/providers"
)

type Clients struct {
	provider   providers.Provider
	httpClient *common.HTTPClient
	JSON       *common.JSONHTTPClient
	XML        *common.XMLHTTPClient
}

func (c *Clients) CopyFrom(other *Clients) {
	c.provider = other.provider
	c.httpClient = other.httpClient
	c.JSON = other.JSON
	c.XML = other.XML
}

// Provider returns the connector provider.
func (c *Clients) Provider() providers.Provider {
	return c.provider
}

// String returns a string representation of the connector, which is useful for logging / debugging.
func (c *Clients) String() string {
	return fmt.Sprintf("%s.Connector", c.Provider())
}

// JSONHTTPClient returns the underlying JSON HTTP client.
func (c *Clients) JSONHTTPClient() *common.JSONHTTPClient {
	return c.JSON
}

func (c *Clients) HTTPClient() *common.HTTPClient {
	return c.httpClient
}

func internalNewClients(provider providers.Provider, parameters any) (*Clients, error) {
	httpClient, err := ExtractHTTPClient(parameters)
	if err != nil {
		return nil, err
	}

	catalogsVars, err := ExtractCatalogVariables(parameters)
	if err != nil {
		return nil, err
	}

	providerInfo, err := providers.ReadInfo(provider, catalogsVars...)
	if err != nil {
		return nil, err
	}

	httpClient.Base = providerInfo.BaseURL

	return &Clients{
		provider:   provider,
		httpClient: httpClient,
		JSON: &common.JSONHTTPClient{
			HTTPClient: httpClient,
		},
		XML: &common.XMLHTTPClient{
			HTTPClient: httpClient,
		},
	}, nil
}

func (c *Clients) BaseURL() string {
	return c.httpClient.Base
}

func (c *Clients) WithBaseURL(newURL string) {
	c.httpClient.Base = newURL
}

func (c *Clients) WithErrorHandler(handler interpreter.ErrorHandler) {
	c.httpClient.ErrorHandler = handler.Handle
}
