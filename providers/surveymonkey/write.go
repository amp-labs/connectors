package surveymonkey

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/surveymonkey/metadata"
	"github.com/spyzhov/ajson"
)

// SurveyMonkey write API references:
// - POST /contacts, PATCH /contacts/{contact_id}
// - POST /contact_lists, PATCH /contact_lists/{contact_list_id}
// - POST /surveys, PATCH /surveys/{survey_id}
// - POST /survey_folders (create only)
// - PATCH /contact_fields/{contact_field_id} (update only).
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	if err := validateWriteOperation(params); err != nil {
		return nil, err
	}

	record, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	url, method, err := c.buildWriteURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Connector) buildWriteURL(params common.WriteParams) (*urlbuilder.URL, string, error) {
	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, "", err
	}

	path = strings.TrimSpace(path)

	if params.IsCreate() {
		url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)

		return url, http.MethodPost, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path, params.RecordId)

	return url, http.MethodPatch, err
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
			RecordId: params.RecordId,
		}, nil
	}

	recordID, data, err := parseWriteRecord(body, params.RecordId)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}

func parseWriteRecord(body *ajson.Node, fallbackID string) (string, map[string]any, error) {
	recordID, err := jsonquery.New(body).TextWithDefault("id", "")
	if err != nil {
		return "", nil, err
	}

	if recordID != "" {
		data, err := jsonquery.Convertor.ObjectToMap(body)
		if err != nil {
			return "", nil, err
		}

		return recordID, data, nil
	}

	records, err := jsonquery.New(body).ArrayOptional("data")
	if err != nil {
		return fallbackID, nil, err
	}

	if len(records) == 0 {
		return fallbackID, nil, nil
	}

	recordID, err = jsonquery.New(records[0]).TextWithDefault("id", fallbackID)
	if err != nil {
		return "", nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(records[0])
	if err != nil {
		return "", nil, err
	}

	return recordID, data, nil
}
