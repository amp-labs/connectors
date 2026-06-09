package zoominfo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// searchRequestBody is the JSON:API envelope for a POST search/enrich call.
type searchRequestBody struct {
	Data searchRequestData `json:"data"`
}

type searchRequestData struct {
	Type       string         `json:"type"`
	Attributes map[string]any `json:"attributes"`
}

// buildSingleObjectMetadataRequest constructs a sampling request for one object.
// The request shape depends on the object kind:
//   - search: POST {dataAPIPath}/{resource}/search with a JSON:API body and page[size]=1
//   - lookup: GET  {dataAPIPath}/lookup/{fieldName}
//   - enrich: POST {dataAPIPath}/{segments...}/enrich with a JSON:API body
//   - get:    GET  {segments...} (segments carry their own version prefix)
func (c *Connector) buildSingleObjectMetadataRequest(
	ctx context.Context,
	objectName string,
) (*http.Request, error) {
	switch kindOf(objectName) {
	case kindSearch:
		return c.buildSearchMetadataRequest(ctx, objectName)
	case kindLookup:
		return c.buildLookupMetadataRequest(ctx, objectName)
	case kindEnrich:
		return c.buildEnrichMetadataRequest(ctx, objectName)
	case kindGet:
		return c.buildGetMetadataRequest(ctx, objectName)
	case kindUnknown:
		fallthrough
	default:
		return nil, fmt.Errorf("%w: %q", common.ErrObjectNotSupported, objectName)
	}
}

func (c *Connector) buildSearchMetadataRequest(
	ctx context.Context,
	objectName string,
) (*http.Request, error) {
	def := searchObjects[objectName]

	// The object name is the resource path segment (e.g. "contacts").
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, dataAPIPath, objectName, "search")
	if err != nil {
		return nil, err
	}

	// One record is enough to sample the field set.
	url.WithQueryParam("page[size]", metadataPageSize)

	return c.newJSONAPIPostRequest(ctx, url, def.searchType, def.sampleCriteria)
}

func (c *Connector) buildEnrichMetadataRequest(
	ctx context.Context,
	objectName string,
) (*http.Request, error) {
	def := enrichObjects[objectName]

	segments := append([]string{dataAPIPath}, def.segments...)

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, segments...)
	if err != nil {
		return nil, err
	}

	return c.newJSONAPIPostRequest(ctx, url, def.enrichType, nil)
}

func (c *Connector) buildLookupMetadataRequest(
	ctx context.Context,
	objectName string,
) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, dataAPIPath, "lookup", objectName)
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

// newJSONAPIPostRequest builds a POST request carrying a JSON:API envelope with
// the given attributes (a nil map is encoded as {}). Objects that still require
// input beyond what we seed (e.g. most enrich endpoints, intent search) return a
// descriptive 4xx, which the schema provider records per-object in
// ListObjectMetadataResult.Errors.
func (c *Connector) newJSONAPIPostRequest(
	ctx context.Context,
	url *urlbuilder.URL,
	resourceType string,
	attributes map[string]any,
) (*http.Request, error) {
	if attributes == nil {
		attributes = map[string]any{}
	}

	payload, err := json.Marshal(searchRequestBody{
		Data: searchRequestData{
			Type:       resourceType,
			Attributes: attributes,
		},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", jsonAPIMediaType)
	req.Header.Set("Content-Type", jsonAPIMediaType)

	return req, nil
}

func (c *Connector) newJSONAPIGetRequest(ctx context.Context, url *urlbuilder.URL) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", jsonAPIMediaType)

	return req, nil
}

// parseSingleObjectMetadataResponse infers an object's fields from the first
// record of a JSON:API response. ZoomInfo returns either a "data" array (most
// endpoints) or a singleton "data" object (e.g. customer-settings); both are
// handled. Record fields nested under "attributes" are promoted to the top level
// (mirroring how Read will flatten them) alongside top-level keys like "id".
func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	metadata := &common.ObjectMetadata{
		DisplayName: displayNameFor(objectName),
		Fields:      make(common.FieldsMetadata),
	}

	record, err := firstRecord(response, objectName)
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

// firstRecord returns the first JSON:API resource object from a response,
// transparently handling both the data[] (list) and data{} (singleton) shapes.
// It relies on jsonquery's typed errors rather than inspecting the raw node:
// ArrayOptional reports ErrNotArray for a singleton object, which is the signal
// to fall back to ObjectOptional.
func firstRecord(response *common.JSONHTTPResponse, objectName string) (map[string]any, error) {
	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	records, err := jsonquery.New(body).ArrayOptional("data")

	switch {
	case err == nil:
		if len(records) == 0 {
			return nil, fmt.Errorf("%w: no records returned to sample fields for %q",
				common.ErrMissingExpectedValues, objectName)
		}

		rows, err := jsonquery.Convertor.ArrayToMap(records)
		if err != nil {
			return nil, err
		}

		return rows[0], nil
	case errors.Is(err, jsonquery.ErrNotArray):
		// Singleton data{} shape (e.g. customer-settings).
		node, err := jsonquery.New(body).ObjectOptional("data")
		if err != nil {
			return nil, err
		}

		if node == nil {
			return nil, fmt.Errorf("%w: missing \"data\" for %q",
				common.ErrMissingExpectedValues, objectName)
		}

		return jsonquery.Convertor.ObjectToMap(node)
	default:
		return nil, err
	}
}

// flattenRecord promotes the contents of a JSON:API record's "attributes" object
// to the top level, preserving other top-level keys (id, type). Attributes take
// precedence on key collisions, keeping the result deterministic regardless of
// map iteration order.
func flattenRecord(record map[string]any) map[string]any {
	out := make(map[string]any, len(record))

	for key, value := range record {
		if key == "attributes" {
			continue
		}

		out[key] = value
	}

	if attrs, ok := record["attributes"].(map[string]any); ok {
		maps.Copy(out, attrs)
	}

	return out
}
