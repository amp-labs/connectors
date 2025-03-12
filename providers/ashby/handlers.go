package ashby

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/ashby/metadata"
)

const (
	pageSizeKey = "limit"
	pageSize    = "100"
	pageKey     = "cursor"
	sinceKey    = "createdAfter"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	path, err := metadata.Schemas.LookupURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	body := buildRequestbody(params)

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonData))
}

func buildRequestbody(params common.ReadParams) map[string]any {
	body := make(map[string]any)

	if supportSince.Has(params.ObjectName) && !params.Since.IsZero() {
		body[sinceKey] = params.Since.UnixMilli()
	}

	if supportPagination.Has(params.ObjectName) && params.NextPage != "" {
		body[pageKey] = params.NextPage
		body[pageSizeKey] = pageSize
	}

	return body
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
		common.GetRecordsUnderJSONPath(responseKey),
		makeNextRecordsURL,
		common.GetMarshaledData,
		params.Fields,
	)
}
