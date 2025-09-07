package salesloft

import (
	"context"
	"strconv"

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

	rsp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		getRecords,
		makeNextRecordsURL(url),
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
	url, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("per_page", strconv.Itoa(DefaultPageSize))

	// Documentation states ISO8601, while server accepts different formats
	// but for consistency we are sticking to one format to be sent.
	// For the reference any API resource that includes time data type mentions iso8601 string format.
	// One example, say accounts is https://developers.salesloft.com/docs/api/accounts-index
	if !config.Since.IsZero() {
		updatedSince := datautils.Time.FormatRFC3339inUTCWithMicrosecondsAndOffset(config.Since)
		url.WithQueryParam("updated_at[gte]", updatedSince)
	}

	if !config.Until.IsZero() {
		updatedUntil := datautils.Time.FormatRFC3339inUTCWithMicrosecondsAndOffset(config.Until)
		url.WithQueryParam("updated_at[lte]", updatedUntil)
	}

	return url, nil
}
