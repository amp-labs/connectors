package zoominfo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// writeRequestBody is the JSON:API envelope for create/update. id is included
// only for upsert-style updates (omitted on create).
type writeRequestBody struct {
	Data writeRequestData `json:"data"`
}

type writeRequestData struct {
	Type       string `json:"type"`
	ID         string `json:"id,omitempty"`
	Attributes any    `json:"attributes"`
}

// buildWriteRequest creates or updates a writable object. The request shape
// depends on the object's write style:
//   - upsert (Copilot config): POST {collection}; an existing RecordId is carried
//     in the JSON:API body as data.id.
//   - createUpdate (Studio): POST {collection} to create; PATCH {collection}/{id}
//     to update.
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	def, ok := writeObjects[params.ObjectName]
	if !ok {
		return nil, fmt.Errorf("%w: %q", common.ErrObjectNotSupported, params.ObjectName)
	}

	method := http.MethodPost
	segments := def.segments
	body := writeRequestData{Type: def.recordType, Attributes: params.RecordData}

	if params.RecordId != "" {
		switch def.style {
		case styleUpsert:
			// Same collection endpoint; the id travels in the body.
			body.ID = params.RecordId
		case styleCreateUpdate:
			method = http.MethodPatch

			segments = append(append([]string{}, def.segments...), params.RecordId)
		}
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, segments...)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(writeRequestBody{Data: body})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", jsonAPIMediaType)
	req.Header.Set("Content-Type", jsonAPIMediaType)

	return req, nil
}

// parseWriteResponse reads the created/updated resource's id (and data) from the
// JSON:API response. Some updates may return 204 with no body, in which case the
// caller's RecordId is echoed back.
func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
	}

	recordID, err := jsonquery.New(body, "data").StrWithDefault("id", params.RecordId)
	if err != nil {
		return nil, err
	}

	dataNode, err := jsonquery.New(body).ObjectOptional("data")
	if err != nil {
		return nil, err
	}

	var data map[string]any
	if dataNode != nil {
		if data, err = jsonquery.Convertor.ObjectToMap(dataNode); err != nil {
			return nil, err
		}
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}
