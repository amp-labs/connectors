// nolint
package attio

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

var dummyNextPageFunc = func(*ajson.Node) (string, error) {
	return "", nil
}

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedObjectsByRead.Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := c.getApiURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", "1")
	url.WithQueryParam("offset", "0")

	rsp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		common.GetRecordsUnderJSONPath("data"),
		dummyNextPageFunc,
		common.GetMarshaledData,
		config.Fields,
	)
}
