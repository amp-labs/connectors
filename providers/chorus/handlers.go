package chorus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const apiVersion = "v1"

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	// All objects use the v1 version, except the engagements object (which uses v3 for Get Conversations).
	switch objectName {
	case objectEngagement:
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, "v3", objectName)
	default:
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, "api", apiVersion, objectName)
	}

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.api+json")

	return req, nil
}

// nolint:funlen
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

	// All objects have the nodePath value as "data", except the engagements object, which uses "engagements".
	// https://api-docs.chorus.ai/#03ff1d49-b8fb-4c8a-9407-d32e5f975964
	nodePath := "data"

	if objectName == objectEngagement {
		nodePath = "engagements"
	}

	res, err := jsonquery.New(body).ArrayRequired(nodePath)
	if err != nil {
		return nil, err
	}

	record, err := jsonquery.Convertor.ArrayToMap(res)
	if err != nil {
		return nil, err
	}

	if len(record) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	// helper to create FieldMetadata
	newField := func(name string) common.FieldMetadata {
		return common.FieldMetadata{
			DisplayName:  name,
			ValueType:    common.ValueTypeOther,
			ProviderType: "", // not available
			Values:       nil,
		}
	}

	// Attributes represent the object fields in the response. All actual data is embedded under the "attributes" field.
	// Sample response:
	// {
	//   "data": [
	//     {
	//       "attributes": {
	//         "filter_name": "string",
	//         "filter_type": "string",
	//         "field_type": "string",
	//         "filter_values": null
	//       },
	//       "type": "engagement_filter",
	//       "id": "123"
	//     }
	//   ]
	// }
	// Refer to the API response documentation at:
	// https://api-docs.chorus.ai/#f8b34d44-df36-47eb-a42e-a112aa0ec474.
	for field, value := range record[0] {
		if field == "attributes" {
			if subfields, ok := value.(map[string]any); ok {
				for subfield := range subfields {
					objectMetadata.Fields[subfield] = newField(subfield)
				}
			} else {
				return nil, common.ErrMissingFields
			}

			continue
		}

		objectMetadata.Fields[field] = newField(field)
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	switch params.ObjectName {
	case objectEngagement:
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, "v3", params.ObjectName)
	default:
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, "api", apiVersion, params.ObjectName)
	}

	if err != nil {
		return nil, err
	}

	if PaginationObject.Has(params.ObjectName) {
		url.WithQueryParam("page[size]", strconv.Itoa(PageSize))

		if params.NextPage != "" {
			url.WithQueryParam("page[number]", params.NextPage.String())
		}
	}

	// continuation_key is the pagination parameter used to retrieve the next page of results in the engagements object.
	if params.ObjectName == objectEngagement && params.NextPage != "" {
		url.WithQueryParam("continuation_key", params.NextPage.String())
	}

	if IncrementalObjectQueryParam.Has(params.ObjectName) {
		startDate := params.Since.Format(time.RFC3339)

		endDate := params.Until.Format(time.RFC3339)

		url.WithQueryParam(IncrementalObjectQueryParam.Get(params.ObjectName), startDate+":"+endDate)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.api+json")

	return req, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	var (
		nextPage int
		err      error
	)

	if params.NextPage != "" {
		nextPage, err = strconv.Atoi(params.NextPage.String())
		if err != nil {
			return nil, err
		}
	}

	nodePath := "data"

	if params.ObjectName == objectEngagement {
		nodePath = "engagements"
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(nodePath),
		makeNextRecord(nextPage, nodePath),
		DataMarshall(response, nodePath),
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "api", apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPut
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.api+json")

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

	recordID, err := jsonquery.New(body, "data").StrWithDefault("id", "")
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
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "api", apiVersion, params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
}

func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if resp.Code != http.StatusNoContent {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, resp.Code)
	}

	// A successful delete returns 200 OK
	return &common.DeleteResult{
		Success: true,
	}, nil
}
