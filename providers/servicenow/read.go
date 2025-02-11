package servicenow

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildURL(config)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Get(ctx, url)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(resp,
		common.GetRecordsUnderJSONPath("result"),
		getNextRecordsURL(resp.Headers.Get("Link")),
		common.GetMarshaledData,
		config.Fields,
	)
}
