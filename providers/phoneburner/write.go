package phoneburner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

func buildWriteRequest(ctx context.Context, baseURL string, params common.WriteParams) (*http.Request, error) {
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
	case "contacts", "members", "folders", "customfields":
		if params.RecordId == "" {
			u, err := urlbuilder.New(baseURL, restPrefix, restVer, params.ObjectName)
			return u, http.MethodPost, err
		}

		u, err := urlbuilder.New(baseURL, restPrefix, restVer, params.ObjectName, params.RecordId)
		return u, http.MethodPut, err
	case "dialsession":
		// Create only.
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
	case "contacts", "members", "customfields":
		b, err := encodeForm(record)
		if err != nil {
			return nil, "", err
		}
		return b, "application/x-www-form-urlencoded", nil
	default:
		return nil, "", common.ErrOperationNotSupportedForObject
	}
}

func encodeForm(record map[string]any) ([]byte, error) {
	values := url.Values{}

	keys := make([]string, 0, len(record))
	for k := range record {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := record[k]
		if v == nil {
			continue
		}

		switch typed := v.(type) {
		case string:
			values.Set(k, typed)
		case []string:
			for _, item := range typed {
				values.Add(k, item)
			}
		case []any:
			for _, item := range typed {
				if item == nil {
					continue
				}
				switch itemTyped := item.(type) {
				case map[string]any, []any:
					b, err := json.Marshal(itemTyped)
					if err != nil {
						return nil, err
					}
					values.Add(k, string(b))
				default:
					values.Add(k, fmt.Sprint(item))
				}
			}
		case map[string]any:
			b, err := json.Marshal(typed)
			if err != nil {
				return nil, err
			}
			values.Set(k, string(b))
		default:
			values.Set(k, fmt.Sprint(v))
		}
	}

	return []byte(values.Encode()), nil
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
		contactsWrapper, err := jsonquery.New(body).ObjectOptional("contacts")
		if err != nil || contactsWrapper == nil {
			return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
		}

		contactNode, err := jsonquery.New(contactsWrapper).ObjectOptional("contacts")
		if err != nil || contactNode == nil {
			array, err2 := jsonquery.New(contactsWrapper).ArrayOptional("contacts")
			if err2 != nil || len(array) == 0 {
				return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
			}
			contactNode = array[0]
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
		membersWrapper, err := jsonquery.New(body).ObjectOptional("members")
		if err != nil || membersWrapper == nil {
			return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
		}

		memberNode, err := jsonquery.New(membersWrapper).ObjectOptional("members")
		if err != nil || memberNode == nil {
			array, err2 := jsonquery.New(membersWrapper).ArrayOptional("members")
			if err2 != nil || len(array) == 0 {
				return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
			}
			memberNode = array[0]
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
	case "customfields":
		wrapper, err := jsonquery.New(body).ObjectOptional("customfields")
		if err != nil || wrapper == nil {
			return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
		}

		array, err := jsonquery.New(wrapper).ArrayOptional("customfields")
		if err != nil || len(array) == 0 {
			return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
		}

		node := array[0]
		recordID, err := jsonquery.New(node).TextWithDefault("custom_field_id", params.RecordId)
		if err != nil {
			return nil, err
		}

		data, err := jsonquery.Convertor.ObjectToMap(node)
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
