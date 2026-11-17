package mail

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/google/internal/mail/internal/imf"
)

var objectGroupWritableMessages = datautils.NewSet( // nolint:gochecknoglobals
	objectNameDrafts, virtualObjectInboxMessages,
)

func (a *Adapter) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := a.getWriteUrl(params.ObjectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	if params.IsUpdate() {
		url.AddPath(params.RecordId)

		if params.ObjectName == virtualObjectInboxMessages {
			return nil, common.ErrOperationNotSupportedForObject
		}

		method = http.MethodPut
	}

	reader, err := createPayload(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

// createPayload builds the request payload for a Gmail write operation.
// For regular objects, the source record is returned unchanged as the request body.
//
// For message-based objects, such as emails and drafts, Gmail requires the
// message content in Internet Message Format (IMF).
// Conversion and encoding are delegated to createEmailPayload.
func createPayload(params common.WriteParams) (*bytes.Reader, error) {
	if objectGroupWritableMessages.Has(params.ObjectName) {
		return createEmailPayload(params)
	}

	return params.GetRecordReader()
}

// createEmailPayload converts a Gmail message record into the JSON payload
// expected by a Gmail message write operation.
//
// If the record already contains a raw IMF message, it is assumed to already
// be base64url encoded and is used directly. Otherwise, FormatMessage builds
// the IMF representation from the record and this function encodes it.
func createEmailPayload(params common.WriteParams) (*bytes.Reader, error) {
	record, err := params.GetRecord()
	if err != nil {
		return nil, err
	}

	message, isEncoded, err := imf.FormatMessage(record)
	if err != nil {
		return nil, err
	}

	var raw string
	if isEncoded {
		raw = string(message)
	} else {
		raw = base64.RawURLEncoding.EncodeToString(message)
	}

	payload := make(map[string]any)
	if params.ObjectName == objectNameDrafts {
		payload["message"] = map[string]any{"raw": raw}
	} else {
		payload["raw"] = raw
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal email payload: %w", err)
	}

	return bytes.NewReader(jsonData), nil
}

func (a *Adapter) parseWriteResponse(ctx context.Context, params common.WriteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		// it is unlikely to have no payload
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
		Errors:   nil,
		Data:     data,
	}, nil
}
