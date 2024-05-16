package connector

import (
	"net/http"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) interpretError(res *http.Response, body []byte) error {
	return common.InterpretError(res, body)
}
