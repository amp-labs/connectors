package pipedrive

import (
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// Read retrieves data based on the provided read parameters.
// https://developers.pipedrive.com/docs/api/v1
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(resp,
		common.ExtractOptionalRecordsFromPath("data"),
		nextRecordsURL(url),
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

	url, err := c.getReadURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	// begin fetching objects at provided start date
	// Supporting objects are: [Notes].
	// https://developers.pipedrive.com/docs/api/v1/Notes#getNotes
	if !config.Since.IsZero() {
		since := config.Since.UTC().Format(time.DateTime)
		url.WithQueryParam("start_date", since)
	}

	if !config.Until.IsZero() {
		until := config.Until.UTC().Format(time.DateTime)
		url.WithQueryParam("end_date", until)
	}

	return url, nil
}
