package dynamicscrm

import (
	"context"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// nolint:lll
// Microsoft API supports other capabilities like filtering, grouping, and sorting which we can potentially tap into later.
// See https://learn.microsoft.com/en-us/power-apps/developer/data-platform/webapi/query-data-web-api#odata-query-options
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	link, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	// always include annotations header
	// response will describe enums, foreign relationship, etc.
	rsp, err := c.Client.Get(ctx, link.String(), newPaginationHeader(DefaultPageSize), common.Header{
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
		getMarshalledData,
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		return constructURL(config.NextPage.String())
	}

	// First page
	link, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.Fields) != 0 {
		link.WithQueryParam("$select", strings.Join(config.Fields, ","))
	}

	return link, nil
}

func newPaginationHeader(pageSize int) common.Header {
	return common.Header{
		Key:   "Prefer",
		Value: fmt.Sprintf("odata.maxpagesize=%v", pageSize),
	}
}
