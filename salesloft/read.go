package salesloft

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/linkutils"
)

// nolint:lll
// Microsoft API supports other capabilities like filtering, grouping, and sorting which we can potentially tap into later.
// See https://learn.microsoft.com/en-us/power-apps/developer/data-platform/webapi/query-data-web-api#odata-query-options
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	var link *linkutils.URL

	if len(config.NextPage) == 0 {
		// First page
		link = c.getURL(config.ObjectName)
		link.WithQueryParam("per_page", strconv.Itoa(DefaultPageSize))
	} else {
		// Next page
		link = linkutils.NewURL(config.NextPage.String())
	}

	rsp, err := c.get(ctx, link.String())
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
