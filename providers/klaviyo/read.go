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

	if !supportedObjectsByRead[c.Module.ID].Has(config.ObjectName) {
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
		common.GetMarshaledData,
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

	filter := filterBuilder{
		custom: config.Filter,
	}

	if !config.Since.IsZero() {
		if sinceField, found := objectsNameToSinceFieldName[c.Module.ID][config.ObjectName]; found {
			// Documentation about filtering: https://developers.klaviyo.com/en/docs/filtering_
			// Ex: ?filter=greater-than(datetime,2023-03-01T01:00:00Z)
			sinceValue := datautils.Time.FormatRFC3339inUTC(config.Since)
			filter.since = fmt.Sprintf("greater-than(%v,%v)", sinceField, sinceValue)
		}
	}

	if queryParam := filter.queryParameter(); len(queryParam) != 0 {
		url.WithQueryParam("filter", queryParam)
	}

	return url, nil
}

type filterBuilder struct {
	since  string
	custom string
}

func (b filterBuilder) queryParameter() string {
	if len(b.since) == 0 {
		return b.custom
	}

	if len(b.custom) == 0 {
		return b.since
	}

	// Both values are set. As per documentation these values can be comma separated.
	// Reference: https://developers.klaviyo.com/en/docs/filtering_
	return b.since + "," + b.custom
}
