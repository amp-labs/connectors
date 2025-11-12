package sellsy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/sellsy/internal/metadata"
	"github.com/spyzhov/ajson"
)

// Every request has a page limit in range [0,100].
// https://docs.sellsy.com/api/v2/#operation/get-contacts
const defaultPageSize = "100"

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	readURL, err := c.constructReadURL(ctx, params)
	if err != nil {
		return nil, err
	}

	method, jsonData, err := createReadOperation(readURL, params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, readURL.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) constructReadURL(ctx context.Context, params common.ReadParams) (*urlbuilder.URL, error) {
	if params.NextPage != "" {
		return urlbuilder.New(params.NextPage.String())
	}

	// This is the first, initial page for the object.
	// Page size query parameters:
	// https://docs.sellsy.com/api/v2/#operation/get-contacts
	readURL, err := c.getReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	readURL.WithQueryParam("limit", defaultPageSize)

	// Request custom field embedding.
	definitions, err := c.fetchCustomFieldDefinitions(ctx, []string{params.ObjectName})
	if err != nil {
		return nil, err
	}

	embedQueryParams := make([]string, 0)
	for _, fieldDefinitionID := range definitions[params.ObjectName].getIDs() {
		embedQueryParams = append(embedQueryParams, fmt.Sprintf("cf.%v", fieldDefinitionID))
	}

	if len(embedQueryParams) != 0 {
		readURL.WithQueryParamList("embed[]", embedQueryParams)
	}

	return readURL, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)

	return common.ParseResultFiltered(params, resp,
		common.MakeRecordsFunc(responseFieldName),
		makeFilterFunc(params, request),
		common.MakeMarshaledDataFunc(flattenCustomEmbed),
		params.Fields,
	)
}

func flattenCustomEmbed(node *ajson.Node) (map[string]any, error) {
	object, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	customFieldsResponse, err := jsonquery.ParseNode[customFieldReadResponse](node)
	if err != nil {
		return nil, err
	}

	// Attach custom fields on the top read object.
	for _, customField := range customFieldsResponse.Embed.CustomFields {
		object[customField.Code] = customField.Value
	}

	return object, nil
}

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	metadataResult, err := metadata.Schemas.Select(common.ModuleRoot, objectNames)
	if err != nil {
		return nil, err
	}

	definitions, err := c.fetchCustomFieldDefinitions(ctx, objectNames)
	if err != nil {
		return nil, err
	}

	for objectName, fields := range definitions {
		objectMetadata := metadataResult.Result[objectName]
		for _, field := range fields {
			objectMetadata.AddFieldMetadata(field.Code, common.FieldMetadata{
				DisplayName:  field.Name,
				ValueType:    field.getValueType(),
				ProviderType: field.Type,
				Values:       field.getValues(),
			})
		}
	}

	return metadataResult, nil
}

/*
Pagination uses cursor pagination which in Sellsy documentation is referred to as "Seek" Method.
https://docs.sellsy.com/api/v2/#section/Pagination-on-list-and-search-requests

When number of records is less than the max page size this signifies that we can ignore making the next page request.

Read Response format:

	{
	  ...
	  "pagination": {
		"limit": 2,
		"count": 2,
		"total": 32,
		"offset": "WyI0Il0="
	  }
	}
*/
func makeNextRecordsURL(requestURL *url.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		seekOffset, err := jsonquery.New(node, "pagination").StrWithDefault("offset", "")
		if err != nil {
			return "", err
		}

		if seekOffset == "" {
			// Next page doesn't exist.
			return "", nil
		}

		counter, _ := jsonquery.New(node, "pagination").IntegerWithDefault("count", 0)
		limit, _ := jsonquery.New(node, "pagination").IntegerWithDefault("limit", 0)

		if counter < limit {
			// This is the last page.
			// The next page cannot contain more records, so stop here.
			return "", nil
		}

		nextURL, err := urlbuilder.FromRawURL(requestURL)
		if err != nil {
			return "", err
		}

		nextURL.WithQueryParam("offset", seekOffset)

		return nextURL.String(), nil
	}
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	writeURL, err := c.getWriteURL(params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost
	if len(params.RecordId) != 0 {
		method = http.MethodPut
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, writeURL.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

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
		Errors:   nil,
		Data:     data,
	}, nil
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	deleteURL, err := c.getWriteURL(params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

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
