package outreach

import (
	"context"
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res, getTotalSize,
		getRecords,
		getNextRecordsURL,
		getMarshalledData,
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	// If NextPage is set, then we're reading the next page of results.
	// The NextPage URL has all the necessary parameters.
	if len(config.NextPage) > 0 {
		return constructURL(config.NextPage.String())
	}

	url, err := c.getApiURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	// If Since is not set, then we're doing a backfill. We read all rows (in pages)
	// Making sure the time passed, is in the format outreach expects.
	if !config.Since.IsZero() {
		time := config.Since.Format(time.DateOnly)
		fmtTime := fmt.Sprintf("%s..inf", time)
		url.WithQueryParam("filter[updatedAt]", fmtTime)
	}

	return url, nil
}
