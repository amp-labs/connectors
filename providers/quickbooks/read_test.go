package quickbooks

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

	responseAccount := testutils.DataFromFile(t, "account-read.json")
	responseCustomer := testutils.DataFromFile(t, "customer-read.json")

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
