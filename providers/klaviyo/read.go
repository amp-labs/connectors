package klaviyo

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/klaviyo/internal/filtering"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedObjectsByRead[common.ModuleRoot].Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, url.String(), c.revisionHeader())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res,
		getRecords,
		getNextRecordsURL,
		common.MakeMarshaledDataFunc(common.FlattenNestedFields("attributes")),
		config.Fields,
	)
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Next page
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := c.getReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	// Build query.
	query := filtering.NewQuery().
		WithCustomFiltering(params.Filter)

	// Attach time-based filtering if supported
	if timeField, found := objectsNameToSinceFieldName[common.ModuleRoot][params.ObjectName]; found {
		query.
			WithSince(params.Since, timeField).
			WithUntil(params.Until, timeField)
	}

	// Apply query to URL.
	if queryParam := query.String(); len(queryParam) != 0 {
		url.WithQueryParam("filter", queryParam)
	}

	return url, nil
}
