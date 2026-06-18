package instantlyai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if !directResponseEndpoints.Has(params.ObjectName) {
		url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))
	}

	// https://developer.instantly.ai/api/v2/analytics/getdailycampaignanalytics
	if !params.Since.IsZero() && sinceSupportedEndpoints.Has(params.ObjectName) {
		url.WithQueryParam("start_date", params.Since.Format(time.DateOnly))
	}

	// https://developer.instantly.ai/api/v2/analytics/getdailycampaignanalytics
	if !params.Until.IsZero() && untilSupportedEndpoints.Has(params.ObjectName) {
		url.WithQueryParam("end_date", params.Until.Format(time.DateOnly))
	}

	if len(params.NextPage) != 0 {
		// Next page.
		url, err = urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}
	}

	if postEndpointsOfRead.Has(params.ObjectName) {
		return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader([]byte("{}")))
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	path := "items"

	if directResponseEndpoints.Has(params.ObjectName) {
		path = ""
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(path),
		makeNextRecordsURL(request.URL, params.ObjectName),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	if len(params.RecordId) > 0 {
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
		return &common.WriteResult{ // nolint:nilerr
			Success: true,
		}, nil
	}

	recordID, err := jsonquery.New(body).StrWithDefault("id", "")
	if err != nil {
		return nil, err
	}

	resp, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     resp,
	}, nil
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	// For the delete functionality, the lead-labels endpoint requires an empty object ({}) in the body parameters,
	// while other endpoints require a null value in the body parameters.
	// Refer sample delete objects body params https://developer.instantly.ai/api/v2/account/deleteaccount.
	body := []byte(`null`)

	// Refer link for lead-labels object https://developer.instantly.ai/api/v2/leadlabel/deleteleadlabel.
	if params.ObjectName == "lead-labels" {
		body = []byte("{}")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if resp.Code != http.StatusOK {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, resp.Code)
	}

	// A successful delete returns 200 OK
	return &common.DeleteResult{
		Success: true,
	}, nil
}
