package orb

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,maintidx
	t.Parallel()

	responseNotFound := testutils.DataFromFile(t, "error-not-found.json")
	responseCustomersFirstPage := testutils.DataFromFile(t, "customers-first-page.json")
	responseInvoicesFirstPage := testutils.DataFromFile(t, "invoices-first-page.json")
	responseInvoicesSecondPage := testutils.DataFromFile(t, "invoices-second-page.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{Fields: connectors.Fields("id")},
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
			Name:         "Unknown object is not supported",
			Input:        common.ReadParams{ObjectName: "events", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Error response is parsed",
			Input: common.ReadParams{ObjectName: "customers", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrNotFound,
				testutils.StringError("A specified resource was not found.: " +
					"Could not find a resource of type Customer with identifier: connie-test-nonexistent."),
			},
		},
		{
			Name: "First page of customers is read and links to the next cursor",
			Input: common.ReadParams{
				ObjectName: "customers",
				Fields:     connectors.Fields("external_customer_id", "timezone"),
				PageSize:   2,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/customers"),
					mockcond.QueryParam("limit", "2"),
					mockcond.QueryParamsMissing("cursor", "created_at[gt]", "created_at[gte]"),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomersFirstPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id: "TuwejzVgofS2q9vf",
					Fields: map[string]any{
						"external_customer_id": "claude-billedaction-20261020193033-customer",
						"timezone":             "Etc/UTC",
					},
					Raw: map[string]any{
						"name": "claude-billedaction-20261020193033-customer",
					},
				}, {
					Id: "oT64AUpHxfBLLuNn",
					Fields: map[string]any{
						"external_customer_id": "claude-summary-20261020181156-customer",
						"timezone":             "Etc/UTC",
					},
					Raw: map[string]any{"id": "oT64AUpHxfBLLuNn"},
				}},
				NextPage: testconn.URLTestServer + "/v1/customers?cursor=eN3gvMFjPt5FjLbc&limit=2",
				Done:     false,
			},
		},
		{
			Name: "Invoices of every status are requested and the page size is capped at 100",
			Input: common.ReadParams{
				ObjectName: "invoices",
				Fields:     connectors.Fields("status", "total"),
				PageSize:   500,
				Since:      time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/invoices"),
					mockcond.QueryParam("limit", "100"),
					mockcond.QueryParam("status[]", "draft", "issued", "paid", "synced", "void"),
					mockcond.QueryParamsMissing("invoice_date[gt]", "invoice_date[gte]"),
				},
				Then: mockserver.Response(http.StatusOK, responseInvoicesFirstPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id:     "UFxvPGn6Smon5VGt",
					Raw:    map[string]any{"id": "UFxvPGn6Smon5VGt"},
					Fields: map[string]any{"status": "draft", "total": "8333.00"},
				}, {
					Id:     "mLHnMtYxYJxxHfMf",
					Raw:    map[string]any{"id": "mLHnMtYxYJxxHfMf"},
					Fields: map[string]any{"status": "draft", "total": "0.00"},
				}},
				NextPage: testconn.URLTestServer + "/v1/invoices?cursor=UDB67YkhpVDMatKL&limit=100" +
					"&status%5B%5D=draft&status%5B%5D=issued&status%5B%5D=paid&status%5B%5D=synced&status%5B%5D=void",
				Done: false,
			},
		},
		{
			Name: "Next page of invoices follows the cursor",
			Input: common.ReadParams{
				ObjectName: "invoices",
				Fields:     connectors.Fields("status"),
				NextPage: testconn.URLTestServer + "/v1/invoices?cursor=UDB67YkhpVDMatKL&limit=2" +
					"&status%5B%5D=draft&status%5B%5D=issued&status%5B%5D=paid&status%5B%5D=synced&status%5B%5D=void",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/invoices"),
					mockcond.QueryParam("cursor", "UDB67YkhpVDMatKL"),
					mockcond.QueryParam("limit", "2"),
				},
				Then: mockserver.Response(http.StatusOK, responseInvoicesSecondPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id:     "UDB67YkhpVDMatKL",
					Raw:    map[string]any{"id": "UDB67YkhpVDMatKL"},
					Fields: map[string]any{"status": "draft"},
				}, {
					Id:     "RznCTmNF4VqL5g6L",
					Raw:    map[string]any{"id": "RznCTmNF4VqL5g6L"},
					Fields: map[string]any{"status": "draft"},
				}},
				NextPage: testconn.URLTestServer + "/v1/invoices?cursor=hpwhpptfRw3vsUhL&limit=2" +
					"&status%5B%5D=draft&status%5B%5D=issued&status%5B%5D=paid&status%5B%5D=synced&status%5B%5D=void",
				Done: false,
			},
		},
		{
			Name:  "Last page has no next cursor",
			Input: common.ReadParams{ObjectName: "metrics", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK,
					`{"data": [], "pagination_metadata": {"has_more": false, "next_cursor": null}}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 0, NextPage: "", Done: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
