package zoom

import (
	"context"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/zoom/metadata"
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

	rsp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	responseFieldName := metadata.Schemas.LookupArrayFieldName(common.ModuleRoot, config.ObjectName)

	return common.ParseResult(
		rsp,
		common.ExtractRecordsFromPath(responseFieldName),
		getNextRecordURL(url),
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

	url.WithQueryParam("page_size", strconv.Itoa(DefaultPageSize))

	addTimeFilterParams(config, url)

	return url, nil
}

// nolint:nestif
func addTimeFilterParams(config common.ReadParams, url *urlbuilder.URL) {
	// Zoom API uses "from" and "to" query parameters with date format YYYY-MM-DD.
	// Reference: https://developers.zoom.us/docs/api/meetings/#tag/archiving/get/archive_files
	if incrementalObjects.Has(config.ObjectName) {
		// Some objects require mandatory from/to parameters
		if mandatoryDateObjects.Has(config.ObjectName) {
			// Set default date range to last 29 days if not provided
			startDate := time.Now().AddDate(0, 0, -29)
			endDate := time.Now()

			if !config.Since.IsZero() {
				startDate = config.Since
			}

			if !config.Until.IsZero() {
				endDate = config.Until
			}

			url.WithQueryParam("from", startDate.UTC().Format(ZoomDateFormat))
			url.WithQueryParam("to", endDate.UTC().Format(ZoomDateFormat))
		} else {
			if !config.Since.IsZero() {
				url.WithQueryParam("from", config.Since.UTC().Format(ZoomDateFormat))
			}

			if !config.Until.IsZero() {
				url.WithQueryParam("to", config.Until.UTC().Format(ZoomDateFormat))
			}
		}
	}
}
