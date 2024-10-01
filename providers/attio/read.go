// nolint
package attio

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams, QueryParam map[string]string) (*common.ReadResult, error) {
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

	for key, val := range QueryParam {
		if val != "" {
			url.WithQueryParam(key, val)
		}
	}

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
