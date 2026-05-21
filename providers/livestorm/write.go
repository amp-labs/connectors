package livestorm

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// Livestorm write API references:
// - Create event: https://developers.livestorm.co/reference/post_events
// - Update event: https://developers.livestorm.co/reference/patch_events-id
// - Create user (team member): https://developers.livestorm.co/reference/post_users
//
// Session bulk registrants (POST …/sessions/{id}/people/bulk) are not implemented on Write;
// that flow belongs with a future Bulk API surface (see e.g. providers/salesforce/bulk-write.go).
const (
	objectUsers               = "users"
	jsonAPIContentType        = "application/vnd.api+json"
	jsonAPIResourceTypeEvents = "events"
	jsonAPIResourceTypeUsers  = "users"
)

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if err := validateLivestormWrite(params); err != nil {
		return nil, err
	}

	record, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	body, err := buildLivestormWriteBody(params, record)
	if err != nil {
		return nil, err
	}

	u, method, err := c.buildWriteURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", jsonAPIContentType)
	req.Header.Set("Content-Type", jsonAPIContentType)

	return req, nil
}

func validateLivestormWrite(params common.WriteParams) error {
	if err := params.ValidateParams(); err != nil {
		return err
	}

	switch params.ObjectName {
	case objectEvents:
		return nil
	case objectUsers:
		if params.IsUpdate() {
			return common.ErrOperationNotSupportedForObject
		}

		return nil
	default:
		return common.ErrOperationNotSupportedForObject
	}
}

func (c *Connector) buildWriteURL(params common.WriteParams) (*urlbuilder.URL, string, error) {
	base := c.ProviderInfo().BaseURL

	switch params.ObjectName {
	case objectEvents:
		if params.IsCreate() {
			u, err := urlbuilder.New(base, apiVersion, objectEvents)

			return u, http.MethodPost, err
		}

		u, err := urlbuilder.New(base, apiVersion, objectEvents, params.RecordId)

		return u, http.MethodPatch, err
	case objectUsers:
		u, err := urlbuilder.New(base, apiVersion, objectUsers)

		return u, http.MethodPost, err
	default:
		return nil, "", common.ErrOperationNotSupportedForObject
	}
}

func buildLivestormWriteBody(params common.WriteParams, record map[string]any) ([]byte, error) {
	switch params.ObjectName {
	case objectEvents:
		return marshalEventsWriteBody(params, record)
	case objectUsers:
		return marshalUsersWriteBody(record)
	default:
		return nil, common.ErrOperationNotSupportedForObject
	}
}

func marshalEventsWriteBody(params common.WriteParams, record map[string]any) ([]byte, error) {
	if _, has := record["data"]; has {
		return json.Marshal(record)
	}

	delete(record, "id")

	payload := map[string]any{
		"data": map[string]any{
			"type":       jsonAPIResourceTypeEvents,
			"attributes": record,
		},
	}

	if params.IsUpdate() {
		data, ok := payload["data"].(map[string]any)
		if !ok {
			return nil, common.ErrBadRequest
		}

		data["id"] = params.RecordId
	}

	return json.Marshal(payload)
}

func marshalUsersWriteBody(record map[string]any) ([]byte, error) {
	if _, has := record["data"]; has {
		return json.Marshal(record)
	}

	return json.Marshal(map[string]any{
		"data": map[string]any{
			"type":       jsonAPIResourceTypeUsers,
			"attributes": record,
		},
	})
}

func (c *Connector) parseWriteResponse(
	_ context.Context,
	params common.WriteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok || body == nil {
		return &common.WriteResult{
			Success:  true,
			RecordId: fallbackWriteRecordID(params),
		}, nil
	}

	recordID := extractJSONAPIResourceID(body)
	if recordID == "" {
		recordID = fallbackWriteRecordID(params)
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}

func fallbackWriteRecordID(params common.WriteParams) string {
	if params.ObjectName == objectEvents && params.IsCreate() {
		return ""
	}

	return params.RecordId
}

// extractJSONAPIResourceID reads id from JSON:API data.id or a root-level id (some job responses).
func extractJSONAPIResourceID(body *ajson.Node) string {
	if id, err := jsonquery.New(body, "data").StringOptional("id"); err == nil && id != nil && *id != "" {
		return *id
	}

	if id, err := jsonquery.New(body).StringOptional("id"); err == nil && id != nil && *id != "" {
		return *id
	}

	return ""
}
