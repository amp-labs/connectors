package recurly

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseCreateSubscription := testutils.DataFromFile(t, "create-subscription.json")
	responseUpdateSubscription := testutils.DataFromFile(t, "update-subscription.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "leads"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Successfully create of a Subscription",
			Input: common.WriteParams{
				ObjectName: "subscriptions",
				RecordData: map[string]any{
					"plan_code": "basic_plan",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/subscriptions"),
				},
				Then: mockserver.Response(http.StatusCreated, responseCreateSubscription),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "ierqn34o3hoife",
				Errors:   nil,
				Data: map[string]any{
					"id":                         "ierqn34o3hoife",
					"object":                     "subscription",
					"code":                       "test_account_001",
					"uuid":                       "550e8400-e29b-41d4-a716-446655440000",
					"plan_code":                  "basic_plan",
					"state":                      "active",
					"created_at":                 "2025-11-27T10:30:45Z",
					"updated_at":                 "2025-11-27T10:30:45Z",
					"activated_at":               "2025-11-27T10:30:45Z",
					"canceled_at":                nil,
					"expires_at":                 "2026-11-27T10:30:45Z",
					"bank_account_authorized_at": nil,
					"gateway_code":               "stripe",
					"billing_info_id":            "billing_9876543210",
					"active_invoice_id":          "inv_abc123def456",
					"business_entity_id":         "be_12345",
					"started_with_gift":          false,
					"converted_at":               nil,
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully Update subscription",
			Input: common.WriteParams{
				ObjectName: "subscriptions",
				RecordId:   "sub_2x3y4z5a6b7c8901",
				RecordData: map[string]any{
					"plan_code":         "premium_plan",
					"collection_method": "manual",
					"net_terms":         30,
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/subscriptions/sub_2x3y4z5a6b7c8901"),
				},
				Then: mockserver.Response(http.StatusOK, responseUpdateSubscription),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "sub_2x3y4z5a6b7c8901",
				Errors:   nil,
				Data: map[string]any{
					"id":                         "sub_2x3y4z5a6b7c8901",
					"object":                     "subscription",
					"uuid":                       "6ba7b811-9dad-11d1-80b4-00c04fd430c8",
					"plan_code":                  "premium_plan",
					"state":                      "active",
					"created_at":                 "2025-11-20T08:15:30Z",
					"updated_at":                 "2025-11-27T11:45:30Z",
					"activated_at":               "2025-11-20T08:15:30Z",
					"canceled_at":                nil,
					"expires_at":                 "2026-11-20T08:15:30Z",
					"bank_account_authorized_at": nil,
					"gateway_code":               "braintree",
					"billing_info_id":            "billing_premium_456",
					"active_invoice_id":          "inv_premium_789",
					"business_entity_id":         "be_67890",
					"started_with_gift":          false,
					"converted_at":               "2025-11-22T14:20:15Z",
					"collection_method":          "manual",
					"net_terms":                  float64(30),
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully create account",
			Input: common.WriteParams{
				ObjectName: "accounts",
				RecordData: map[string]any{
					"code":       "test_account_002",
					"first_name": "Jane",
					"last_name":  "Smith",
					"email":      "jane.smith@example.com",
					"company":    "Test Corp",
					"address": map[string]any{
						"street1":     "123 Main St",
						"city":        "Anytown",
						"region":      "CA",
						"postal_code": "12345",
						"country":     "US",
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/accounts"),
				},
				Then: mockserver.Response(http.StatusCreated, testutils.DataFromFile(t, "create-account.json")),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "acc_test_789xyz",
				Errors:   nil,
				Data: map[string]any{
					"id":                        "acc_test_789xyz",
					"object":                    "account",
					"state":                     "active",
					"hosted_login_token":        "hlt_456def789ghi",
					"has_live_subscription":     true,
					"has_active_subscription":   true,
					"has_future_subscription":   false,
					"has_canceled_subscription": false,
					"has_paused_subscription":   false,
					"has_past_due_invoice":      false,
					"created_at":                "2025-11-27T09:15:30Z",
					"updated_at":                "2025-11-27T09:15:30Z",
					"deleted_at":                nil,
					"code":                      "test_account_002",
					"username":                  "jane_smith",
					"email":                     "jane.smith@example.com",
					"first_name":                "Jane",
					"last_name":                 "Smith",
					"company":                   "Test Corp",
					"vat_number":                nil,
					"tax_exempt":                false,
					"entity_use_code":           nil,
					"bill_date":                 "2025-12-01T00:00:00Z",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully update account",
			Input: common.WriteParams{
				ObjectName: "accounts",
				RecordId:   "acc_67890",
				RecordData: map[string]any{
					"first_name": "Jane",
					"last_name":  "Doe",
					"company":    "Updated Corp",
					"email":      "jane.doe@example.com",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/accounts/acc_67890"),
				},
				Then: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "create-account.json")),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "acc_test_789xyz",
				Errors:   nil,
				Data: map[string]any{
					"id":                        "acc_test_789xyz",
					"object":                    "account",
					"state":                     "active",
					"hosted_login_token":        "hlt_456def789ghi",
					"has_live_subscription":     true,
					"has_active_subscription":   true,
					"has_future_subscription":   false,
					"has_canceled_subscription": false,
					"has_paused_subscription":   false,
					"has_past_due_invoice":      false,
					"created_at":                "2025-11-27T09:15:30Z",
					"updated_at":                "2025-11-27T09:15:30Z",
					"deleted_at":                nil,
					"code":                      "test_account_002",
					"username":                  "jane_smith",
					"email":                     "jane.smith@example.com",
					"first_name":                "Jane",
					"last_name":                 "Smith",
					"company":                   "Test Corp",
					"vat_number":                nil,
					"tax_exempt":                false,
					"entity_use_code":           nil,
					"bill_date":                 "2025-12-01T00:00:00Z",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create coupon via POST",
			Input: common.WriteParams{
				ObjectName: "coupons",
				RecordData: map[string]any{
					"code":                        "SAVE20",
					"name":                        "20% Off Sale",
					"coupon_type":                 "single_use",
					"max_redemptions":             100,
					"max_redemptions_per_account": 1,
					"hosted_page_description":     "Get 20% off your purchase",
					"invoice_description":         "20% Discount",
					"redeem_by":                   "2024-12-31T23:59:59Z",
					"discount": map[string]any{
						"type":    "percent",
						"percent": 20,
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/coupons"),
				},
				Then: mockserver.Response(http.StatusCreated, testutils.DataFromFile(t, "create-coupons.json")),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "coupon_bf2023_holiday",
				Errors:   nil,
				Data: map[string]any{
					"id":                          "coupon_bf2023_holiday",
					"object":                      "coupon",
					"code":                        "BLACKFRIDAY50",
					"name":                        "Black Friday 2023 - 50% Off",
					"state":                       "redeemable",
					"max_redemptions":             float64(500),
					"max_redemptions_per_account": float64(1),
					"unique_coupon_codes_count":   float64(1),
					"unique_code_template":        "HOLIDAY-{code}",
					"coupon_type":                 "single_code",
					"hosted_page_description":     "Save 50% on your first year with this exclusive Black Friday offer!",
					"invoice_description":         "Black Friday Promotion - 50% Discount",
					"redeem_by":                   "2023-12-31T23:59:59Z",
					"created_at":                  "2023-11-15T10:00:00Z",
					"updated_at":                  "2025-11-27T11:45:30Z",
					"expired_at":                  nil,
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
