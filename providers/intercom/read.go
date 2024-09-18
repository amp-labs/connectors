package intercom

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
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

	rsp, err := c.Client.Get(ctx, url.String(), apiVersionHeader)
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
		return constructURL(config.NextPage.String())
	}

	// First page
	url, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("per_page", strconv.Itoa(DefaultPageSize))

	return url, nil
}
