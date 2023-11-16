package salesforce

import "github.com/amp-labs/connectors/common"

// HTTPClient returns the underlying JSON HTTP client.
func (c *Connector) HTTPClient() *common.HTTPClient {
	return c.Client
}
