package pipeliner

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	link, err := c.buildReadURL(config)
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
		getMarshaledData,
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	link, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	link.WithQueryParam("first", strconv.Itoa(DefaultPageSize))

	if len(config.NextPage) != 0 {
		// Next page
		link.WithQueryParam("after", config.NextPage.String())
	}

	return link, nil
}
