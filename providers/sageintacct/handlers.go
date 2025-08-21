package sageintacct

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/sageintacct/metadata"
)

const (
	apiVersion      = "ia/api/v1"
	defaultPageSize = 3
	pageSizeParam   = "size"
	pageParam       = "start"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, body, err := buildURL(c.Module(), params, c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, err
	}

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
	responseKey := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(responseKey),
		makeNextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}
