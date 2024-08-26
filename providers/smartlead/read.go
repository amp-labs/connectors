package smartlead

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if len(config.ObjectName) == 0 {
		return nil, common.ErrMissingObjects
	}

	link, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	rsp, err := c.Client.Get(ctx, link.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		getTotalSize,
		getRecords,
		getNextRecordsURL,
		common.GetMarshaledData,
		config.Fields,
	)
}
