package devrev

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
	responseAccountsEmpty := testutils.DataFromFile(t, "read-accounts-empty.json")
	responseAccountsFirstPage := testutils.DataFromFile(t, "read-accounts-first-page.json")
	responseAccountsLastPage := testutils.DataFromFile(t, "read-accounts-last-page.json")
	responseArticlesEmpty := testutils.DataFromFile(t, "read-articles-empty.json")
	responseArticles := testutils.DataFromFile(t, "read-articles.json")

	tests := []testroutines.Read{
		{
			Name: "Read accounts empty",
			Input: common.ReadParams{
				ObjectName: "accounts",
				Fields:     connectors.Fields("id", "display_name", "modified_date"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/accounts.list"),
					mockcond.QueryParam("limit", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountsEmpty),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read accounts first page with pagination",
			Input: common.ReadParams{
				ObjectName: "accounts",
				Fields:     connectors.Fields("id", "modified_date"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/accounts.list"),
					mockcond.QueryParam("limit", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":            "don:identity:devrev:ACCT-1",
							"modified_date": "2026-02-20T17:51:38.642Z",
						},
						Raw: map[string]any{
							"id":            "don:identity:devrev:ACCT-1",
							"modified_date": "2026-02-20T17:51:38.642Z",
						},
					},
				},
				NextPage: testroutines.URLTestServer + "/accounts.list?limit=100&cursor=cursor_page_2",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read accounts second page using NextPage",
			Input: common.ReadParams{
				ObjectName: "accounts",
				Fields:     connectors.Fields("id", "display_name", "modified_date"),
				NextPage:   testroutines.URLTestServer + "/accounts.list?limit=100&cursor=cursor_page_2",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/accounts.list"),
					mockcond.QueryParam("cursor", "cursor_page_2"),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountsLastPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":            "don:identity:devrev:ACCT-2",
							"display_name":  "Beta Inc",
							"modified_date": "2026-02-20T18:00:00.000Z",
						},
						Raw: map[string]any{
							"id":            "don:identity:devrev:ACCT-2",
							"display_name":  "Beta Inc",
							"modified_date": "2026-02-20T18:00:00.000Z",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read accounts with PageSize uses limit query param",
			Input: common.ReadParams{
				ObjectName: "accounts",
				Fields:     connectors.Fields("id", "modified_date"),
				PageSize:   50,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/accounts.list"),
					mockcond.QueryParam("limit", "50"),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":            "don:identity:devrev:ACCT-1",
							"modified_date": "2026-02-20T17:51:38.642Z",
						},
						Raw: map[string]any{
							"id":            "don:identity:devrev:ACCT-1",
							"modified_date": "2026-02-20T17:51:38.642Z",
						},
					},
				},
				NextPage: testroutines.URLTestServer + "/accounts.list?limit=50&cursor=cursor_page_2",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read accounts with Since adds modified_date.after",
			Input: common.ReadParams{
				ObjectName: "accounts",
				Fields:     connectors.Fields("id", "modified_date"),
				Since:      time.Date(2026, 2, 20, 17, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/accounts.list"),
					mockcond.QueryParam("limit", "100"),
					mockcond.QueryParam("modified_date.after", "2026-02-20T17:00:00Z"),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":            "don:identity:devrev:ACCT-1",
							"modified_date": "2026-02-20T17:51:38.642Z",
						},
						Raw: map[string]any{
							"id":            "don:identity:devrev:ACCT-1",
							"modified_date": "2026-02-20T17:51:38.642Z",
						},
					},
				},
				NextPage: testroutines.URLTestServer + "/accounts.list?limit=100&cursor=cursor_page_2&modified_date.after=2026-02-20T17:00:00Z",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read articles empty",
			Input: common.ReadParams{
				ObjectName: "articles",
				Fields:     connectors.Fields("id", "description", "created_date"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/articles.list"),
					mockcond.QueryParam("limit", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseArticlesEmpty),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read articles",
			Input: common.ReadParams{
				ObjectName: "articles",
				Fields:     connectors.Fields("id", "title", "created_date", "modified_date"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/articles.list"),
					mockcond.QueryParam("limit", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseArticles),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":            "don:core:devrev:article/1",
							"title":         "api test",
							"created_date":  "2026-02-20T17:51:38.642Z",
							"modified_date": "2026-02-20T17:51:38.642Z",
						},
						Raw: map[string]any{
							"id":            "don:core:devrev:article/1",
							"title":         "api test",
							"created_date":  "2026-02-20T17:51:38.642Z",
							"modified_date": "2026-02-20T17:51:38.642Z",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read articles with Since filters connector-side (no server filter)",
			Input: common.ReadParams{
				ObjectName: "articles",
				Fields:     connectors.Fields("id", "modified_date"),
				Since:      time.Date(2026, 2, 20, 17, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/articles.list"),
					mockcond.QueryParam("limit", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseArticles),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":            "don:core:devrev:article/1",
							"modified_date": "2026-02-20T17:51:38.642Z",
						},
						Raw: map[string]any{
							"id":            "don:core:devrev:article/1",
							"modified_date": "2026-02-20T17:51:38.642Z",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read articles with Since after all records returns empty",
			Input: common.ReadParams{
				ObjectName: "articles",
				Fields:     connectors.Fields("id", "modified_date"),
				Since:      time.Date(2026, 2, 21, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/articles.list"),
					mockcond.QueryParam("limit", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseArticles),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
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
