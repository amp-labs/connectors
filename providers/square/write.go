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

// writeOperation is the create or update half of a writeConfig, resolved
// against the incoming WriteParams.
type writeOperation struct {
	method           string
	path             string
	envelopeKey      string
	needsIdempotency bool
	idInBody         bool
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	cfg, ok := writeObjects[params.ObjectName]
	if !ok {
		return nil, fmt.Errorf("%w: %q", common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	operation, err := resolveWriteOperation(cfg, params)
	if err != nil {
		return nil, err
	}

	urlParts := []string{apiVersion, operation.path}
	if params.IsUpdate() && !operation.idInBody {
		urlParts = append(urlParts, params.RecordId)
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, urlParts...)
	if err != nil {
		return nil, err
	}

	body, err := buildWriteBody(cfg, operation, params)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, operation.method, url.String(), bytes.NewReader(jsonData))
}

// resolveWriteOperation picks the create or update endpoint for the request,
// erroring when the object doesn't support that operation.
func resolveWriteOperation(cfg writeConfig, params common.WriteParams) (*writeOperation, error) {
	if params.IsUpdate() {
		if cfg.updatePath == "" {
			return nil, fmt.Errorf("%w: %q cannot be updated", common.ErrOperationNotSupportedForObject, params.ObjectName)
		}

		method := cfg.updateMethod
		if method == "" {
			method = http.MethodPut
		}

		return &writeOperation{
			method:           method,
			path:             cfg.updatePath,
			envelopeKey:      cfg.updateKey,
			needsIdempotency: cfg.updateNeedsIdempotency,
			idInBody:         cfg.updateIDInBody,
		}, nil
	}

	if cfg.createPath == "" {
		return nil, fmt.Errorf("%w: %q cannot be created", common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	return &writeOperation{
		method:           http.MethodPost,
		path:             cfg.createPath,
		envelopeKey:      cfg.createKey,
		needsIdempotency: cfg.createNeedsIdempotency,
	}, nil
}

// buildWriteBody shapes RecordData into the request body Square expects:
// topLevelFields and idempotency_key are hoisted out of the record to the top
// level, the rest is wrapped under the envelope key (or sent flat without one).
// When the endpoint requires an idempotency key and the caller didn't supply
// one, a UUID is generated so callers don't need to know Square's quirk.
func buildWriteBody(cfg writeConfig, operation *writeOperation, params common.WriteParams) (map[string]any, error) {
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
	} else if operation.needsIdempotency {
		body[idempotencyKeyField] = uuid.NewString()
	}

	if operation.idInBody && params.RecordId != "" {
		record["id"] = params.RecordId
	}

	if operation.envelopeKey != "" {
		body[operation.envelopeKey] = record
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

	cfg := writeObjects[params.ObjectName]

	// Square wraps the written record in an envelope, e.g. {"customer": {...}}.
	record := body

	if cfg.responseKey != "" {
		wrapped, err := jsonquery.New(body).ObjectOptional(cfg.responseKey)
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
