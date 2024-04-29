package microsoftdynamicscrm

import (
	"context"
	"fmt"
	"strings"

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
		link = linkutils.NewURL(c.getURL(config.ObjectName))
		if len(config.Fields) != 0 {
			link.WithQueryParam("$select", strings.Join(config.Fields, ","))
		}
	} else {
		// Next page
		link = linkutils.NewURL(config.NextPage.String())
	}

	// always include annotations header
	// response will describe enums, foreign relationship, etc.
	rsp, err := c.get(ctx, link.String(), newPaginationHeader(DefaultPageSize), common.Header{
		Key:   "Prefer",
		Value: `odata.include-annotations="*"`,
	})
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

func newPaginationHeader(pageSize int) common.Header {
	return common.Header{
		Key:   "Prefer",
		Value: fmt.Sprintf("odata.maxpagesize=%v", pageSize),
	}
}
