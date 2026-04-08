package salesloft

import (
	"context"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(resp,
		common.MakeRecordsFunc("data"),
		makeNextRecordsURL(url),
		common.MakeMarshaledDataFunc(flattenCustomEmbed),
		params.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		return urlbuilder.New(config.NextPage.String())
	}

	// First page
	url, err := c.getObjectURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("per_page", strconv.Itoa(DefaultPageSize))

	// Always use cursor-based polling as recommended by Salesloft for efficient data retrieval.
	// Results are sorted by updated_at ascending so we can use the last record's timestamp
	// as the cursor for the next request, avoiding deep pagination (page 500+) which causes
	// rate limit cost escalation and server errors.
	// See: https://developers.salesloft.com/docs/platform/guides/building-an-efficient-cursor-poller/
	url.WithQueryParam("sort_direction", "ASC")

	if !config.Since.IsZero() {
		updatedSince := config.Since.Format(time.RFC3339Nano)
		url.WithQueryParam("updated_at[gte]", updatedSince)
	}

	return url, nil
}
