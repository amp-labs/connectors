package intercom

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	link, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	rsp, err := c.get(ctx, link.String(), common.Header{
		Key:   "Intercom-Version",
		Value: apiVersion,
	})
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		getTotalSize,
		getRecords,
		makeNextRecordsURL(link),
		getMarshaledData,
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		return constructURL(config.NextPage.String())
	}

	// First page
	link, err := c.getURL(naming.NewPluralString(config.ObjectName).String())
	if err != nil {
		return nil, err
	}

	link.WithQueryParam("per_page", strconv.Itoa(DefaultPageSize))

	return link, nil
}
