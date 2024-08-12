package gong

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if len(config.ObjectName) == 0 {
		return nil, common.ErrMissingObjects
	}

	url, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.NextPage) != 0 { // not the first page, add a cursor
		url.WithQueryParam("cursor", config.NextPage.String())
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res,
		makeGetTotalSize(config.ObjectName),
		makeGetRecords(config.ObjectName),
		getNextRecordsURL,
		getMarshalledData,
		config.Fields,
	)
}
