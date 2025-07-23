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

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res,
		getRecords,
		getNextRecordsURL,
		common.MakeMarshaledDataFunc(common.FlattenNestedFields(attributesKey)),
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

	// Time based querying:
	// https://developers.outreach.io/api/making-requests/#filter-by-a-less-than-or-equal-to-condition
	// https://developers.outreach.io/api/making-requests/#new-filter-syntax
	if !config.Since.IsZero() {
		sinceTimestamp := config.Since.UTC().Format(time.DateOnly)

		untilTimestamp := "inf"
		if !config.Until.IsZero() {
			untilTimestamp = config.Until.UTC().Format(time.DateOnly)
		}

		url.WithQueryParam("filter[updatedAt]", sinceTimestamp+".."+untilTimestamp)
	}

	return url, nil
}
