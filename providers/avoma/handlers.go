package avoma

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const objectNameTemplate = "template"

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		url, err := urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName, common.WithTrailingSlash)
	if err != nil {
		return nil, err
	}

	if endpointsWithResultsPath.Has(params.ObjectName) {
		url.WithQueryParam("page_size", pageSize)

		// Query parameters `from_date` and `to_date` are required for paginated objects in Avoma.
		if !params.Since.IsZero() && !params.Until.IsZero() {
			url.WithQueryParam("from_date", datautils.Time.FormatRFC3339inUTC(params.Since))
			url.WithQueryParam("to_date", datautils.Time.FormatRFC3339inUTC(params.Until))
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	nodePath := ""

	if endpointsWithResultsPath.Has(params.ObjectName) {
		nodePath = "results"
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(nodePath),
		makeNextRecordsURL(),
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

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		// The connector supports two endpoints for update the object one is template with PUT method
		// another one is smart_categories with PATCH method.
		switch params.ObjectName {
		case objectNameTemplate:
			method = http.MethodPut
		default:
			method = http.MethodPatch
		}
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(
		ctx,
		method,
		url.String()+common.WithTrailingSlash,
		bytes.NewReader(jsonData),
	)
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

	var searchValue string

	switch params.ObjectName {
	case "smart_categories", objectNameTemplate:
		searchValue = "uuid"
	case "calls":
		searchValue = "external_id"
	}

	recordID, err := jsonquery.New(body).StrWithDefault(searchValue, "")
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
