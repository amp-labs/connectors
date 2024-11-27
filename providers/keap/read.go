package keap

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/keap/metadata"
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

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module.ID, config.ObjectName)

	return common.ParseResult(res,
		common.GetOptionalRecordsUnderJSONPath(responseFieldName),
		makeNextRecordsURL(c.Module.ID),
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

	if c.Module.ID == ModuleV1 {
		url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))

		if !config.Since.IsZero() {
			url.WithQueryParam("since", datautils.Time.FormatRFC3339inUTC(config.Since))
		}
	} else if c.Module.ID == ModuleV2 {
		if config.ObjectName == "contact_link_types" {
			url.WithQueryParam("pageSize", strconv.Itoa(DefaultPageSize))
		} else {
			url.WithQueryParam("page_size", strconv.Itoa(DefaultPageSize))
		}
	}

	return url, nil
}
