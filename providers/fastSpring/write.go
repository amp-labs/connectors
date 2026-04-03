package fastspring

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// FastSpring write API references:
// - Create account: https://developer.fastspring.com/reference/create-an-account
// - Update account: https://developer.fastspring.com/reference/update-an-account (POST /accounts/{account_id})
// - Create or update products: https://developer.fastspring.com/reference/create-or-update-products
// - Update order tags and attributes: https://developer.fastspring.com/reference/update-order-tags-and-attributes
// - Update subscription: https://developer.fastspring.com/reference/update-a-subscription

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if err := validateWriteParams(params); err != nil {
		return nil, err
	}

	baseURL := c.ProviderInfo().BaseURL

	body, err := buildWriteJSONBody(params)
	if err != nil {
		return nil, err
	}

	url, method, err := buildWriteURL(baseURL, params)
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

func validateWriteParams(params common.WriteParams) error {
	switch params.ObjectName {
	case "accounts", "products":
		return nil
	case "orders":
		if params.IsCreate() {
			return common.ErrOperationNotSupportedForObject
		}

		return nil
	case "subscriptions":
		if params.IsCreate() {
			return common.ErrOperationNotSupportedForObject
		}

		return nil
	default:
		return common.ErrOperationNotSupportedForObject
	}
}

func buildWriteURL(baseURL string, params common.WriteParams) (*urlbuilder.URL, string, error) {
	switch params.ObjectName {
	case "accounts":
		if params.IsCreate() {
			u, err := urlbuilder.New(baseURL, "accounts")

			return u, http.MethodPost, err
		}

		u, err := urlbuilder.New(baseURL, "accounts", params.RecordId)

		return u, http.MethodPost, err
	case "products":
		u, err := urlbuilder.New(baseURL, "products")

		return u, http.MethodPost, err
	case "orders":
		u, err := urlbuilder.New(baseURL, "orders")

		return u, http.MethodPost, err
	case "subscriptions":
		u, err := urlbuilder.New(baseURL, "subscriptions", params.RecordId)

		return u, http.MethodPost, err
	default:
		return nil, "", common.ErrOperationNotSupportedForObject
	}
}

func buildWriteJSONBody(params common.WriteParams) ([]byte, error) {
	record, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	switch params.ObjectName {
	case "accounts":
		return json.Marshal(record)
	case "products":
		return marshalProductsWriteBody(record)
	case "orders":
		return marshalOrdersWriteBody(params, record)
	case "subscriptions":
		return json.Marshal(record)
	default:
		return nil, common.ErrOperationNotSupportedForObject
	}
}

// marshalProductsWriteBody sends either a full bulk payload {"products":[...]} or wraps a single product object.
func marshalProductsWriteBody(record map[string]any) ([]byte, error) {
	if _, has := record["products"]; has {
		return json.Marshal(record)
	}

	return json.Marshal(map[string]any{"products": []any{record}})
}

// marshalOrdersWriteBody builds POST /orders for update order tags and attributes.
func marshalOrdersWriteBody(params common.WriteParams, record map[string]any) ([]byte, error) {
	if _, has := record["orders"]; has {
		return json.Marshal(record)
	}

	order := map[string]any{}
	for k, v := range record {
		order[k] = v
	}

	order["order"] = params.RecordId

	return json.Marshal(map[string]any{"orders": []any{order}})
}

func (c *Connector) parseWriteResponse(
	_ context.Context,
	params common.WriteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success:  true,
			RecordId: fallbackWriteRecordID(params),
		}, nil
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	recordID := extractWriteRecordID(params, data)
	if recordID == "" {
		recordID = fallbackWriteRecordID(params)
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}

func fallbackWriteRecordID(params common.WriteParams) string {
	if params.ObjectName == "products" && params.IsCreate() {
		return ""
	}

	return params.RecordId
}

func extractWriteRecordID(params common.WriteParams, data map[string]any) string {
	switch params.ObjectName {
	case "accounts":
		if s, ok := data["id"].(string); ok && s != "" {
			return s
		}

		if s, ok := data["account"].(string); ok && s != "" {
			return s
		}
	case "products":
		if arr, ok := data["products"].([]any); ok && len(arr) > 0 {
			if m, ok := arr[0].(map[string]any); ok {
				if s, ok := m["product"].(string); ok && s != "" {
					return s
				}
			}
		}
	case "orders":
		if arr, ok := data["orders"].([]any); ok && len(arr) > 0 {
			if m, ok := arr[0].(map[string]any); ok {
				if s, ok := m["order"].(string); ok && s != "" {
					return s
				}
			}
		}
	case "subscriptions":
		if s, ok := data["subscription"].(string); ok && s != "" {
			return s
		}

		if arr, ok := data["subscriptions"].([]any); ok && len(arr) > 0 {
			if m, ok := arr[0].(map[string]any); ok {
				if s, ok := m["subscription"].(string); ok && s != "" {
					return s
				}
			}
		}
	}

	return ""
}
