package klaviyo

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedObjectsByRead.Has(config.ObjectName) {
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

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		return urlbuilder.New(config.NextPage.String())
	}

	// First page
	url, err := c.getReadURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if !config.Since.IsZero() {
		if sinceField, found := objectsNameToSinceFieldName[config.ObjectName]; found {
			// Documentation about filtering: https://developers.klaviyo.com/en/docs/filtering_
			// Ex: ?filter=greater-than(datetime,2023-03-01T01:00:00Z)
			sinceValue := datautils.Time.FormatRFC3339inUTC(config.Since)
			url.WithQueryParam("filter", fmt.Sprintf("greater-than(%v,%v)", sinceField, sinceValue))
		}
	}

	return url, nil
}
