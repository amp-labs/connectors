package generic

import (
	"github.com/amp-labs/connectors/common"
)

func (c *Connector) JSONHTTPClient() *common.JSONHTTPClient {
	return c.Client
}

func (c *Connector) HTTPClient() *common.HTTPClient {
	return c.Client.HTTPClient
}
