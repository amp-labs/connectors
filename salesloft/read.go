package salesloft

import (
	"context"
	"strconv"
	"time"

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
		makeNextRecordsURL(link),
		getMarshaledData,
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		return urlbuilder.New(config.NextPage.String())
	}

	// First page
	link, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	link.WithQueryParam("per_page", strconv.Itoa(DefaultPageSize))

	if !config.Since.IsZero() {
		// Documentation states ISO8601, while server accepts different formats
		// but for consistency we are sticking to one format to be sent.
		// For the reference any API resource that includes time data type mentions iso8601 string format.
		// One example, say accounts is https://developers.salesloft.com/docs/api/accounts-index
		updatedSince := config.Since.Format(time.RFC3339Nano)
		link.WithQueryParam("updated_at[gte]", updatedSince)
	}

	return link, nil
}
