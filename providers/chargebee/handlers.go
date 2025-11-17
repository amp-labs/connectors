package chargebee

import (
	"context"
	"fmt"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	apiVersion = "v2"
	pageSize   = 100
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	if objectNameWithListSuffix.Has(objectName) {
		objectName += "/list"
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName)
	if err != nil {
		return nil, err
	}

	// Add limit parameter to get just a single record for sampling
	url.WithQueryParam("limit", "1")

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
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

	res, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if res == nil || len(*res) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	records, ok := (*res)["list"].([]any) //nolint:varnamelen
	if !ok {
		return nil, fmt.Errorf("couldn't convert the list field to an array: %w", common.ErrMissingExpectedValues)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	firstRecord, ok := records[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record to a map: %w", common.ErrMissingExpectedValues)
	}

	objectResponseKey := objectResponseField.Get(objectName)

	// Example response structure
	// 	{
	//     "list": [
	//         {
	//             "customer": {...}
	//         },
	//     ]
	// }

	var objectData map[string]any
	if objectRecord, exists := firstRecord[objectResponseKey]; exists {
		objectData, ok = objectRecord.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("couldn't convert the %s object to a map: %w", objectName, common.ErrMissingExpectedValues)
		}
	} else {
		// If the object name key doesn't exist, use the record itself
		objectData = firstRecord
	}

	for field, value := range objectData {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    inferValueTypeFromData(value),
			ProviderType: "",
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	objectName := params.ObjectName
	if objectNameWithListSuffix.Has(objectName) {
		objectName += "/list"
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName)
	if err != nil {
		return nil, err
	}

	if params.NextPage != "" {
		url.WithQueryParam("offset", params.NextPage.String())
	}

	url.WithQueryParam("limit", strconv.FormatInt(int64(pageSize), 10))

	if supportIncrementalRead.Has(params.ObjectName) {
		if !params.Since.IsZero() {
			url.WithQueryParam("updated_at[after]", strconv.Itoa(int(params.Since.Unix())))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("updated_at[before]", strconv.Itoa(int(params.Until.Unix())))
		}
		// add sorting for faster response. recommended by Chargebee
		url.WithQueryParam("sort_by[asc]", "updated_at")
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
		extractRecords(params.ObjectName),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	fullObjectName := objectNameWrite.Get(params.ObjectName)

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, fullObjectName)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)
	}

	// Convert RecordData to form values
	formData := make(neturl.Values)

	recordMap, ok := params.RecordData.(map[string]any)
	if !ok || len(recordMap) == 0 {
		return nil, common.ErrMissingRecordData
	}

	for key, value := range recordMap {
		if str, ok := value.(string); ok {
			formData.Set(key, str)
		} else if value != nil {
			formData.Set(key, fmt.Sprintf("%v", value))
		}
	}

	body := strings.NewReader(formData.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	objectResponseKey := objectResponseField.Get(params.ObjectName)

	objectData, err := jsonquery.New(body).ObjectRequired(objectResponseKey)
	if err != nil || objectData == nil {
		// If the expected object key is not found, use the entire body as the object data
		objectData = body
	}

	recordID, err := jsonquery.New(objectData).StrWithDefault("id", "")
	if err != nil {
		return nil, err
	}

	resp, err := jsonquery.Convertor.ObjectToMap(objectData)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		RecordId: recordID,
		Errors:   nil,
		Success:  true,
		Data:     resp,
	}, nil
}
