package dynamicscrm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
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
	rsp, err := c.Client.Get(ctx, url.String(), newPaginationHeader(DefaultPageSize), common.Header{
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

	if len(config.Fields) != 0 {
		url.WithQueryParam("$select", strings.Join(config.Fields, ","))
	}

	return url, nil
}

func newPaginationHeader(pageSize int) common.Header {
	return common.Header{
		Key:   "Prefer",
		Value: fmt.Sprintf("odata.maxpagesize=%v", pageSize),
	}
}

// Internal GET request, where we expect JSON payload.
func (c *Connector) performGetRequest(ctx context.Context, url *urlbuilder.URL) (*ajson.Node, error) {
	rsp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	body, ok := rsp.Body()
	if !ok {
		return nil, errors.Join(ErrObjectNotFound, common.ErrEmptyJSONHTTPResponse)
	}

	return body, nil
}
