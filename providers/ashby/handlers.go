package ashby

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/ashby/metadata"
)

const (
	pageSizeKey = "limit"
	pageSize    = "100"
	pageKey     = "cursor"
	sinceKey    = "createdAfter"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	path, err := metadata.Schemas.LookupURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, path)
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

	body[pageSizeKey] = pageSize

	if supportSince.Has(params.ObjectName) && !params.Since.IsZero() {
		body[sinceKey] = params.Since.UnixMilli()
	}

	if supportPagination.Has(params.ObjectName) && params.NextPage != "" {
		body[pageKey] = params.NextPage
	}

	return body
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		getRecords(params.ObjectName, c.Module()),
		makeNextRecordsURL,
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonData))
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	node, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	rawID, err := jsonquery.New(node, "results").StrWithDefault("id", "")
	if err != nil {
		return nil, err
	}

	if rawID == "" {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: rawID,
		Errors:   nil,
		Data:     nil,
	}, nil
}
