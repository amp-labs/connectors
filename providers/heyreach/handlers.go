package heyreach

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	urlPath, err := matchReadObjectNameToEndpointPath(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "public", urlPath)
	if err != nil {
		return nil, err
	}

	body := constructRequestBody(params)

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonData))
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	offset, _ := strconv.Atoi(params.NextPage.String())

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("items"),
		makeNextRecord(offset),
		common.GetMarshaledData,
		params.Fields,
	)
}

// construct body params for giving pagination value.
func constructRequestBody(config common.ReadParams) map[string]any {
	body := map[string]any{}

	if len(config.NextPage) != 0 {
		offset, err := strconv.Atoi(config.NextPage.String())
		if err != nil {
			return nil
		}

		body["offset"] = offset
	}

	body["limit"] = DefaultPageSize

	return body
}
