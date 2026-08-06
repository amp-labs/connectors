package square

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
	"github.com/google/uuid"
)

const idempotencyKeyField = "idempotency_key"

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	cfg, ok := objects[params.ObjectName]
	if !ok || !cfg.supportsWrite {
		return nil, fmt.Errorf("%w: %q", common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	// Writes POST, except updates, which PUT to path/{id} — unless the
	// endpoint is an upsert (upsertPath set), which always POSTs.
	path := cfg.path
	if cfg.upsertPath != "" {
		path = cfg.upsertPath
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	if params.IsUpdate() && cfg.upsertPath == "" {
		method = http.MethodPut

		url.AddPath(params.RecordId)
	}

	body, err := buildWriteBody(cfg, params)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

// buildWriteBody shapes RecordData into the request body Square expects:
// topLevelFields and idempotency_key are hoisted out of the record to the top
// level, the rest is wrapped under writeKey (or sent flat without one).
// When the endpoint requires an idempotency key and the caller didn't supply
// one, a UUID is generated so callers don't need to know Square's quirk.
func buildWriteBody(cfg objectConfig, params common.WriteParams) (map[string]any, error) {
	record, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	// RecordDataToMap may return the caller's own map; clone before mutating.
	record = maps.Clone(record)
	body := make(map[string]any)

	for _, field := range cfg.topLevelFields {
		value, ok := record[field]
		// If the caller provided a value for a top-level field,
		// Use it and remove it from the record so it doesn't get wrapped under writeKey.
		if ok {
			body[field] = value
			delete(record, field)
		}
	}

	value, ok := record[idempotencyKeyField]
	if ok {
		// Caller provided an idempotency key; use it and remove it from the record.
		body[idempotencyKeyField] = value
		delete(record, idempotencyKeyField)
	} else if cfg.needsIdempotency {
		// Caller didn't provide an idempotency key, but the endpoint requires one; generate a UUID.
		body[idempotencyKeyField] = uuid.NewString()
	}

	envelopeKey := cfg.writeKey
	if !params.IsUpdate() && cfg.flatCreate {
		envelopeKey = ""
	}

	if envelopeKey != "" {
		body[envelopeKey] = record
	} else {
		maps.Copy(body, record)
	}

	return body, nil
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{Success: true}, nil
	}

	cfg := objects[params.ObjectName]

	// Square wraps the written record in an envelope, e.g. {"customer": {...}}.
	record := body

	responseKey := cfg.writeResponseKey
	if responseKey == "" {
		responseKey = cfg.writeKey
	}

	if responseKey != "" {
		wrapped, err := jsonquery.New(body).ObjectOptional(responseKey)
		if err != nil {
			return nil, err
		}

		if wrapped != nil {
			record = wrapped
		}
	}

	idField := cfg.idField
	if idField == "" {
		idField = "id"
	}

	recordID, err := jsonquery.New(record).StrWithDefault(idField, params.RecordId)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(record)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}
