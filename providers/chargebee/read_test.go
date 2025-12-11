package chargebee

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseCustomers := testutils.DataFromFile(t, "customers.json")
	responseSubscriptions := testutils.DataFromFile(t, "subscriptions.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "customers"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Successful read of customers with chosen fields",
			Input: common.ReadParams{ObjectName: "customers", Fields: connectors.Fields("id", "first_name", "last_name", "email")}, //nolint:lll
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/customers"),
					mockcond.QueryParam("limit", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomers),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":         "__test__KyVnHhSBWlC1T2cj",
						"first_name": "John",
						"last_name":  "Doe",
						"email":      "john@test.com",
					},
					Raw: map[string]any{
						"allow_direct_debit":      false,
						"auto_collection":         "on",
						"card_status":             "no_card",
						"created_at":              float64(1517505747),
						"deleted":                 false,
						"email":                   "john@test.com",
						"excess_payments":         float64(0),
						"first_name":              "John",
						"id":                      "__test__KyVnHhSBWlC1T2cj",
						"last_name":               "Doe",
						"net_term_days":           float64(0),
						"object":                  "customer",
						"pii_cleared":             "active",
						"preferred_currency_code": "USD",
						"promotional_credits":     float64(0),
						"refundable_credits":      float64(0),
						"resource_version":        float64(1517505747000),
						"taxability":              "taxable",
						"unbilled_charges":        float64(0),
						"updated_at":              float64(1517505747),
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successful read of subscriptions with chosen fields",
			Input: common.ReadParams{ObjectName: "subscriptions", Fields: connectors.Fields("id", "customer_id", "status", "billing_period")}, //nolint:lll
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/subscriptions"),
					mockcond.QueryParam("limit", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseSubscriptions),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{ //nolint:dupl
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":             "__test__8asukSOXdvliPG",
						"customer_id":    "__test__8asukSOXdvg4PD",
						"status":         "active",
						"billing_period": float64(1),
					},
					Raw: map[string]any{
						"activated_at":             float64(1612890920),
						"billing_period":           float64(1),
						"billing_period_unit":      "month",
						"created_at":               float64(1612890920),
						"currency_code":            "USD",
						"current_term_end":         float64(1615310120),
						"current_term_start":       float64(1612890920),
						"customer_id":              "__test__8asukSOXdvg4PD",
						"deleted":                  false,
						"due_invoices_count":       float64(1),
						"due_since":                float64(1612890920),
						"has_scheduled_changes":    false,
						"id":                       "__test__8asukSOXdvliPG",
						"mrr":                      float64(0),
						"next_billing_at":          float64(1615310120),
						"object":                   "subscription",
						"remaining_billing_cycles": float64(1),
						"resource_version":         float64(1612890920000),
						"started_at":               float64(1612890920),
						"status":                   "active",
						"subscription_items": []any{
							map[string]any{
								"amount":         float64(1000),
								"billing_cycles": float64(1),
								"free_quantity":  float64(0),
								"item_price_id":  "basic-USD",
								"item_type":      "plan",
								"object":         "subscription_item",
								"quantity":       float64(1),
								"unit_price":     float64(1000),
							},
						},
						"total_dues": float64(1100),
						"updated_at": float64(1612890920),
					},
				}},
				NextPage: "[\"1612890918000\",\"230000000081\"]",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read of customers with incremental sync",
			Input: common.ReadParams{
				ObjectName: "customers",
				Fields:     connectors.Fields("id", "first_name", "email", "updated_at"),
				Since:      time.Unix(1517505747, 0).UTC(),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/customers"),
					mockcond.QueryParam("limit", "100"),
					mockcond.QueryParam("updated_at[after]", "1517505747"),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomers),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":         "__test__KyVnHhSBWlC1T2cj",
						"first_name": "John",
						"email":      "john@test.com",
						"updated_at": float64(1517505747),
					},
					Raw: map[string]any{
						"allow_direct_debit":      false,
						"auto_collection":         "on",
						"card_status":             "no_card",
						"created_at":              float64(1517505747),
						"deleted":                 false,
						"email":                   "john@test.com",
						"excess_payments":         float64(0),
						"first_name":              "John",
						"id":                      "__test__KyVnHhSBWlC1T2cj",
						"last_name":               "Doe",
						"net_term_days":           float64(0),
						"object":                  "customer",
						"pii_cleared":             "active",
						"preferred_currency_code": "USD",
						"promotional_credits":     float64(0),
						"refundable_credits":      float64(0),
						"resource_version":        float64(1517505747000),
						"taxability":              "taxable",
						"unbilled_charges":        float64(0),
						"updated_at":              float64(1517505747),
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read of subscriptions with incremental sync",
			Input: common.ReadParams{
				ObjectName: "subscriptions",
				Fields:     connectors.Fields("id", "customer_id", "status", "updated_at"),
				Since:      time.Unix(1612890920, 0).UTC(),
				Until:      time.Unix(1612890930, 0).UTC(),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/subscriptions"),
					mockcond.QueryParam("limit", "100"),
					mockcond.QueryParam("updated_at[after]", "1612890920"),
					mockcond.QueryParam("updated_at[before]", "1612890930"),
				},
				Then: mockserver.Response(http.StatusOK, responseSubscriptions),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{ //nolint:dupl
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":          "__test__8asukSOXdvliPG",
						"customer_id": "__test__8asukSOXdvg4PD",
						"status":      "active",
						"updated_at":  float64(1612890920),
					},
					Raw: map[string]any{
						"activated_at":             float64(1612890920),
						"billing_period":           float64(1),
						"billing_period_unit":      "month",
						"created_at":               float64(1612890920),
						"currency_code":            "USD",
						"current_term_end":         float64(1615310120),
						"current_term_start":       float64(1612890920),
						"customer_id":              "__test__8asukSOXdvg4PD",
						"deleted":                  false,
						"due_invoices_count":       float64(1),
						"due_since":                float64(1612890920),
						"has_scheduled_changes":    false,
						"id":                       "__test__8asukSOXdvliPG",
						"mrr":                      float64(0),
						"next_billing_at":          float64(1615310120),
						"object":                   "subscription",
						"remaining_billing_cycles": float64(1),
						"resource_version":         float64(1612890920000),
						"started_at":               float64(1612890920),
						"status":                   "active",
						"subscription_items": []any{
							map[string]any{
								"amount":         float64(1000),
								"billing_cycles": float64(1),
								"free_quantity":  float64(0),
								"item_price_id":  "basic-USD",
								"item_type":      "plan",
								"object":         "subscription_item",
								"quantity":       float64(1),
								"unit_price":     float64(1000),
							},
						},
						"total_dues": float64(1100),
						"updated_at": float64(1612890920),
					},
				}},
				NextPage: "[\"1612890918000\",\"230000000081\"]",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
