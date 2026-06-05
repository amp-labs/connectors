package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const apiVersion = "v25.0"

//nolint:gochecknoglobals
var (
	phoneNumberScopedWrite = datautils.NewStringSet(
		"block_users",
		"business_compliance_info",
		"deregister",
		"messages",
		"register",
		"request_code",
		"settings",
		"verify_code",
		"whatsapp_business_profile",
		"whatsapp_commerce_settings",
	)
	wabaScopedWrite = datautils.NewStringSet(
		"generate_payment_configuration_oauth_link",
		"message_templates",
		"payment_configurations",
		"phone_numbers",
	)

	ErrMissingWhatsAppAccountID = errors.New("whatsapp: whatsappAccountId metadata is required for this object")
	ErrMissingPhoneNumberID     = errors.New("whatsapp: whatsappPhoneNumberId metadata is required for this object")
)

func (a *Adapter) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := a.buildWriteURL(params)
	if err != nil {
		return nil, err
	}

	recordData, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(recordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// buildWriteURL routes phone-number-scoped and WABA-scoped objects under /v25.0/{id}/...
func (a *Adapter) buildWriteURL(params common.WriteParams) (*urlbuilder.URL, error) {
	baseURL := a.ModuleInfo().BaseURL

	switch {
	case phoneNumberScopedWrite.Has(params.ObjectName):
		if a.phoneNumberId == "" {
			return nil, ErrMissingPhoneNumberID
		}

		return urlbuilder.New(baseURL, apiVersion, a.phoneNumberId, params.ObjectName)

	case wabaScopedWrite.Has(params.ObjectName):
		if a.whatsappAccountId == "" {
			return nil, ErrMissingWhatsAppAccountID
		}

		return urlbuilder.New(baseURL, apiVersion, a.whatsappAccountId, params.ObjectName)

	default:
		return nil, common.ErrOperationNotSupportedForObject
	}
}

func (a *Adapter) parseWriteResponse(ctx context.Context, params common.WriteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	recordID, err := extractWriteRecordID(params.ObjectName, body)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}

func extractWriteRecordID(objectName string, body *ajson.Node) (string, error) {
	switch objectName {
	case "messages":
		arr, err := jsonquery.New(body).ArrayOptional("messages")
		if err != nil {
			return "", err
		}

		if len(arr) == 0 {
			return "", nil
		}

		return jsonquery.New(arr[0]).StrWithDefault("id", "")

	case "message_templates":
		return jsonquery.New(body).StrWithDefault("id", "")

	case "phone_numbers":
		id, err := jsonquery.New(body, "successful_creation", "value").StrWithDefault("id", "")
		if err != nil {
			return "", err
		}

		if id != "" {
			return id, nil
		}

		return jsonquery.New(body).StrWithDefault("id", "")

	default:
		return "", nil
	}
}
