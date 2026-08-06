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

	// Everything is a POST except plain updates, which PUT to path/{id}.
	// Upsert endpoints (catalog) keep POSTing to upsertPath with the record id
	// in the body instead.
	path := cfg.path
	if cfg.upsertPath != "" {
		path = cfg.upsertPath
	}

	method := http.MethodPost
	urlParts := []string{apiVersion, path}

	if params.IsUpdate() && cfg.upsertPath == "" {
		method = http.MethodPut

		urlParts = append(urlParts, params.RecordId)
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, urlParts...)
	if err != nil {
		return nil, err
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
		if value, ok := record[field]; ok {
			body[field] = value
			delete(record, field)
		}
	}

	if value, ok := record[idempotencyKeyField]; ok {
		body[idempotencyKeyField] = value
		delete(record, idempotencyKeyField)
	} else if cfg.needsIdempotency {
		body[idempotencyKeyField] = uuid.NewString()
	}

	if cfg.upsertPath != "" && params.IsUpdate() && params.RecordId != "" {
		record["id"] = params.RecordId
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
