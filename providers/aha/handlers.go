package aha

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/aha/metadata"
	"github.com/spyzhov/ajson"
)

const (
	pageSizeKey = "per_page"
	pageSize    = "200"
	pageKey     = "page"
	sinceKey    = "created_since"
	untilKey    = "created_before"
	apiVersion  = "v1"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	path, err := metadata.Schemas.LookupURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
	if err != nil {
		return nil, err
	}

	addQueryParams(url, params, 1)

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseKey := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)

	path, err := metadata.Schemas.LookupURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	baseURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(responseKey),
		makeNextRecordsURL(baseURL, params),
		common.GetMarshaledData,
		params.Fields,
	)
}

func makeNextRecordsURL(baseURL *urlbuilder.URL, params common.ReadParams) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		pagination, err := jsonquery.New(node).ObjectRequired("pagination")
		if err != nil {
			return "", nil //nolint:nilerr
		}

		totalPage, err := jsonquery.New(pagination).IntegerRequired("total_pages")
		if err != nil {
			return "", nil //nolint:nilerr
		}

		currentPage, err := jsonquery.New(pagination).IntegerRequired("current_page")
		if err != nil {
			return "", nil //nolint:nilerr
		}

		if currentPage == totalPage {
			return "", nil //nolint:nilerr
		}

		nextPage := currentPage + 1

		addQueryParams(baseURL, params, nextPage)

		return baseURL.String(), nil
	}
}

func addQueryParams(url *urlbuilder.URL, params common.ReadParams, page int64) {
	url.WithQueryParam(pageSizeKey, pageSize)
	url.WithQueryParam(pageKey, strconv.FormatInt(page, 10))

	if supportSince.Has(params.ObjectName) && !params.Since.IsZero() {
		url.WithQueryParam(sinceKey, datautils.Time.FormatRFC3339inUTC(params.Since))
	}

	if supportUntil.Has(params.ObjectName) && !params.Until.IsZero() {
		url.WithQueryParam(untilKey, datautils.Time.FormatRFC3339inUTC(params.Until))
	}
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	var (
		url    *urlbuilder.URL
		err    error
		method = http.MethodPost
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPut
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
		return &common.WriteResult{ // nolint:nilerr
			Success: true,
		}, nil
	}

	resObj, err := jsonquery.New(body).ObjectRequired(naming.NewSingularString(params.ObjectName).String())
	if err != nil {
		return &common.WriteResult{ // nolint:nilerr
			Success: true,
		}, nil
	}

	data, err := jsonquery.Convertor.ObjectToMap(resObj)
	if err != nil {
		return &common.WriteResult{ // nolint:nilerr
			Success: true,
		}, nil
	}

	recordID, err := jsonquery.New(resObj).StringRequired("id")
	if err != nil {
		return &common.WriteResult{ // nolint:nilerr
			Success: true,
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
