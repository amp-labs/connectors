package marketo

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// Read retrieves data based on the provided common.ReadParams configuration parameters.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.getURL(config)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res,
		getRecords,
		getNextRecordsURL,
		common.GetMarshaledData,
		config.Fields,
	)
}
