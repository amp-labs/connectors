package servicenow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers"
	"github.com/spyzhov/ajson"
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

	// Offset-paginated objects (e.g. the Knowledge and Change APIs) page via offset
	// rather than the Link header, so seed the first page's window here. Some accept
	// a page-size param (limitKey); others (Change) only take an offset.
	if pg, ok := offsetPaginationOf(params.ObjectName); ok {
		if pg.limitKey != "" {
			url.WithQueryParam(pg.limitKey, strconv.Itoa(offsetPageSize))
		}

		url.WithQueryParam(pg.offsetKey, "0")
	} else if pp, ok := pagePaginationOf(params.ObjectName); ok {
		url.WithQueryParam(pp.perPageKey, strconv.Itoa(offsetPageSize))
		url.WithQueryParam(pp.pageKey, "1")
	}

	// Incremental read: filter by sys_updated_on when Since/Until are set and the
	// object's list endpoint accepts sysparm_query.
	if query := incrementalQuery(params); query != "" {
		url.WithQueryParam("sysparm_query", query)
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
	// Objects that paginate via limit/offset (e.g. the Knowledge API) advance the
	// window from the request URL; everything else follows the Link header.
	nextPage := getNextRecordsURL(response, c.ProviderInfo().BaseURL)
	if _, ok := offsetPaginationOf(params.ObjectName); ok {
		nextPage = offsetNextPage(params.ObjectName, request)
	} else if _, ok := pagePaginationOf(params.ObjectName); ok {
		nextPage = pageNextPage(params.ObjectName, request)
	}

	return common.ParseResult(response,
		recordsFunc(params.ObjectName),
		nextPage,
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
		// No "result" envelope. Some APIs (lead and other TMF/Open APIs) return the
		// created record as a bare top-level object, so recover sys_id from the root.
		return c.writeResultFromRecord(ctx, params, body) //nolint: nilerr
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
	return c.writeResultFromRecord(ctx, params, result)
}

// writeResultFromRecord builds a WriteResult from a record object node, capturing
// its fields as Data and its sys_id as the RecordId. A node that isn't an object,
// or that carries no sys_id, still yields a successful result.
func (c *Connector) writeResultFromRecord(
	ctx context.Context,
	params common.WriteParams,
	record *ajson.Node,
) (*common.WriteResult, error) {
	if !record.IsObject() {
		return &common.WriteResult{Success: true}, nil
	}

	data, err := jsonquery.Convertor.ObjectToMap(record)
	if err != nil {
		logging.Logger(ctx).Error("failed to convert result object to map", "object", params.ObjectName, "err", err.Error())

		return &common.WriteResult{Success: true}, nil
	}

	recordID, err := jsonquery.New(record).StringOptional("sys_id")
	if err != nil || recordID == nil {
		return &common.WriteResult{ //nolint: nilerr
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
