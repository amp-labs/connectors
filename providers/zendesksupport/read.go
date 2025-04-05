package zendesksupport

import (
	"context"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
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

	rsp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module.ID, config.ObjectName)

	return common.ParseResult(
		rsp,
		common.ExtractRecordsFromPath(responseFieldName),
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

	pageSizeStr := metadata.Schemas.PageSize(c.Module.ID, config.ObjectName)
	isIncremental := metadata.Schemas.IsIncrementalRead(c.Module.ID, config.ObjectName)

	if isIncremental {
		// Incremental endpoints requires start query parameter.
		// Even if no Since parameter is empty the start_time must be set to 0.
		// This is effectively to say read everything since the beginning of time.
		// https://developer.zendesk.com/api-reference/ticketing/ticket-management/incremental_exports/#start_time
		url.WithQueryParam("start_time", formatStartTime(config))
		url.WithQueryParam("per_page", pageSizeStr)
	} else {
		// Different objects have different pagination types.
		// https://developer.zendesk.com/api-reference/introduction/pagination/#using-offset-pagination
		ptype := metadata.Schemas.LookupPaginationType(c.Module.ID, config.ObjectName)
		if ptype == "cursor" {
			url.WithQueryParam("page[size]", pageSizeStr)
		}
	}

	return url, nil
}

func formatStartTime(config common.ReadParams) string {
	if config.Since.IsZero() {
		return "0"
	}

	// Records cannot be requested if they are less than 1 minute old.
	unixTime := config.Since.Unix()
	timeWindow := time.Since(config.Since)

	if timeWindow.Minutes() < 1 {
		unixTime = time.Now().Add(-1 * time.Minute).Unix()
	}

	return strconv.FormatInt(unixTime, 10)
}
