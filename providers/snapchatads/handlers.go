package snapchatads

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := c.constructURL(objectName)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	node, ok := response.Body() // nolint:varnamelen
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	objectResponse, err := jsonquery.New(node).ArrayRequired(objectName)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ArrayToMap(objectResponse)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	objKey := naming.NewSingularString(objectName).String()

	// Extract and assert the inner map
	innerData, ok := data[0][objKey].(map[string]any)
	if !ok {
		return nil, ErrObjNotFound
	}

	for field := range innerData {
		objectMetadata.FieldsMap[field] = field
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.constructURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", strconv.Itoa(defaultPageSize))

	if len(params.NextPage) != 0 {
		// Next page.
		url, err = urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
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
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(params.ObjectName),
		makeNextRecordsURL(),
		DataMarshall(response, params.ObjectName),
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := c.constructURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	// When updating an object, the record ID is not included in the URL path; instead,
	// it must be provided in the body parameters.
	// EX. Refer https://developers.snap.com/api/marketing-api/Ads-API/billing-centers#update-a-billing-center.
	if params.RecordId != "" {
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

	objectResponse, err := jsonquery.New(body).ArrayRequired(params.ObjectName)
	if err != nil {
		return nil, err
	}

	objKey := naming.NewSingularString(params.ObjectName).String()

	recordID, err := jsonquery.New(objectResponse[0], objKey).StringRequired("id")
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

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
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
	body, ok := resp.Body()
	if !ok {
		return &common.DeleteResult{ // nolint:nilerr
			Success: true,
		}, nil
	}

	objectResponse, err := jsonquery.New(body).ArrayRequired(params.ObjectName)
	if err != nil {
		return nil, err
	}

	res, err := jsonquery.Convertor.ArrayToMap(objectResponse)
	if err != nil {
		return nil, err
	}

	if len(res) != 0 {
		return nil, fmt.Errorf("%v", res[0]["sub_request_error_reason"])
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}
