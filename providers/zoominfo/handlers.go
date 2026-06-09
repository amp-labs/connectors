package zoominfo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// searchRequestBody is the JSON:API envelope for a POST /{resource}/search call.
type searchRequestBody struct {
	Data searchRequestData `json:"data"`
}

type searchRequestData struct {
	Type       string         `json:"type"`
	Attributes map[string]any `json:"attributes"`
}

// buildSingleObjectMetadataRequest constructs a sampling request for one object.
// The request shape depends on the object kind:
//   - search: POST /gtm/data/v1/{resource}/search with a JSON:API body and page[size]=1
//   - lookup: GET  /gtm/data/v1/lookup/{fieldName}
func (c *Connector) buildSingleObjectMetadataRequest(
	ctx context.Context,
	objectName string,
) (*http.Request, error) {
	switch kindOf(objectName) {
	case kindSearch:
		return c.buildSearchMetadataRequest(ctx, objectName)
	case kindLookup:
		return c.buildLookupMetadataRequest(ctx, objectName)
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

	// Minimal JSON:API envelope with empty criteria. Objects that require search
	// criteria (e.g. contacts, intent) will return a descriptive 4xx, which the
	// schema provider records per-object in ListObjectMetadataResult.Errors.
	payload, err := json.Marshal(searchRequestBody{
		Data: searchRequestData{
			Type:       def.searchType,
			Attributes: map[string]any{},
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

func (c *Connector) buildLookupMetadataRequest(
	ctx context.Context,
	objectName string,
) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, dataAPIPath, "lookup", objectName)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", jsonAPIMediaType)

	return req, nil
}

// parseSingleObjectMetadataResponse infers an object's fields from the first
// record in a JSON:API "data" array. ZoomInfo nests record fields under
// "attributes"; those are promoted to the top level (mirroring how Read will
// flatten them) alongside top-level keys like "id" and "type".
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

	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	records, err := jsonquery.New(body).ArrayOptional("data")
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("%w: no records returned to sample fields for %q",
			common.ErrMissingExpectedValues, objectName)
	}

	rows, err := jsonquery.Convertor.ArrayToMap(records)
	if err != nil {
		return nil, err
	}

	for field, value := range flattenRecord(rows[0]) {
		metadata.Fields[field] = common.FieldMetadata{
			DisplayName: field,
			ValueType:   common.InferValueTypeFromData(value),
		}
	}

	return metadata, nil
}

// flattenRecord promotes the contents of a JSON:API record's "attributes" object
// to the top level, preserving other top-level keys (id, type). This matches the
// view a caller gets after the record is flattened during a read.
func flattenRecord(record map[string]any) map[string]any {
	out := make(map[string]any, len(record))

	for key, value := range record {
		if key == "attributes" {
			if attrs, ok := value.(map[string]any); ok {
				maps.Copy(out, attrs)
			}

			continue
		}

		out[key] = value
	}

	return out
}
