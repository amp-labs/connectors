package attio

import (
	"context"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if supportAttioApi.Has(config.ObjectName) {
		return c.readAPI(ctx, config)
	}

	return c.readStandardOrCustomObject(ctx, config)
}

func (c *Connector) readAPI(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
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
		common.ExtractRecordsFromPath("data"),
		makeNextRecordsURL(url, config.ObjectName),
		common.GetMarshaledData,
		config.Fields,
	)
}

func (c *Connector) readStandardOrCustomObject(
	ctx context.Context, config common.ReadParams,
) (*common.ReadResult, error) {
	// To handle standard/custom objects
	url, err := c.getObjectReadURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	body := constructRequestBody(config)

	offset := 0

	if val, ok := body["offset"].(int); ok {
		offset = val
	}

	rsp, err := c.Client.Post(ctx, url.String(), body)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		common.ExtractRecordsFromPath("data"),
		makeNextRecordStandardObj(offset),
		DataMarshall(rsp),
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
		var pageSize int

		if config.ObjectName == objectNameNotes {
			pageSize = DefaultPageSizeForNotesObj
		} else {
			pageSize = DefaultPageSize
		}

		url.WithQueryParam("limit", strconv.Itoa(pageSize))
		url.WithQueryParam("offset", "0")
	}

	return url, nil
}

// construct body params for filter the data using since field.
func constructRequestBody(config common.ReadParams) map[string]any {
	body := map[string]any{}

	if len(config.NextPage) != 0 {
		offset, err := strconv.Atoi(config.NextPage.String())
		if err != nil {
			return nil
		}

		body["offset"] = offset
	}

	if !config.Since.IsZero() {
		filter := make(map[string]any)

		filter["created_at"] = map[string]string{
			"$gte": datautils.Time.FormatRFC3339inUTCWithMilliseconds(config.Since),
		}

		body["filter"] = filter
	}

	body["limit"] = DefaultPageSize

	return body
}

func (c *Connector) geStandardOrCustomObjectsList(
	ctx context.Context,
) ([]objectData, error) {
	url, err := c.getApiURL("objects")
	if err != nil {
		return nil, fmt.Errorf("failed to build objects URL: %w", err)
	}

	rsp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get objects: %w", err)
	}

	result, err := common.UnmarshalJSON[objectListResponse](rsp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal objects response: %w", err)
	}

	return result.Data, nil
}
