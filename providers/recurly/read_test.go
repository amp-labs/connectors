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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseAccountsFirst := testutils.DataFromFile(t, "read-accounts.json")
	responseAccountsLast := testutils.DataFromFile(t, "read-accounts-last.json")
	responseInvoicesFirst := testutils.DataFromFile(t, "read-invoices.json")
	responseInvoicesLast := testutils.DataFromFile(t, "read-invoices-last.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "accounts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Successful read of accounts with chosen fields",
			Input: common.ReadParams{ObjectName: "accounts", Fields: connectors.Fields("id", "code", "company")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/accounts"),
				Then:  mockserver.Response(http.StatusOK, responseAccountsFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":      "string",
						"code":    "string",
						"company": "string",
					},
					Raw: map[string]any{
						"id":                        "string",
						"object":                    "string",
						"state":                     "active",
						"hosted_login_token":        "string",
						"has_live_subscription":     true,
						"has_active_subscription":   true,
						"has_future_subscription":   true,
						"has_canceled_subscription": true,
						"has_paused_subscription":   true,
						"has_past_due_invoice":      true,
						"created_at":                "2019-08-24T14:15:22Z",
						"updated_at":                "2019-08-24T14:15:22Z",
						"deleted_at":                "2019-08-24T14:15:22Z",
						"code":                      "string",
						"username":                  "string",
						"company":                   "string",
						"vat_number":                "string",
						"tax_exempt":                true,
						"entity_use_code":           "string",
						"bill_date":                 "2019-08-24T14:15:22Z",
					},
				}},
				NextPage: "/accounts?cursor=string",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successful read of invoices with chosen fields",
			Input: common.ReadParams{ObjectName: "invoices", Fields: connectors.Fields("id", "uuid", "state")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/invoices"),
				Then:  mockserver.Response(http.StatusOK, responseInvoicesFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    "string",
						"uuid":  "string",
						"state": "all",
					},
					Raw: map[string]any{
						"id":                  "string",
						"uuid":                "string",
						"object":              "string",
						"type":                "charge",
						"origin":              "carryforward_credit",
						"state":               "all",
						"created_at":          "2019-08-24T14:15:22Z",
						"updated_at":          "2019-08-24T14:15:22Z",
						"due_at":              "2019-08-24T14:15:22Z",
						"closed_at":           "2019-08-24T14:15:22Z",
						"dunning_campaign_id": "string",
						"dunning_events_sent": float64(0),
						"final_dunning_event": true,
						"business_entity_id":  "string",
					},
				}},
				NextPage: "/invoices?cursor=string",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Next page is missing for the last accounts page",
			Input: common.ReadParams{ObjectName: "accounts", Fields: connectors.Fields("id"), NextPage: common.NextPageToken("/accounts?cursor=string")}, //nolint:lll
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/accounts"),
					mockcond.QueryParam("cursor", "string"),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountsLast),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": "string",
					},
					Raw: map[string]any{
						"id":      "string",
						"object":  "string",
						"state":   "active",
						"code":    "string",
						"company": "string",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Next page is missing for the last invoices page",
			Input: common.ReadParams{ObjectName: "invoices", Fields: connectors.Fields("id"), NextPage: common.NextPageToken("/invoices?cursor=string")}, //nolint:lll
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/invoices"),
					mockcond.QueryParam("cursor", "string"),
				},
				Then: mockserver.Response(http.StatusOK, responseInvoicesLast),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": "string",
					},
					Raw: map[string]any{
						"id":     "string",
						"uuid":   "string",
						"object": "string",
						"state":  "all",
					},
				}},
				NextPage: "",
				Done:     true,
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
