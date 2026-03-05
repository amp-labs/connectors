package greenhouse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/greenhouse/metadata"
	"github.com/spyzhov/ajson"
)

const (
	// maxPageSize is the maximum number of records per page allowed by the Greenhouse Harvest API.
	// https://developers.greenhouse.io/harvest.html#pagination
	maxPageSize = 500
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Cursor-based pagination: use the full URL from the Link header directly.
		return urlbuilder.New(params.NextPage.String())
	}

	// First page: build URL from scratch.
	path, err := metadata.Schemas.FindURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "v3", path)
	if err != nil {
		return nil, err
	}

	// Respect user-provided page size, capped at the API maximum.
	pageSize := params.PageSize
	if pageSize <= 0 || pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	url.WithQueryParam("per_page", strconv.Itoa(pageSize))

	if !params.Since.IsZero() {
		url.WithQueryParam("updated_after", datautils.Time.FormatRFC3339inUTC(params.Since))
	}

	return url, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		resp,
		// Empty path means records are at the root level (response is a JSON array, not nested).
		common.ExtractOptionalRecordsFromPath(""),
		makeNextRecordsURL(resp),
		common.GetMarshaledData,
		params.Fields,
	)
}

// Next page is communicated via `Link` header under the `next` rel.
// https://developers.greenhouse.io/harvest.html#pagination
func makeNextRecordsURL(resp *common.JSONHTTPResponse) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return httpkit.HeaderLink(resp, "next"), nil
	}
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	path, err := metadata.Schemas.FindURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	var writeURL *urlbuilder.URL

	if len(params.RecordId) != 0 {
		method = http.MethodPatch
		writeURL, err = urlbuilder.New(c.ProviderInfo().BaseURL, "v3", path, params.RecordId)
	} else {
		writeURL, err = urlbuilder.New(c.ProviderInfo().BaseURL, "v3", path)
	}

	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, writeURL.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	common.Headers(
		common.TransformWriteHeaders(params.Headers, common.HeaderModeOverwrite),
	).ApplyToRequest(req)

	return req, nil
}

func (c *Connector) parseWriteResponse(ctx context.Context, params common.WriteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		// it is unlikely to have no payload
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	recordID, err := jsonquery.New(body).TextWithDefault("id", params.RecordId)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	path, err := metadata.Schemas.FindURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	deleteURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, "v3", path, params.RecordId)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL.String(), nil)
	if err != nil {
		return nil, err
	}

	common.Headers(
		common.TransformWriteHeaders(params.Headers, common.HeaderModeOverwrite),
	).ApplyToRequest(req)

	return req, nil
}

func (c *Connector) parseDeleteResponse(ctx context.Context, params common.DeleteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if response.Code != http.StatusOK && response.Code != http.StatusNoContent {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	// Response body is not used.
	return &common.DeleteResult{
		Success: true,
	}, nil
}
