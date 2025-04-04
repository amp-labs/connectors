package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/github/metadata"
	"github.com/spyzhov/ajson"
)

const (
	pageSizeKey = "per_page"
	pageSize    = "100"
	pageKey     = "page"
	pageNumber  = "1"
	sinceKey    = "since"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	path, err := metadata.Schemas.LookupURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := c.RootClient.URL(path)
	if err != nil {
		return nil, err
	}

	if supportPagination.Has(params.ObjectName) {
		url.WithQueryParam(pageSizeKey, pageSize)
		url.WithQueryParam(pageKey, pageNumber)
	}

	if supportSince.Has(params.ObjectName) && !params.Since.IsZero() {
		// https://docs.github.com/en/rest/gists/gists?apiVersion=2022-11-28#list-gists-for-the-authenticated-user
		url.WithQueryParam(sinceKey, datautils.Time.FormatRFC3339inUTC(params.Since))
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
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
		makeNextRecordsURL(response.Headers),
		common.GetMarshaledData,
		params.Fields,
	)
}

func makeNextRecordsURL(responseHeaders http.Header) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextURL := httpkit.HeaderLink(&common.JSONHTTPResponse{Headers: responseHeaders}, "next")
		if nextURL == "" {
			return "", nil
		}

		return nextURL, nil
	}
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) { //nolint:lll
	url, err := c.RootClient.URL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	if params.RecordId != "" {
		if !supportByUpdate.Has(params.ObjectName) {
			return nil, common.ErrOperationNotSupportedForObject
		}

		url.AddPath(params.RecordId)

		method = http.MethodPatch
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil { //nolint:nilerr
		return &common.WriteResult{
			Success: true,
			Errors:  nil,
			Data:    nil,
		}, nil
	}

	recordID, err := jsonquery.New(body).StrWithDefault("id", "")
	if err != nil { // nolint:nilerr
		return &common.WriteResult{
			Success: true,
			Errors:  nil,
			Data:    data,
		}, nil
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := c.RootClient.URL(params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
}

func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if response.Code != http.StatusOK && response.Code != http.StatusNoContent {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	// A successful delete returns 200 OK
	return &common.DeleteResult{
		Success: true,
	}, nil
}
