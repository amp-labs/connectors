package constantcontact

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/constantcontact/metadata"
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
		makeNextRecordsURL(c.BaseURL),
		common.GetMarshaledData,
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		// Cursor query parameter is base64 encoded data which preserves all query parameters from initial request.
		// Therefore, this URL is ready for usage as is.
		// Example:
		// https://api.cc.email/v3/contacts?
		//			cursor=bGltaXQ9MSZuZXh0PTImdXBkYXRlZF9hZnRlcj0yMDIyLTAzLTExVDIyJTNBMDklM0EwMiUyQjAwJTNBMDA=
		return urlbuilder.New(config.NextPage.String())
	}

	// First page
	url, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))

	if !config.Since.IsZero() {
		switch config.ObjectName {
		case objectNameEmailCampaigns:
			sinceValue := datautils.Time.FormatRFC3339inUTC(config.Since)
			url.WithQueryParam("after_date", sinceValue)
		case objectNameContacts:
			sinceValue := datautils.Time.FormatRFC3339inUTC(config.Since)
			url.WithQueryParam("updated_after", sinceValue)
		}
	}

	return url, nil
}
