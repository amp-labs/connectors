package servicenow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	path, err := objectPath(objectName)
	if err != nil {
		return nil, err
	}

	// Metadata is sampled from a list response, so it is only available for objects
	// whose collection can be listed.
	if !slices.Contains(readSupportedObjects, objectName) {
		return nil, fmt.Errorf("%w: %s does not support metadata", common.ErrOperationNotSupportedForObject, objectName)
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, path)
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
	body, ok := response.Body()
	if !ok {
		return nil, common.ErrFailedToUnmarshalBody
	}

	// Extract records using the object's response shape (default/SCIM/nested/array),
	// so metadata works for every supported object, not only the {"result":[...]} ones.
	records, err := recordsFunc(objectName)(body)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	fields := make(common.FieldsMetadata)

	// Use the first record to sample fields. ServiceNow REST responses carry no
	// field type metadata, so we infer the value type from the sampled value.
	for field, value := range records[0] {
		fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    common.InferValueTypeFromData(value),
			ProviderType: "", // not provided by ServiceNow
		}
	}

	return common.NewObjectMetadata(
		objectName,
		fields,
	), nil
}

func (c *Connector) constructReadURL(params common.ReadParams) (string, error) {
	if params.NextPage != "" {
		return params.NextPage.String(), nil
	}

	path, err := objectPath(params.ObjectName)
	if err != nil {
		return "", err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, path)
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.constructReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(response,
		recordsFunc(params.ObjectName),
		getNextRecordsURL(response, c.ProviderInfo().BaseURL),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	logging.With(ctx, "connector", providers.ServiceNow)

	method := http.MethodPost

	path, err := objectPath(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, path)
	if err != nil {
		return nil, err
	}

	if len(params.RecordId) > 0 {
		url.AddPath(params.RecordId)

		method = http.MethodPatch
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, fmt.Errorf("marshalling request body: %w", err)
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

	result, err := body.GetKey("result")
	if err != nil {
		// No "result" field (e.g. an empty body). The write still succeeded.
		return &common.WriteResult{Success: true}, nil
	}

	// Some scoped APIs (e.g. Contact, Consumer) return only the new record's sys_id
	// as a bare string: {"result": "<sys_id>"}. We capture it as the RecordId.
	if result.IsString() {
		recordID := result.MustString()

		return &common.WriteResult{
			Success:  true,
			RecordId: recordID,
			Data:     map[string]any{"sys_id": recordID},
		}, nil
	}

	// Most APIs (Table API and the like) return the written record object:
	// {"result": {...}}.
	if !result.IsObject() {
		return &common.WriteResult{Success: true}, nil
	}

	data, err := jsonquery.Convertor.ObjectToMap(result)
	if err != nil {
		logging.Logger(ctx).Error("failed to convert result object to map", "object", params.ObjectName, "err", err.Error())

		return &common.WriteResult{Success: true}, nil
	}

	recordID, err := jsonquery.New(result).StringOptional("sys_id")
	if err != nil || recordID == nil {
		return &common.WriteResult{
			Success: true,
			Data:    data,
		}, nil
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: *recordID,
		Data:     data,
	}, nil
}
