package asana

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {

	err := config.ValidateParams((true))

	if err != nil {
		return nil, err
	}

	if !supportedObjectsByRead.Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := c.buildURL(config)
	if err != nil {
		return nil, err
	}

	rsp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		common.GetOptionalRecordsUnderJSONPath("data"),
		makeNextRecordsURL(),
		common.GetMarshaledData,
		config.Fields,
	)

}

func (c *Connector) buildURL(config common.ReadParams) (*urlbuilder.URL, error) {

	if len(config.NextPage) != 0 {
		return urlbuilder.New(config.NextPage.String())
	}

	url, err := c.geAPIURL(config.ObjectName)

	if err != nil {
		return nil, err
	}

	if supportLimitAndOffset.Has(config.ObjectName) {
		url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))
	}

	return url, err

}
