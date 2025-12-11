package paddle

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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	createCustomerResponse := testutils.DataFromFile(t, "create-customer.json")
	createDiscountResponse := testutils.DataFromFile(t, "create-discount.json")

	tests := []testroutines.Write{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "customers"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Successfully creation of a customer",
			Input: common.WriteParams{
				ObjectName: "customers",
				RecordData: map[string]any{
					"name":  "Jo Brown",
					"email": "jo@example.com",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/customers"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, createCustomerResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "ctm_01hv6y1jedq4p1n0yqn5ba3ky4",
				Data: map[string]any{
					"id":                "ctm_01hv6y1jedq4p1n0yqn5ba3ky4",
					"status":            "active",
					"custom_data":       nil,
					"name":              "Jo Brown",
					"email":             "jo@example.com",
					"marketing_consent": false,
					"locale":            "en",
					"created_at":        "2024-04-11T15:57:24.813Z",
					"updated_at":        "2024-04-11T15:57:24.813Z",
					"import_meta":       nil,
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update customer as PATCH",
			Input: common.WriteParams{
				ObjectName: "customers",
				RecordId:   "ctm_01hv6y1jedq4p1n0yqn5ba3ky4",
				RecordData: map[string]any{
					"name": "Jo Brown Updated",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/customers/ctm_01hv6y1jedq4p1n0yqn5ba3ky4"),
					mockcond.MethodPATCH(),
				},
				Then: mockserver.Response(http.StatusOK, createCustomerResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "ctm_01hv6y1jedq4p1n0yqn5ba3ky4",
				Data: map[string]any{
					"id":                "ctm_01hv6y1jedq4p1n0yqn5ba3ky4",
					"status":            "active",
					"custom_data":       nil,
					"name":              "Jo Brown",
					"email":             "jo@example.com",
					"marketing_consent": false,
					"locale":            "en",
					"created_at":        "2024-04-11T15:57:24.813Z",
					"updated_at":        "2024-04-11T15:57:24.813Z",
					"import_meta":       nil,
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully creation of a discount",
			Input: common.WriteParams{
				ObjectName: "discounts",
				RecordData: map[string]any{
					"description":                 "All orders (10% off)",
					"enabled_for_checkout":        true,
					"code":                        "BF10OFF",
					"type":                        "percentage",
					"amount":                      "10",
					"recur":                       true,
					"maximum_recurring_intervals": 3,
					"expires_at":                  "2024-12-03T00:00:00Z",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/discounts"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, createDiscountResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "dsc_01hv6scyf7qdnzcdq01t2y8dx4",
				Data: map[string]any{
					"id":                          "dsc_01hv6scyf7qdnzcdq01t2y8dx4",
					"status":                      "active",
					"description":                 "All orders (10% off)",
					"enabled_for_checkout":        true,
					"code":                        "BF10OFF",
					"type":                        "percentage",
					"mode":                        "standard",
					"amount":                      "10",
					"currency_code":               nil,
					"recur":                       true,
					"maximum_recurring_intervals": float64(3),
					"usage_limit":                 nil,
					"restrict_to":                 nil,
					"expires_at":                  "2024-12-03T00:00:00Z",
					"times_used":                  float64(0),
					"discount_group_id":           "dsg_01js2gqehzccfkywgx1jk2mtsp",
					"custom_data":                 nil,
					"import_meta":                 nil,
					"created_at":                  "2024-11-28T14:36:14.695Z",
					"updated_at":                  "2024-11-28T14:36:14.695Z",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update discount as PATCH",
			Input: common.WriteParams{
				ObjectName: "discounts",
				RecordId:   "dsc_01hv6scyf7qdnzcdq01t2y8dx4",
				RecordData: map[string]any{
					"description": "Updated discount description",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/discounts/dsc_01hv6scyf7qdnzcdq01t2y8dx4"),
					mockcond.MethodPATCH(),
				},
				Then: mockserver.Response(http.StatusOK, createDiscountResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "dsc_01hv6scyf7qdnzcdq01t2y8dx4",
				Data: map[string]any{
					"id":                          "dsc_01hv6scyf7qdnzcdq01t2y8dx4",
					"status":                      "active",
					"description":                 "All orders (10% off)",
					"enabled_for_checkout":        true,
					"code":                        "BF10OFF",
					"type":                        "percentage",
					"mode":                        "standard",
					"amount":                      "10",
					"currency_code":               nil,
					"recur":                       true,
					"maximum_recurring_intervals": float64(3),
					"usage_limit":                 nil,
					"restrict_to":                 nil,
					"expires_at":                  "2024-12-03T00:00:00Z",
					"times_used":                  float64(0),
					"discount_group_id":           "dsg_01js2gqehzccfkywgx1jk2mtsp",
					"custom_data":                 nil,
					"import_meta":                 nil,
					"created_at":                  "2024-11-28T14:36:14.695Z",
					"updated_at":                  "2024-11-28T14:36:14.695Z",
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
