package components

import (
	"github.com/amp-labs/connectors/common"
)

// Setters for the Connector.
func (c *Connector) SetErrorHandler(h common.ErrorHandler) {
	c.Transport.json.HTTPClient.ErrorHandler = h
}

func (c *Connector) SetResponseHandler(h common.ResponseHandler) {
	c.Transport.json.HTTPClient.ResponseHandler = h
}
