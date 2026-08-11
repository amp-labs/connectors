package zoominfo

import (
	"context"
	"fmt"
	"maps"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// buildSingleObjectMetadataRequest constructs the metadata request for one object.
// The request shape depends on the object kind:
//   - search: GET {dataAPIPath}/lookup/search?filter[entity]={entity}&filter[fieldType]=output
//   - enrich: GET {dataAPIPath}/lookup/enrich?filter[entity]={entity}&filter[fieldType]=output
//     (both describe the declared output schema — no live sampling)
//   - lookup: GET {dataAPIPath}/lookup/{fieldName}   (sample one record)
//   - get:    GET {segments...}                       (sample one record)
func (c *Connector) buildSingleObjectMetadataRequest(
	ctx context.Context,
	objectName string,
) (*http.Request, error) {
	switch kindOf(objectName) {
	case kindSearch:
		return c.buildLookupFieldsRequest(ctx, "search", searchObjects[objectName].entity)
	case kindEnrich:
		return c.buildLookupFieldsRequest(ctx, segEnrich, enrichObjects[objectName].entity)
	case kindLookup:
		return c.buildLookupMetadataRequest(ctx, objectName)
	case kindGet:
		return c.buildGetMetadataRequest(ctx, objectName)
	case kindUnknown:
		fallthrough
	default:
		return nil, fmt.Errorf("%w: %q", common.ErrObjectNotSupported, objectName)
	}
}

// buildLookupFieldsRequest discovers an entity's output fields via the
// lookup/{operation} endpoint (operation is "search" or "enrich"). These
// endpoints describe an entity's fields directly — deterministic and without
// needing criteria, match input, or live data.
func (c *Connector) buildLookupFieldsRequest(
	ctx context.Context,
	operation, entity string,
) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, dataAPIPath, segLookup, operation)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(filterEntityParam, entity)
	url.WithQueryParam(filterFieldTypeParam, outputFieldType)

	return c.newJSONAPIGetRequest(ctx, url)
}

func (c *Connector) buildLookupMetadataRequest(
	ctx context.Context,
	objectName string,
) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, dataAPIPath, segLookup, objectName)
	if err != nil {
		return nil, err
	}

	return c.newJSONAPIGetRequest(ctx, url)
}

func (c *Connector) buildGetMetadataRequest(
	ctx context.Context,
	objectName string,
) (*http.Request, error) {
	def := getObjects[objectName]

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, def.segments...)
	if err != nil {
		return nil, err
	}

	return c.newJSONAPIGetRequest(ctx, url)
}

func (c *Connector) newJSONAPIGetRequest(ctx context.Context, url *urlbuilder.URL) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", jsonAPIMediaType)

	return req, nil
}

// fieldListResponse is the lookup/{search,enrich} payload: a list of field
// descriptors under data[].attributes.fieldName.
type fieldListResponse struct {
	Data []struct {
		Attributes struct {
			FieldName string `json:"fieldName"`
		} `json:"attributes"`
	} `json:"data"`
}

// parseSingleObjectMetadataResponse builds an object's metadata. Search and enrich
// objects come from the lookup/{search,enrich} endpoints (a declared list of
// output field names); every other kind is sampled from the first returned record.
func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	switch kindOf(objectName) {
	case kindSearch, kindEnrich:
		return parseFieldListResponse(objectName, response)
	case kindLookup, kindGet, kindUnknown:
		return parseSampledRecordResponse(objectName, response)
	default:
		return parseSampledRecordResponse(objectName, response)
	}
}

// parseFieldListResponse builds metadata from a lookup/{search,enrich} output-field
// list. These endpoints return field names (and descriptions) but no type
// information, so ValueType is left as "other".
func parseFieldListResponse(
	objectName string,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	metadata := &common.ObjectMetadata{
		DisplayName: displayNameFor(objectName),
		Fields:      make(common.FieldsMetadata),
	}

	resp, err := common.UnmarshalJSON[fieldListResponse](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(resp.Data) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	for _, field := range resp.Data {
		name := field.Attributes.FieldName
		if name == "" {
			continue
		}

		metadata.Fields[name] = common.FieldMetadata{
			DisplayName: name,
			ValueType:   common.ValueTypeOther,
		}
	}

	return metadata, nil
}

// parseSampledRecordResponse infers an object's fields from the first record of a
// JSON:API data[] response. Record fields nested under "attributes" are promoted
// to the top level (mirroring how Read flattens them) alongside top-level keys
// like "id".
func parseSampledRecordResponse(
	objectName string,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	metadata := &common.ObjectMetadata{
		DisplayName: displayNameFor(objectName),
		Fields:      make(common.FieldsMetadata),
	}

	record, err := firstRecord(response)
	if err != nil {
		return nil, err
	}

	for field, value := range flattenRecord(record) {
		metadata.Fields[field] = common.FieldMetadata{
			DisplayName: field,
			ValueType:   common.InferValueTypeFromData(value),
		}
	}

	return metadata, nil
}

// firstRecord returns the first JSON:API resource object from a data[] response,
// or ErrMissingExpectedValues when there are no records to sample.
func firstRecord(response *common.JSONHTTPResponse) (map[string]any, error) {
	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	records, err := jsonquery.New(body).ArrayOptional("data")
	if err != nil {
		return nil, err
	}

	rows, err := jsonquery.Convertor.ArrayToMap(records)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	return rows[0], nil
}

// flattenRecord promotes the contents of a JSON:API record's "attributes" object
// to the top level, preserving other top-level keys (id, type). Attributes take
// precedence on key collisions, keeping the result deterministic regardless of
// map iteration order.
func flattenRecord(record map[string]any) map[string]any {
	out := make(map[string]any, len(record))

	for key, value := range record {
		if key == attributesField {
			continue
		}

		out[key] = value
	}

	if attrs, ok := record[attributesField].(map[string]any); ok {
		maps.Copy(out, attrs)
	}

	return out
}
