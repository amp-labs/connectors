package greenhouse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/greenhouse/metadata"
	"github.com/spyzhov/ajson"
)

const (
	// maxPageSize is the maximum number of records per page allowed by the Greenhouse Harvest API.
	// https://developers.greenhouse.io/harvest.html#pagination
	maxPageSize = 500
)

// ListObjectMetadata returns metadata for the requested objects, including custom fields.
// Base metadata comes from the static OpenAPI schema. Custom field definitions are fetched
// from the Greenhouse API and added as top-level fields.
// https://harvestdocs.greenhouse.io/reference/get_v3-custom-fields
func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	metadataResult, err := metadata.Schemas.Select(c.Module(), objectNames)
	if err != nil {
		return nil, err
	}

	for _, objectName := range objectNames {
		fields, err := c.requestCustomFields(ctx, objectName)
		if err != nil {
			metadataResult.Errors[objectName] = err

			continue
		}

		objectMetadata := metadataResult.GetObjectMetadata(objectName)
		if objectMetadata == nil {
			continue
		}

		// Collect IDs for select-type fields to fetch options.
		var selectFieldIDs []int

		for _, field := range fields {
			if field.ValueType == "single_select" || field.ValueType == "multi_select" {
				selectFieldIDs = append(selectFieldIDs, field.ID)
			}
		}

		// Fetch options for select-type custom fields.
		options, err := c.requestCustomFieldOptions(ctx, selectFieldIDs)
		if err != nil {
			metadataResult.Errors[objectName] = err

			continue
		}

		// Add each custom field to the object metadata.
		for nameKey, field := range fields {
			fieldValues := getFieldValues(options, field.ID)

			objectMetadata.AddFieldMetadata(nameKey, common.FieldMetadata{
				DisplayName:  field.Name,
				ValueType:    field.getValueType(),
				ProviderType: field.ValueType,
				IsCustom:     goutils.Pointer(true),
				Values:       fieldValues,
			})
		}

		metadataResult.Result[objectName] = *objectMetadata
	}

	return metadataResult, nil
}

// getFieldValues returns the list of possible values for select-type custom fields.
func getFieldValues(options map[int][]customFieldOption, fieldID int) common.FieldValues {
	opts, ok := options[fieldID]
	if !ok || len(opts) == 0 {
		return nil
	}

	values := make(common.FieldValues, len(opts))
	for i, opt := range opts {
		values[i] = common.FieldValue{
			Value:        opt.Name,
			DisplayValue: opt.Name,
		}
	}

	return values
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Cursor-based pagination: use the full URL from the Link header directly.
		return urlbuilder.New(params.NextPage.String())
	}

	// First page: build URL from scratch.
	path, err := metadata.Schemas.FindURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "v3", path)
	if err != nil {
		return nil, err
	}

	// Respect user-provided page size, capped at the API maximum.
	pageSize := params.PageSize
	if pageSize <= 0 || pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	url.WithQueryParam("per_page", strconv.Itoa(pageSize))

	if !params.Since.IsZero() {
		url.WithQueryParam("updated_after", datautils.Time.FormatRFC3339inUTC(params.Since))
	}

	return url, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		resp,
		// Greenhouse v3 list endpoints return bare JSON arrays at root level.
		getRecords,
		makeNextRecordsURL(resp),
		common.MakeMarshaledDataFunc(flattenCustomFields),
		params.Fields,
	)
}

// getRecords extracts records from the root-level JSON array.
func getRecords(node *ajson.Node) ([]*ajson.Node, error) {
	if node == nil {
		return nil, nil //nolint:nilnil
	}

	if node.IsArray() {
		return node.GetArray()
	}

	return nil, nil //nolint:nilnil
}

// Next page is communicated via `Link` header under the `next` rel.
// https://developers.greenhouse.io/harvest.html#pagination
func makeNextRecordsURL(resp *common.JSONHTTPResponse) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return httpkit.HeaderLink(resp, "next"), nil
	}
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	path, err := metadata.Schemas.FindURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	var writeURL *urlbuilder.URL

	if len(params.RecordId) != 0 {
		method = http.MethodPatch
		writeURL, err = urlbuilder.New(c.ProviderInfo().BaseURL, "v3", path, params.RecordId)
	} else {
		writeURL, err = urlbuilder.New(c.ProviderInfo().BaseURL, "v3", path)
	}

	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, writeURL.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	common.Headers(
		common.TransformWriteHeaders(params.Headers, common.HeaderModeOverwrite),
	).ApplyToRequest(req)

	return req, nil
}

func (c *Connector) parseWriteResponse(ctx context.Context, params common.WriteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
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
		Data:     data,
	}, nil
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	path, err := metadata.Schemas.FindURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	deleteURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, "v3", path, params.RecordId)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL.String(), nil)
	if err != nil {
		return nil, err
	}

	common.Headers(
		common.TransformWriteHeaders(params.Headers, common.HeaderModeOverwrite),
	).ApplyToRequest(req)

	return req, nil
}

func (c *Connector) parseDeleteResponse(ctx context.Context, params common.DeleteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if response.Code != http.StatusOK && response.Code != http.StatusNoContent {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}
