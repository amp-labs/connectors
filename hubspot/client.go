package hubspot

import "github.com/amp-labs/connectors/common"

// JSONHTTPClient returns the underlying JSON HTTP client.
func (c *Connector) JSONHTTPClient() *common.JSONHTTPClient {
	return c.Client
}
