package phoneburner

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// API reference:
// https://www.phoneburner.com/developer/route_list

func buildWriteRequest(ctx context.Context, baseURL string, params common.WriteParams) (*http.Request, error) {
	if params.ObjectName == "" {
		return nil, common.ErrMissingObjects
	}

	url, method, err := buildWriteURL(baseURL, params)
	if err != nil {
		return nil, err
	}

	body, contentType, err := buildWriteBody(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return req, nil
}

func buildWriteURL(baseURL string, params common.WriteParams) (*urlbuilder.URL, string, error) {
	switch params.ObjectName {
	case "contacts", "members", "folders":
		if params.IsCreate() {
			u, err := urlbuilder.New(baseURL, restPrefix, restVer, params.ObjectName)
			return u, http.MethodPost, err
		}

		u, err := urlbuilder.New(baseURL, restPrefix, restVer, params.ObjectName, params.RecordId)
		return u, http.MethodPut, err
	case "dialsession":
		// Create only.
		if params.IsUpdate() {
			return nil, "", common.ErrOperationNotSupportedForObject
		}
		u, err := urlbuilder.New(baseURL, restPrefix, restVer, params.ObjectName)
		return u, http.MethodPost, err
	default:
		return nil, "", common.ErrOperationNotSupportedForObject
	}
}

func buildWriteBody(params common.WriteParams) ([]byte, string, error) {
	record, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, "", err
	}

	// PhoneBurner endpoints are mixed: some are JSON-body, others are form-url-encoded.
	switch params.ObjectName {
	case "dialsession", "folders":
		// JSON payload.
		b, err := json.Marshal(record)
		if err != nil {
			return nil, "", err
		}
		return b, "application/json", nil
	case "contacts", "members":
		b, err := httpkit.EncodeForm(record)
		if err != nil {
			return nil, "", err
		}
		return b, "application/x-www-form-urlencoded", nil
	default:
		return nil, "", common.ErrOperationNotSupportedForObject
	}
}

func parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	_ = ctx
	_ = request

	if err := interpretPhoneBurnerEnvelopeError(response); err != nil {
		return nil, err
	}

	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success:  true,
			RecordId: params.RecordId,
		}, nil
	}

	switch params.ObjectName {
	case "contacts":
		contacts := jsonquery.New(body, "contacts")

		contactNode, err := contacts.ObjectOptional("contacts")
		if err != nil || contactNode == nil {
			return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
		}

		recordID, err := jsonquery.New(contactNode).TextWithDefault("contact_user_id", params.RecordId)
		if err != nil {
			return nil, err
		}

		data, err := jsonquery.Convertor.ObjectToMap(contactNode)
		if err != nil {
			return nil, err
		}

		return &common.WriteResult{Success: true, RecordId: recordID, Data: data}, nil
	case "members":
		members := jsonquery.New(body, "members")

		memberNode, err := members.ObjectOptional("members")
		if err != nil || memberNode == nil {
			return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
		}

		recordID, err := jsonquery.New(memberNode).TextWithDefault("user_id", params.RecordId)
		if err != nil {
			return nil, err
		}

		data, err := jsonquery.Convertor.ObjectToMap(memberNode)
		if err != nil {
			return nil, err
		}

		return &common.WriteResult{Success: true, RecordId: recordID, Data: data}, nil
	case "folders":
		foldersNode, err := jsonquery.New(body).ObjectOptional("folders")
		if err != nil || foldersNode == nil {
			return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
		}

		m, err := jsonquery.Convertor.ObjectToMap(foldersNode)
		if err != nil {
			return nil, err
		}

		// Pick the first folder object (response is a map keyed by "0", "1", ...).
		for _, v := range m {
			obj, ok := v.(map[string]any)
			if !ok || obj == nil {
				continue
			}

			recordID, _ := obj["folder_id"].(string)
			if recordID == "" {
				recordID = params.RecordId
			}

			return &common.WriteResult{Success: true, RecordId: recordID, Data: obj}, nil
		}

		return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
	case "dialsession":
		dsNode, err := jsonquery.New(body).ObjectOptional("dialsessions")
		if err != nil || dsNode == nil {
			return &common.WriteResult{Success: true}, nil
		}

		data, err := jsonquery.Convertor.ObjectToMap(dsNode)
		if err != nil {
			return nil, err
		}

		return &common.WriteResult{Success: true, Data: data}, nil
	default:
		return nil, common.ErrOperationNotSupportedForObject
	}
}
