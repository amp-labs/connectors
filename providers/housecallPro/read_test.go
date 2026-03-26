package housecallpro

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

// nolint:funlen
func TestRead(t *testing.T) {
	t.Parallel()
	responseCustomersEmpty := testutils.DataFromFile(t, "read-customers-empty.json")
	responseCustomersFirst := testutils.DataFromFile(t, "read-customers-first.json")
	responseCustomersLast := testutils.DataFromFile(t, "read-customers-last.json")
	responseMaterialCategoryFirst := testutils.DataFromFile(t, "read-material-category-first.json")
	responseMaterialCategoryLast := testutils.DataFromFile(t, "read-material-category-last.json")
	responseInvoices := testutils.DataFromFile(t, "invoice.json")
	responseEmployees := testutils.DataFromFile(t, "read-employees.json")

	tests := []testroutines.Read{
		{
			Name: "Read customers empty",
			Input: common.ReadParams{
				ObjectName: "customers",
				Fields:     connectors.Fields("id", "first_name", "email"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/customers"),
					mockcond.QueryParam("per_page", defaultPageSize),
					mockcond.QueryParam("sort_by", "updated_at"),
					mockcond.QueryParam("sort_direction", "desc"),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomersEmpty),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read customers first page",
			Input: common.ReadParams{
				ObjectName: "customers",
				Fields:     connectors.Fields("id", "first_name", "email", "updated_at"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/customers"),
					mockcond.QueryParam("per_page", "1"),
					mockcond.QueryParam("sort_by", "updated_at"),
					mockcond.QueryParam("sort_direction", "desc"),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomersFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":         "cus_b8ae7027a36848889d21d9e8d7567edc",
							"first_name": "Updated Name",
							"email":      "john.doe@example.com",
							"updated_at": "2026-03-20T16:44:55Z",
						},
						Raw: map[string]any{
							"id":         "cus_b8ae7027a36848889d21d9e8d7567edc",
							"first_name": "Updated Name",
							"email":      "john.doe@example.com",
							"updated_at": "2026-03-20T16:44:55Z",
						},
					},
				},
				NextPage: testroutines.URLTestServer + "/customers?page=2&per_page=1&sort_by=updated_at&sort_direction=desc",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read customers second page using NextPage",
			Input: common.ReadParams{
				ObjectName: "customers",
				Fields:     connectors.Fields("id", "first_name", "email", "updated_at"),
				NextPage:   testroutines.URLTestServer + "/customers?page=2&per_page=1&sort_by=updated_at&sort_direction=desc",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/customers"),
					mockcond.QueryParam("page", "2"),
					mockcond.QueryParam("per_page", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomersLast),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":         "cus_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
							"first_name": "Alice",
							"email":      "alice.johnson@example.com",
							"updated_at": "2026-02-01T10:30:00Z",
						},
						Raw: map[string]any{
							"id":         "cus_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
							"first_name": "Alice",
							"email":      "alice.johnson@example.com",
							"updated_at": "2026-02-01T10:30:00Z",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},

		{
			Name: "Read customers with Since",
			Input: common.ReadParams{
				ObjectName: "customers",
				Fields:     connectors.Fields("id", "updated_at"),
				PageSize:   1,
				Since:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/customers"),
					mockcond.QueryParam("per_page", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomersFirst),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read employees",
			Input: common.ReadParams{
				ObjectName: "employees",
				Fields:     connectors.Fields("id", "first_name", "email", "created_at"),
				PageSize:   10,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/employees"),
					mockcond.QueryParam("per_page", "10"),
					mockcond.QueryParamsMissing("sort_by", "sort_direction"),
				},
				Then: mockserver.Response(http.StatusOK, responseEmployees),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":         "pro_aabbccdd11223344556677889900aabb",
							"first_name": "Jamie",
							"email":      "jamie.example@example.com",
							"created_at": "2026-01-15T12:00:00Z",
						},
						Raw: map[string]any{
							"id":         "pro_aabbccdd11223344556677889900aabb",
							"first_name": "Jamie",
							"email":      "jamie.example@example.com",
							"created_at": "2026-01-15T12:00:00Z",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read price book material categories first page",
			Input: common.ReadParams{
				ObjectName: "price_book/material_categories",
				Fields:     connectors.Fields("uuid", "name", "updated_at"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/price_book/material_categories"),
					mockcond.QueryParam("per_page", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseMaterialCategoryFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"uuid":       "pbmcat_db619b02f05d40d79470a38fc50332db",
							"name":       "New Category4",
							"updated_at": "2026-03-22T20:24:31Z",
						},
						Raw: map[string]any{
							"uuid":       "pbmcat_db619b02f05d40d79470a38fc50332db",
							"name":       "New Category4",
							"updated_at": "2026-03-22T20:24:31Z",
						},
					},
					{
						Fields: map[string]any{
							"uuid":       "pbmcat_43842dc5415b4ce4a911a4f0064fcfad",
							"name":       "New Category3",
							"updated_at": "2026-03-22T20:24:19Z",
						},
						Raw: map[string]any{
							"uuid":       "pbmcat_43842dc5415b4ce4a911a4f0064fcfad",
							"name":       "New Category3",
							"updated_at": "2026-03-22T20:24:19Z",
						},
					},
				},
				NextPage: testroutines.URLTestServer + "/api/price_book/material_categories?page=2&per_page=1&sort_by=updated_at&sort_direction=desc",
				Done:     false,
			},
			ExpectedErrs: nil,
		},

		{
			Name: "Read price book material categories second page",
			Input: common.ReadParams{
				ObjectName: "price_book/material_categories",
				Fields:     connectors.Fields("uuid", "name", "updated_at"),
				NextPage:   testroutines.URLTestServer + "/api/price_book/material_categories?page=2&per_page=1&sort_by=updated_at&sort_direction=desc",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/price_book/material_categories"),
					mockcond.QueryParam("page", "2"),
					mockcond.QueryParam("per_page", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseMaterialCategoryLast),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"uuid":       "pbmcat_2ea100604c474941bdc767de1e6aa9ec",
							"name":       "New Category2",
							"updated_at": "2026-03-22T20:24:16Z",
						},
						Raw: map[string]any{
							"uuid":       "pbmcat_2ea100604c474941bdc767de1e6aa9ec",
							"name":       "New Category2",
							"updated_at": "2026-03-22T20:24:16Z",
						},
					},
					{
						Fields: map[string]any{
							"uuid":       "pbmcat_56ac8b15d9f5493bad94f72d6965ad98",
							"name":       "New Category",
							"updated_at": "2026-03-22T20:24:11Z",
						},
						Raw: map[string]any{
							"uuid":       "pbmcat_56ac8b15d9f5493bad94f72d6965ad98",
							"name":       "New Category",
							"updated_at": "2026-03-22T20:24:11Z",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read invoices",
			Input: common.ReadParams{
				ObjectName: "invoices",
				Fields:     connectors.Fields("id", "invoice_number", "status", "sent_at"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/invoices"),
					mockcond.QueryParam("per_page", defaultPageSize),
					mockcond.QueryParamsMissing("sort_by", "sort_direction"),
				},
				Then: mockserver.Response(http.StatusOK, responseInvoices),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":             "invoice_a760cd9ddcb3443aba683163845ede30",
							"invoice_number": "2",
							"status":         "paid",
							"sent_at":        "2026-03-25T22:52:41Z",
						},
						Raw: map[string]any{
							"id":             "invoice_a760cd9ddcb3443aba683163845ede30",
							"invoice_number": "2",
							"status":         "paid",
							"sent_at":        "2026-03-25T22:52:41Z",
						},
					},
					{
						Fields: map[string]any{
							"id":             "invoice_718eafe396ba4913bd67289689d77f7d",
							"invoice_number": "1",
							"status":         "paid",
							"sent_at":        "2026-03-25T22:49:34Z",
						},
						Raw: map[string]any{
							"id":             "invoice_718eafe396ba4913bd67289689d77f7d",
							"invoice_number": "1",
							"status":         "paid",
							"sent_at":        "2026-03-25T22:49:34Z",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

// readAllPages follows NextPage until Read returns Done, appending each page’s Data.
// When Done is true, NextPage must be empty.
// If want is non-nil, also asserts len(rows) == len(want) and row.Fields[field] == want[i] for each row.
func readAllPages(
	t *testing.T, conn connectors.ReadConnector, params common.ReadParams, field string, want []string,
) []common.ReadResultRow {
	t.Helper()

	var all []common.ReadResultRow

	for {
		res, err := conn.Read(t.Context(), params)
		if err != nil {
			t.Fatalf("Read: %v", err)
		}

		all = append(all, res.Data...)

		if res.Done {
			if res.NextPage != "" {
				t.Fatalf("expected empty NextPage on last page, got %q", res.NextPage)
			}

			break
		}

		params.NextPage = res.NextPage
	}

	if want != nil {
		if len(all) != len(want) {
			t.Fatalf("expected %d rows across all pages, got %d", len(want), len(all))
		}

		for i, row := range all {
			if row.Fields[field] != want[i] {
				t.Fatalf("unexpected %s at index %d: got %v, want %s", field, i, row.Fields[field], want[i])
			}
		}
	}

	return all
}

func TestReadPriceBookMaterialCategoriesSequentialPagination(t *testing.T) {
	t.Parallel()

	responseFirst := testutils.DataFromFile(t, "read-material-category-first.json")
	responseLast := testutils.DataFromFile(t, "read-material-category-last.json")

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: mockserver.Cases{
			{
				If: mockcond.And{
					mockcond.Path("/api/price_book/material_categories"),
					mockcond.QueryParamsMissing("page"),
				},
				Then: mockserver.Response(http.StatusOK, responseFirst),
			},
			{
				If: mockcond.And{
					mockcond.Path("/api/price_book/material_categories"),
					mockcond.QueryParam("page", "2"),
				},
				Then: mockserver.Response(http.StatusOK, responseLast),
			},
		},
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("constructTestConnector: %v", err)
	}

	params := common.ReadParams{
		ObjectName: "price_book/material_categories",
		Fields:     connectors.Fields("uuid"),
		PageSize:   1,
		Since:      time.Date(2026, 3, 22, 20, 24, 15, 0, time.UTC),
	}

	readAllPages(t, conn, params, "uuid", []string{
		"pbmcat_db619b02f05d40d79470a38fc50332db",
		"pbmcat_43842dc5415b4ce4a911a4f0064fcfad",
		"pbmcat_2ea100604c474941bdc767de1e6aa9ec",
	})
}
