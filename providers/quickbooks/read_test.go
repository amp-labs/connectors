package quickbooks

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

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseAccount := testutils.DataFromFile(t, "account-read.json")
	responseAccountWithNextPage := testutils.DataFromFile(t, "account-read-with-next-page.json")
	responseAccountEmpty := testutils.DataFromFile(t, "account-read-empty.json")
	responseCustomer := testutils.DataFromFile(t, "customer-read.json")
	responseCustomerWithCustomFields := testutils.DataFromFile(t, "custom-fields/customer-read-with-custom-fields.json")
	responseItem := testutils.DataFromFile(t, "item-read.json")
	responseError := testutils.DataFromFile(t, "error-bad-request.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "users"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},

		{
			Name:  "Successful read with chosen fields",
			Input: common.ReadParams{ObjectName: "account", Fields: connectors.Fields("accountsubtype", "accounttype", "active")}, //nolint:lll
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.QueryParam("query", "SELECT * FROM Account STARTPOSITION 1 MAXRESULTS 1000"),
				Then:  mockserver.Response(http.StatusOK, responseAccount),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"accountsubtype": "AccountsReceivable",
						"accounttype":    "Accounts Receivable",
						"active":         true,
					},
					Raw: map[string]any{
						"AccountSubType":     "AccountsReceivable",
						"AccountType":        "Accounts Receivable",
						"Active":             true,
						"Classification":     "Asset",
						"domain":             "QBO",
						"sparse":             false,
						"FullyQualifiedName": "Canadian Accounts Receivable",
						"Name":               "Canadian Accounts Receivable",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Next page is the last page",
			Input: common.ReadParams{
				ObjectName: "customer",
				Fields:     connectors.Fields("domain", "displayName", "job"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.QueryParam("query", "SELECT * FROM Customer STARTPOSITION 1 MAXRESULTS 1000"),
				Then:  mockserver.Response(http.StatusOK, responseCustomer),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read with pagination returns next page",
			Input: common.ReadParams{
				ObjectName: "account",
				Fields:     connectors.Fields("accountsubtype", "accounttype", "active"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.QueryParam("query", "SELECT * FROM Account STARTPOSITION 1 MAXRESULTS 1000"),
				Then:  mockserver.Response(http.StatusOK, responseAccountWithNextPage),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				NextPage: "1001",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read with Since time filter",
			Input: common.ReadParams{
				ObjectName: "account",
				Fields:     connectors.Fields("name"),
				Since:      parseTime("2025-01-01T00:00:00Z"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseAccount),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name": "Canadian Accounts Receivable",
					},
					Raw: map[string]any{
						"Name": "Canadian Accounts Receivable",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read with Until time filter",
			Input: common.ReadParams{
				ObjectName: "account",
				Fields:     connectors.Fields("name"),
				Until:      parseTime("2025-12-31T23:59:59Z"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseAccount),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name": "Canadian Accounts Receivable",
					},
					Raw: map[string]any{
						"Name": "Canadian Accounts Receivable",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read with Since and Until time filters",
			Input: common.ReadParams{
				ObjectName: "account",
				Fields:     connectors.Fields("name"),
				Since:      parseTime("2025-01-01T00:00:00Z"),
				Until:      parseTime("2025-12-31T23:59:59Z"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseAccount),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name": "Canadian Accounts Receivable",
					},
					Raw: map[string]any{
						"Name": "Canadian Accounts Receivable",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read with NextPage parameter",
			Input: common.ReadParams{
				ObjectName: "account",
				Fields:     connectors.Fields("name"),
				NextPage:   "1001",
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseAccount),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name": "Canadian Accounts Receivable",
					},
					Raw: map[string]any{
						"Name": "Canadian Accounts Receivable",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read returns empty results",
			Input: common.ReadParams{
				ObjectName: "account",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.QueryParam("query", "SELECT * FROM Account STARTPOSITION 1 MAXRESULTS 1000"),
				Then:  mockserver.Response(http.StatusOK, responseAccountEmpty),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read with error response",
			Input: common.ReadParams{
				ObjectName: "account",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseError),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrCaller,
			},
		},
		{
			Name: "Successfully read item object",
			Input: common.ReadParams{
				ObjectName: "item",
				Fields:     connectors.Fields("name", "type", "active"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.QueryParam("query", "SELECT * FROM Item STARTPOSITION 1 MAXRESULTS 1000"),
				Then:  mockserver.Response(http.StatusOK, responseItem),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name":   "Services",
						"type":   "Service",
						"active": true,
					},
					Raw: map[string]any{
						"Name":   "Services",
						"Type":   "Service",
						"Active": true,
						"domain": "QBO",
						"sparse": false,
						"Level":  "0",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read customer with custom fields extracts custom field values to Fields",
			Input: common.ReadParams{
				ObjectName: "customer",
				Fields:     connectors.Fields("displayName", "projectCode", "budgetAmount", "startDate"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.QueryParam("query", "SELECT * FROM Customer STARTPOSITION 1 MAXRESULTS 1000"),
					mockcond.QueryParam("include", "enhancedAllCustomFields"),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomerWithCustomFields),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"displayname":  "Amy's Bird Sanctuary",
						"projectcode":  "PROJ-001",
						"budgetamount": float64(50000),
						"startdate":    "2024-01-15",
					},
					Raw: map[string]any{
						"Id":          "1",
						"DisplayName": "Amy's Bird Sanctuary",
					},
				}, {
					Fields: map[string]any{
						"displayname": "Bob's Burgers",
						"projectcode": "PROJ-002",
					},
					Raw: map[string]any{
						"Id":          "2",
						"DisplayName": "Bob's Burgers",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read account does not request custom fields",
			Input: common.ReadParams{
				ObjectName: "account",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.QueryParam("query", "SELECT * FROM Account STARTPOSITION 1 MAXRESULTS 1000"),
					// Note: no include=enhancedAllCustomFields for account (not in objectsWithCustomFields)
					mockcond.QueryParamsMissing("include"),
				},
				Then: mockserver.Response(http.StatusOK, responseAccount),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name": "Canadian Accounts Receivable",
					},
					Raw: map[string]any{
						"Name": "Canadian Accounts Receivable",
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
