package freshchat

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	apiVersion = "v2"
	pageSize   = "100"
)

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.PageSize != 0 {
		url.WithQueryParam("items_per_page", strconv.Itoa(params.PageSize))
	} else {
		url.WithQueryParam("items_per_page", pageSize)
	}

	// users requires 1 filtering query parameter.
	isDefaultTimeRange := params.Since.IsZero() && params.Until.IsZero()
	if params.ObjectName == "users" && isDefaultTimeRange {
		url.WithQueryParam("created_from", time.Unix(0, 0).UTC().GoString())
	}

	if !params.Since.IsZero() {
		if supportFilteringByTime.Has(params.ObjectName) {
			url.WithQueryParam("updated_from", params.Since.Format(time.RFC3339))
		}
	}

	if !params.Until.IsZero() {
		if supportFilteringByTime.Has(params.ObjectName) {
			url.WithQueryParam("updated_to", params.Until.Format(time.RFC3339))
		}
	}

	return url, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	urlbuild, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, urlbuild.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		records(params.ObjectName),
		nextRecordsURL(c.ProviderInfo().BaseURL),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	method := http.MethodPost

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
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
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	resp, err := jsonquery.New(body).ObjectRequired("")
	if err != nil {
		return nil, err
	}

	recordId, err := jsonquery.New(resp).StringOptional("id")
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(resp)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: *recordId,
		Data:     data,
	}, nil
}
