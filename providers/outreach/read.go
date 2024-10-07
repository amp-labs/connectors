package outreach

import (
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// Read retrieves data based on the provided configuration parameters.
//
// This function executes a read operation using the given context and
// configuration parameters. It returns the nested Attributes values read results or an error
// if the operation fails.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	res, err := c.JSON.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res,
		getRecords,
		getNextRecordsURL,
		common.GetMarshaledData,
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	// If NextPage is set, then we're reading the next page of results.
	// The NextPage URL has all the necessary parameters.
	if len(config.NextPage) > 0 {
		return urlbuilder.New(config.NextPage.String())
	}

	url, err := c.getApiURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	// If Since is not set, then we're doing a backfill. We read all rows (in pages)
	// If Since is present, we turn it into the format the Outreach API expects
	if !config.Since.IsZero() {
		t := config.Since.Format(time.DateOnly)
		fmtTime := t + "..inf"
		url.WithQueryParam("filter[updatedAt]", fmtTime)
	}

	return url, nil
}
