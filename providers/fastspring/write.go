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
	case objectAccounts, objectProducts:
		return nil
	case objectOrders:
		if params.IsCreate() {
			return common.ErrOperationNotSupportedForObject
		}

		return nil
	case objectSubscriptions:
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
	case objectAccounts:
		if params.IsCreate() {
			u, err := urlbuilder.New(baseURL, objectAccounts)

			return u, http.MethodPost, err
		}

		u, err := urlbuilder.New(baseURL, objectAccounts, params.RecordId)

		return u, http.MethodPost, err
	case objectProducts:
		u, err := urlbuilder.New(baseURL, objectProducts)

		return u, http.MethodPost, err
	case objectOrders:
		u, err := urlbuilder.New(baseURL, objectOrders)

		return u, http.MethodPost, err
	case objectSubscriptions:
		u, err := urlbuilder.New(baseURL, objectSubscriptions, params.RecordId)

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
	case objectAccounts:
		return json.Marshal(record)
	case objectProducts:
		return marshalProductsWriteBody(record)
	case objectOrders:
		return marshalOrdersWriteBody(params, record)
	case objectSubscriptions:
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
	if params.ObjectName == objectProducts && params.IsCreate() {
		return ""
	}

	return params.RecordId
}

func extractWriteRecordID(params common.WriteParams, data map[string]any) string {
	switch params.ObjectName {
	case objectAccounts:
		return extractAccountWriteRecordID(data)
	case objectProducts:
		return extractProductWriteRecordID(data)
	case objectOrders:
		return extractOrderWriteRecordID(data)
	case objectSubscriptions:
		return extractSubscriptionWriteRecordID(data)
	default:
		return ""
	}
}

func stringFieldFromMap(m map[string]any, key string) string {
	s, ok := m[key].(string)
	if !ok || s == "" {
		return ""
	}

	return s
}

func extractAccountWriteRecordID(data map[string]any) string {
	if s := stringFieldFromMap(data, "id"); s != "" {
		return s
	}

	return stringFieldFromMap(data, "account")
}

func extractRecordIDFromFirstSliceElement(data map[string]any, sliceKey, fieldName string) string {
	arr, ok := data[sliceKey].([]any)
	if !ok || len(arr) == 0 {
		return ""
	}

	m, ok := arr[0].(map[string]any)
	if !ok {
		return ""
	}

	return stringFieldFromMap(m, fieldName)
}

func extractProductWriteRecordID(data map[string]any) string {
	return extractRecordIDFromFirstSliceElement(data, "products", "product")
}

func extractOrderWriteRecordID(data map[string]any) string {
	return extractRecordIDFromFirstSliceElement(data, "orders", "order")
}

func extractSubscriptionWriteRecordID(data map[string]any) string {
	if s := stringFieldFromMap(data, "subscription"); s != "" {
		return s
	}

	return extractRecordIDFromFirstSliceElement(data, "subscriptions", "subscription")
}
