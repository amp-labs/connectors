package fastspring

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
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
	maps.Copy(order, record)

	// API uses the JSON key "order" for the order id (not e.g. "id"); see payload shape:
	// https://developer.fastspring.com/reference/update-order-tags-and-attributes
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

	recordID, ambiguousMultiple := extractWriteRecordID(params, body)
	if recordID == "" && !ambiguousMultiple {
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
	if params.ObjectName == objectProducts && params.IsCreate() {
		return ""
	}

	return params.RecordId
}

// extractWriteRecordID returns a response record id when it identifies exactly one record.
// ambiguousMultiple is true when the relevant array in the payload has more than one element,
// in which case the id is left empty and the caller must not substitute fallbackWriteRecordID.
func extractWriteRecordID(params common.WriteParams, body *ajson.Node) (recordID string, ambiguousMultiple bool) {
	switch params.ObjectName {
	case objectAccounts:
		return extractAccountWriteRecordID(body)
	case objectProducts:
		return extractProductWriteRecordID(body)
	case objectOrders:
		return extractOrderWriteRecordID(body)
	case objectSubscriptions:
		return extractSubscriptionWriteRecordID(body)
	default:
		return "", false
	}
}

func extractAccountWriteRecordID(body *ajson.Node) (recordID string, ambiguousMultiple bool) {
	if id, err := jsonquery.New(body).TextWithDefault("id", ""); err == nil && id != "" {
		return id, false
	}

	if id, err := jsonquery.New(body).TextWithDefault("account", ""); err == nil && id != "" {
		return id, false
	}

	return "", false
}

func firstArrayElementTextField(body *ajson.Node, arrayKey, fieldKey string) (recordID string, ambiguousMultiple bool) {
	arr, err := jsonquery.New(body).ArrayOptional(arrayKey)
	if err != nil || len(arr) == 0 {
		return "", false
	}

	if len(arr) > 1 {
		return "", true
	}

	id, err := jsonquery.New(arr[0]).TextWithDefault(fieldKey, "")
	if err != nil {
		return "", false
	}

	return id, false
}

func extractProductWriteRecordID(body *ajson.Node) (recordID string, ambiguousMultiple bool) {
	return firstArrayElementTextField(body, "products", "product")
}

func extractOrderWriteRecordID(body *ajson.Node) (recordID string, ambiguousMultiple bool) {
	return firstArrayElementTextField(body, "orders", "order")
}

func extractSubscriptionWriteRecordID(body *ajson.Node) (recordID string, ambiguousMultiple bool) {
	if id, err := jsonquery.New(body).TextWithDefault("subscription", ""); err == nil && id != "" {
		return id, false
	}

	return firstArrayElementTextField(body, "subscriptions", "subscription")
}
