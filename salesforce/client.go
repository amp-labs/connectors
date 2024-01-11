package salesforce

import "github.com/amp-labs/connectors/common"

// JSONHTTPClient returns the underlying JSON HTTP client.
func (c *Connector) JSONHTTPClient() *common.JSONHTTPClient {
	return c.Client
}

func (c *Connector) HTTPClient() *common.HTTPClient {
	return c.Client.HTTPClient
}
