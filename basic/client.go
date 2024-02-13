package basic

import (
	"github.com/amp-labs/connectors/common"
)

func (c *Connector) HTTPClient() *common.HTTPClient {
	return c.Client.HTTPClient
}

func (c *Connector) Close() error {
	return nil
}
