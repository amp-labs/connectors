package dynamicsbusiness

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

	responseErrorFormatJSON := testutils.DataFromFile(t, "read/balance-sheets/error-since.json")
	responseErrorFormatXML := testutils.DataFromFile(t, "err-missing-environment.xml")
	responseBalanceSheets := testutils.DataFromFile(t, "read/balance-sheets/1-last-page.json")
	responseCustomersFirstPage := testutils.DataFromFile(t, "read/customers/1-first-page.json")
	responseCustomersLastPage := testutils.DataFromFile(t, "read/customers/2-last-page.json")

	millisecondInNano := int(time.Millisecond.Nanoseconds())
	date := time.Date(2024, 9, 19, 4, 30, 45, 621*millisecondInNano,
		time.FixedZone("UTC-8", -8*60*60))

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "Groups"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Correct error message is understood from JSON response",
			Input: common.ReadParams{ObjectName: "BalanceSheets", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseErrorFormatJSON),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("Could not find a property named 'lastModifiedDateTime'"),
			},
		},
		{
			Name:  "Correct error message is understood from XML response",
			Input: common.ReadParams{ObjectName: "BalanceSheets", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentXML(),
				Always: mockserver.Response(http.StatusNotFound, responseErrorFormatXML),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				common.ErrNotFound,
				testutils.StringError("Environment does not exist."),
			},
		},
		{
			Name: "Incremental read of customers first page with next page",
			Input: common.ReadParams{
				ObjectName: "Customers",
				Fields:     connectors.Fields("id", "displayName", "email"),
				Since:      date,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					// Probing request succeeds to indicate incremental reading is supported by the object.
					If: mockcond.And{
						mockcond.Path(
							"/v2.0/test-workspace/test-environment/api/v2.0/companies(test-company-id)/Customers"),
						mockcond.QueryParam("$filter", "lastModifiedDateTime ge 2024-09-19T12:30:45.621Z"),
						mockcond.Header(http.Header{"Prefer": []string{"odata.maxpagesize=1"}}),
					},
					Then: mockserver.Response(http.StatusOK, responseCustomersFirstPage),
				}, {
					If: mockcond.And{
						mockcond.Path(
							"/v2.0/test-workspace/test-environment/api/v2.0/companies(test-company-id)/Customers"),
						mockcond.QueryParam("$filter", "lastModifiedDateTime ge 2024-09-19T12:30:45.621Z"),
						mockcond.Header(http.Header{"Prefer": []string{"odata.maxpagesize=100"}}),
					},
					Then: mockserver.Response(http.StatusOK, responseCustomersFirstPage),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":          "55e545bc-f6f9-ef11-9344-6045bdc8c234",
						"displayname": "Adatum Corporation",
						"email":       "robert.townes@contoso.com",
					},
					Raw: map[string]any{
						"taxAreaDisplayName": "ATLANTA, GA",
						"postalCode":         "31772",
					},
				}, {
					Fields: map[string]any{
						"id":          "5de545bc-f6f9-ef11-9344-6045bdc8c234",
						"displayname": "Trey Research",
						"email":       "helen.ray@contoso.com",
					},
					Raw: map[string]any{
						"taxAreaDisplayName": "CHICAGO, IL",
						"postalCode":         "61236",
					},
				}},
				NextPage: "https://api.businesscentral.dynamics.com/v2.0/5c6241d0-74cc-48a2-b667-3eb0d738af72/Production/api/v2.0/companies(70c0c603-f4f9-ef11-9344-6045bdc8c234)/customers?aid=FIN&$skiptoken=5de545bc-f6f9-ef11-9344-6045bdc8c234", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read customers last page",
			Input: common.ReadParams{
				ObjectName: "Customers",
				Fields:     connectors.Fields("id", "displayName", "email"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.Path(
					"/v2.0/test-workspace/test-environment/api/v2.0/companies(test-company-id)/Customers"),
				Then: mockserver.Response(http.StatusOK, responseCustomersLastPage),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     0,
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			// Balance Sheets is one of the objects that will error out
			// if connector would send filter on undefined field.
			// We expect a probing request to take place first which will return 400 Bad Request.
			Name: "Balance Sheets incremental reading doesn't send undefined time query params",
			Input: common.ReadParams{
				ObjectName: "BalanceSheets",
				Fields:     connectors.Fields("display"),
				Since:      date,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					// Probing request fails, BalanceSheets doesn't have `lastModifiedDateTime` property.
					If: mockcond.And{
						mockcond.Path(
							"/v2.0/test-workspace/test-environment/api/v2.0/companies(test-company-id)/BalanceSheets"),
						mockcond.QueryParam("$filter", "lastModifiedDateTime ge 2024-09-19T12:30:45.621Z"),
						mockcond.Header(http.Header{"Prefer": []string{"odata.maxpagesize=1"}}),
					},
					Then: mockserver.Response(http.StatusBadRequest, responseErrorFormatJSON),
				}, {
					If: mockcond.And{
						mockcond.Path(
							"/v2.0/test-workspace/test-environment/api/v2.0/companies(test-company-id)/BalanceSheets"),
						mockcond.QueryParamsMissing("$filter"),
						mockcond.Header(http.Header{"Prefer": []string{"odata.maxpagesize=100"}}),
					},
					Then: mockserver.Response(http.StatusOK, responseBalanceSheets),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"display": "Current Assets",
					},
					Raw: map[string]any{
						"id":       "2a294eb6-f6f9-ef11-9344-6045bdc8c234",
						"lineType": "header",
					},
				}, {
					Fields: map[string]any{
						"display": "Cash",
					},
					Raw: map[string]any{
						"id":       "2b294eb6-f6f9-ef11-9344-6045bdc8c234",
						"lineType": "detail",
					},
				}, {
					Fields: map[string]any{
						"display": "Equipment",
					},
					Raw: map[string]any{
						"id":       "31294eb6-f6f9-ef11-9344-6045bdc8c234",
						"lineType": "detail",
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
