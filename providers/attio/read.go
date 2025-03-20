package attio

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

	if supportAttioGeneralApi.Has(config.ObjectName) {
		return c.readGeneralAPI(ctx, config)
	}

	return c.readStandardOrCustomObject(ctx, config)
}

func (c *Connector) readGeneralAPI(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
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
		common.GetRecordsUnderJSONPath("data"),
		makeNextRecordsURL(url),
		common.GetMarshaledData,
		config.Fields,
	)
}

func (c *Connector) readStandardOrCustomObject(
	ctx context.Context, config common.ReadParams,
) (*common.ReadResult, error) {
	// To handle standarad/custom objects
	url, err := c.getObjectReadURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	body := constructBody(config)

	rsp, err := c.Client.Post(ctx, url.String(), body)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		common.GetRecordsUnderJSONPath("data"),
		makeNextRecordStandardObj(body),
		common.GetMarshaledData,
		config.Fields,
	)
}

func (c *Connector) buildURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page.
		return urlbuilder.New(config.NextPage.String())
	}

	url, err := c.getApiURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if supportLimitAndOffset.Has(config.ObjectName) {
		url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))
		url.WithQueryParam("offset", "0")
	}

	return url, nil
}

// construct body params for filter the data using since field.
func constructBody(config common.ReadParams) map[string]any {
	filter := make(map[string]any)

	body := map[string]any{}

	if len(config.NextPage) != 0 {
		offset, err := strconv.Atoi(config.NextPage.String())
		if err != nil {
			return nil
		}

		body["offset"] = offset
	}

	if !config.Since.IsZero() {
		filter["created_at"] = map[string]string{
			"$gte": config.Since.Format("2006-01-02T15:04:05.999999999Z"),
		}
	}

	body["limit"] = DefaultPageSize

	body["filter"] = filter

	return body
}
