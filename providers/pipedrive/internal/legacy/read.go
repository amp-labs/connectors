package legacy

import (
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if c.crmAdapter != nil {
		return c.crmAdapter.Read(ctx, params)
	}

	return c.readLegacy(ctx, params)
}

// Read retrieves data based on the provided read parameters.
// https://developers.pipedrive.com/docs/api/v1
func (a *Adapter) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := a.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	resp, err := a.Client.Get(ctx, url.String())
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

func (a *Adapter) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	// If NextPage is set, then we're reading the next page of results.
	// The NextPage URL has all the necessary parameters.
	if len(config.NextPage) > 0 {
		return urlbuilder.New(config.NextPage.String())
	}

	url, err := a.getReadURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	// begin fetching objects at provided start date
	// Supporting objects are: Activities & Notes only.
	if !config.Since.IsZero() {
		since := config.Since.UTC().Format(time.DateTime)
		url.WithQueryParam("start_date", since)
	}

	return url, nil
}
