package fastspring

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	accountCreateResp := []byte(`{"id":"acc_new","account":"acc_new","action":"account.create","result":"success"}`)
	accountUpdateResp := []byte(`{"id":"acc_1","account":"acc_1","action":"account.update","result":"success"}`)
	productResp := []byte(`{"products":[{"product":"my-product","action":"product.create","result":"success"}]}`)
	productBulkResp := []byte(`{"products":[{"product":"p-a","action":"product.create","result":"success"},{"product":"p-b","action":"product.create","result":"success"}]}`)
	orderResp := []byte(`{"orders":[{"order":"ord_1","action":"order.update","result":"success"}]}`)
	subscriptionResp := []byte(`{"subscription":"sub_1","action":"subscription.update","result":"success"}`)

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "accounts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.WriteParams{ObjectName: "events-processed", RecordData: map[string]any{"x": 1}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Orders do not support create",
			Input: common.WriteParams{
				ObjectName: "orders",
				RecordData: map[string]any{"tags": map[string]any{"k": "v"}},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Subscriptions do not support create",
			Input: common.WriteParams{
				ObjectName: "subscriptions",
				RecordData: map[string]any{"state": "active"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create account",
			Input: common.WriteParams{
				ObjectName: "accounts",
				RecordData: map[string]any{
					"contact": map[string]any{"first": "A", "last": "B", "email": "a@example.com"},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/accounts"),
					mockcond.Body(`{"contact":{"email":"a@example.com","first":"A","last":"B"}}`),
				},
				Then: mockserver.Response(http.StatusOK, accountCreateResp),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "acc_new",
				Data: map[string]any{
					"id":      "acc_new",
					"account": "acc_new",
					"action":  "account.create",
					"result":  "success",
				},
			},
		},
		{
			Name: "Update account",
			Input: common.WriteParams{
				ObjectName: "accounts",
				RecordId:   "acc_1",
				RecordData: map[string]any{"language": "en"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/accounts/acc_1"),
					mockcond.Body(`{"language":"en"}`),
				},
				Then: mockserver.Response(http.StatusOK, accountUpdateResp),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "acc_1",
				Data: map[string]any{
					"id":      "acc_1",
					"account": "acc_1",
					"action":  "account.update",
					"result":  "success",
				},
			},
		},
		{
			Name: "Create product wraps single object in products array",
			Input: common.WriteParams{
				ObjectName: "products",
				RecordData: map[string]any{
					"product": "my-product",
					"display": map[string]any{"en": "Widget"},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/products"),
					mockcond.Body(mustJSON(t, map[string]any{
						"products": []any{
							map[string]any{
								"product": "my-product",
								"display": map[string]any{"en": "Widget"},
							},
						},
					})),
				},
				Then: mockserver.Response(http.StatusOK, productResp),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "my-product",
				Data: map[string]any{
					"products": []any{
						map[string]any{
							"product": "my-product",
							"action":  "product.create",
							"result":  "success",
						},
					},
				},
			},
		},
		{
			Name: "Create product with multiple products in response omits record id",
			Input: common.WriteParams{
				ObjectName: "products",
				RecordData: map[string]any{
					"product": "bulk-a",
					"display": map[string]any{"en": "A"},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/products"),
					mockcond.Body(mustJSON(t, map[string]any{
						"products": []any{
							map[string]any{
								"product": "bulk-a",
								"display": map[string]any{"en": "A"},
							},
						},
					})),
				},
				Then: mockserver.Response(http.StatusOK, productBulkResp),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "",
				Data: map[string]any{
					"products": []any{
						map[string]any{
							"product": "p-a",
							"action":  "product.create",
							"result":  "success",
						},
						map[string]any{
							"product": "p-b",
							"action":  "product.create",
							"result":  "success",
						},
					},
				},
			},
		},
		{
			Name: "Update order tags",
			Input: common.WriteParams{
				ObjectName: "orders",
				RecordId:   "ord_1",
				RecordData: map[string]any{"tags": map[string]any{"k1": "v1"}},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/orders"),
					mockcond.Body(`{"orders":[{"order":"ord_1","tags":{"k1":"v1"}}]}`),
				},
				Then: mockserver.Response(http.StatusOK, orderResp),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "ord_1",
				Data: map[string]any{
					"orders": []any{
						map[string]any{
							"order":  "ord_1",
							"action": "order.update",
							"result": "success",
						},
					},
				},
			},
		},
		{
			Name: "Update subscription",
			Input: common.WriteParams{
				ObjectName: "subscriptions",
				RecordId:   "sub_1",
				RecordData: map[string]any{"product": "my-app"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/subscriptions/sub_1"),
					mockcond.Body(`{"product":"my-app"}`),
				},
				Then: mockserver.Response(http.StatusOK, subscriptionResp),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "sub_1",
				Data: map[string]any{
					"subscription": "sub_1",
					"action":       "subscription.update",
					"result":       "success",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()

	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	return string(b)
}
