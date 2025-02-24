package zendesksupport

import (
	"context"
	"strconv"

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
		common.GetRecordsUnderJSONPath(responseFieldName),
		getNextRecordsURL,
		common.GetMarshaledData,
		config.Fields,
	)
}

const DefaultPageSizeStr = "100"

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		return urlbuilder.New(config.NextPage.String())
	}

	// First page
	url, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	isIncremental := metadata.Schemas.IsIncrementalRead(c.Module.ID, config.ObjectName)
	if isIncremental {
		// Incremental endpoints requires start query parameter.
		// Even if no Since parameter is empty the start_time must be set to 0.
		// This is effectively to say read everything since the beginning of time.
		startTime := "0"
		if !config.Since.IsZero() {
			startTime = strconv.FormatInt(config.Since.Unix(), 10)
		}

		url.WithQueryParam("start_time", startTime)
		pageSizeQueryParam := metadata.Schemas.LookupPageSizeQP(c.Module.ID, config.ObjectName)
		url.WithQueryParam(pageSizeQueryParam, DefaultPageSizeStr)

	} else {
		// Different objects have different pagination types.
		// https://developer.zendesk.com/api-reference/introduction/pagination/#using-offset-pagination
		ptype := metadata.Schemas.LookupPaginationType(c.Module.ID, config.ObjectName)
		if ptype == "cursor" {
			url.WithQueryParam("page[size]", DefaultPageSizeStr)
		}
	}

	return url, nil
}
