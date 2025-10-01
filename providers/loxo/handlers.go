package loxo

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	if objectWithPrefixValue.Has(objectName) {
		objectName = "scorecards/" + objectName
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, objectName)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	nodePath := objectsNodePath.Get(objectName)

	res, err := jsonquery.New(body).ArrayOptional(nodePath)
	if err != nil {
		return nil, err
	}

	record, err := jsonquery.Convertor.ArrayToMap(res)
	if err != nil {
		return nil, err
	}

	if len(record) == 0 {
		return nil, common.ErrMissingMetadata
	}

	for field := range record[0] {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    common.ValueTypeOther,
			ProviderType: "",
			ReadOnly:     false,
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if len(params.NextPage) != 0 {
		// Next page.
		url, err := urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	if objectWithPrefixValue.Has(params.ObjectName) {
		params.ObjectName = "scorecards/" + params.ObjectName
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if paginationObjects.Has(params.ObjectName) {
		url.WithQueryParam("per_page", strconv.Itoa(defaultPageSize))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	url, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(objectsNodePath.Get(params.ObjectName)),
		makeNextRecordsURL(url, params.ObjectName),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if objectWithPrefixValue.Has(params.ObjectName) {
		params.ObjectName = "scorecards/" + params.ObjectName
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	if len(params.RecordId) > 0 {
		url.AddPath(params.RecordId)

		method = http.MethodPut
	}

	var requestBody bytes.Buffer

	// For the Post endpoints, form-data is requires in the body param.
	writer := multipart.NewWriter(&requestBody)

	fields, ok := params.RecordData.(map[string]any)
	if !ok {
		return nil, common.ErrFieldTypeUnknown
	}

	// Add all fields from the map to the form
	for key, value := range fields {
		if err := writer.WriteField(key, fmt.Sprintf("%v", value)); err != nil {
			return nil, err
		}
	}

	// Close writer to finalize the form-data
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), &requestBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	req.Header.Set("Accept", "application/json")

	return req, nil
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

	nodePath := naming.NewSingularString(params.ObjectName).String()

	if writeObjectWithNoNodePath.Has(params.ObjectName) {
		nodePath = ""
	}

	responseObj, err := jsonquery.New(body).ObjectRequired(nodePath)
	if err != nil {
		return nil, err
	}

	recordID, err := jsonquery.New(responseObj).IntegerWithDefault("id", 0)
	if err != nil {
		return nil, err
	}

	resp, err := jsonquery.Convertor.ObjectToMap(responseObj)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: strconv.Itoa(int(recordID)),
		Errors:   nil,
		Data:     resp,
	}, nil
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	if objectWithPrefixValue.Has(params.ObjectName) {
		params.ObjectName = "scorecards/" + params.ObjectName
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

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
