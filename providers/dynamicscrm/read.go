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
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	// always include annotations header
	// response will describe enums, foreign relationship, etc.
	rsp, err := c.JSONHTTPClient().Get(ctx, url.String(), newPaginationHeader(DefaultPageSize), common.Header{
		Key:   "Prefer",
		Value: `odata.include-annotations="*"`,
	})
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		getRecords,
		getNextRecordsURL,
		common.GetMarshaledData,
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		return constructURL(config.NextPage.String())
	}

	// First page
	url, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	fields := config.Fields.List()
	if len(fields) != 0 {
		url.WithQueryParam("$select", strings.Join(fields, ","))
	}

	return url, nil
}

func newPaginationHeader(pageSize int) common.Header {
	return common.Header{
		Key:   "Prefer",
		Value: fmt.Sprintf("odata.maxpagesize=%v", pageSize),
	}
}
