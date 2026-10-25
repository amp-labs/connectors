package twenty

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

var (
	errDuplicateField = errors.New("field given twice under different cases")
	errWriteResponse  = errors.New("unexpected write response")
)

// buildWriteRequest creates with POST /rest/{namePlural} and updates with PATCH /rest/{namePlural}/{id},
// depth=0 so the response is the record in the shape Read returns: the OpenAPI document each instance serves
// at /rest/open-api/core (v2.50.0).
//
// The record may be built from ReadResult.Fields, whose keys are lowercase (renewalnotes), while Twenty
// accepts only its own camelCase names (renewalNotes; "doesn't have any \"renewalnotes\" field" otherwise),
// so keys are resolved against the object's schema first. Each write therefore reads every page of the
// workspace's object metadata before its create or update request.
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	record, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	objects, err := c.fetchObjects(ctx)
	if err != nil {
		return nil, err
	}

	// An object the schema does not list is sent as is; Twenty answers 400 for it. An inactive object is
	// still served on /rest (seen live), so its payload is resolved like any other.
	if object, ok := objects[params.ObjectName]; ok {
		record, err = object.writePayload(record)
		if err != nil {
			return nil, err
		}
	}

	method := http.MethodPost
	recordID := []string{}

	if params.IsUpdate() {
		method = http.MethodPatch

		recordID = append(recordID, params.RecordId)
	}

	writeURL, err := c.recordsURL(params.ObjectName, recordID...)
	if err != nil {
		return nil, err
	}

	writeURL.WithQueryParam("depth", "0")

	body, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, writeURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// writePayload renames each key to the Twenty field it names, ignoring case, and drops the fields Twenty
// manages itself (readOnly: id, createdAt, updatedAt, deletedAt, createdBy, updatedBy, searchVector).
// They are dropped rather than sent because Twenty stores what it is given: a PATCH carrying the updatedAt
// read earlier set updatedAt back to that value (seen live on v2.50.0), which would hide the edit from the
// next incremental read; and it refuses searchVector outright. Keys the schema does not know are sent as is.
func (o objectMetadata) writePayload(record map[string]any) (map[string]any, error) {
	names := o.keyIndex()
	payload := make(map[string]any, len(record))

	for key, value := range record {
		name, err := names.resolve(key)
		if err != nil {
			return nil, err
		}

		if field, ok := names.exact[name]; ok && field.readOnly() {
			continue
		}

		if _, ok := payload[name]; ok {
			return nil, fmt.Errorf("%w: %w: %s", common.ErrCaller, errDuplicateField, name)
		}

		payload[name] = value
	}

	return payload, nil
}

// resolve returns the Twenty name for a key: the only field it matches ignoring case, else the key
// unchanged. When the key matches several fields (renewalNotes and renewalnotes), only a key that carries
// capitals and names one of them exactly is taken: an all-lowercase key may come from ReadResult.Fields and
// stand for either, so it is refused rather than written to the wrong field.
func (n keyIndex) resolve(key string) (string, error) {
	candidates := n.byLowercase[lowercase(key)]

	switch {
	case len(candidates) == 0:
		return key, nil
	case len(candidates) == 1:
		return candidates[0], nil
	}

	if _, exact := n.exact[key]; exact && key != lowercase(key) {
		return key, nil
	}

	return "", fmt.Errorf("%w: %w: %s matches %v", common.ErrCaller, errCaseCollision, key, candidates)
}

// parseWriteResponse reads the record under data.create<Object> or data.update<Object>, the only key of data.
func (c *Connector) parseWriteResponse(
	_ context.Context,
	_ common.WriteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	data, err := jsonquery.New(body).ObjectRequired("data")
	if err != nil {
		return nil, err
	}

	keys := data.Keys()
	if len(keys) != 1 {
		return nil, fmt.Errorf("%w: data holds %v", errWriteResponse, keys)
	}

	node, err := jsonquery.New(data).ObjectRequired(keys[0])
	if err != nil {
		return nil, err
	}

	recordID, err := jsonquery.New(node).StringRequired("id")
	if err != nil {
		return nil, err
	}

	record, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     record,
	}, nil
}
