package pylon

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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseAccountsFirst := testutils.DataFromFile(t, "accounts-first-page.json")
	responseAccountsSecond := testutils.DataFromFile(t, "accounts-second-page.json")
	responseIssuesEmpty := testutils.DataFromFile(t, "issues-empty-response.json")

	tests := []testconn.TestCaseRead{
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
			Input: common.ReadParams{ObjectName: "accounts", Fields: connectors.Fields("id", "name", "domain")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/accounts"),
				Then:  mockserver.Response(http.StatusOK, responseAccountsFirst),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":     "fb7c3918-5a16-45e9-ba4a-35fed505d07f",
						"name":   "Aurasell",
						"domain": "aurasell.ai",
					},
					Raw: map[string]any{
						"id":                            "fb7c3918-5a16-45e9-ba4a-35fed505d07f",
						"name":                          "Aurasell",
						"owner":                         nil,
						"domain":                        "aurasell.ai",
						"primary_domain":                "aurasell.ai",
						"type":                          "customer",
						"created_at":                    "2025-08-26T16:11:58Z",
						"tags":                          nil,
						"latest_customer_activity_time": "2025-08-28T18:08:38Z",
						"external_ids":                  nil,
					},
				}},
				NextPage: "MjAyNS0wOC0yNlQxNjoxMTo1OC44NzFafGZiN2MzOTE4LTVhMTYtNDVlOS1iYTRhLTM1ZmVkNTA1ZDA3Zg==", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Empty response with only request_id returns zero records without error",
			Input: common.ReadParams{ObjectName: "issues", Fields: connectors.Fields("id", "title")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/issues/search"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, responseIssuesEmpty),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     0,
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Issues are read via POST to the search endpoint",
			Input: common.ReadParams{ObjectName: "issues", Fields: connectors.Fields("id", "title")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/issues/search"),
					mockcond.MethodPOST(),
					// filter is optional, and is omitted when no window is requested.
					mockcond.Body(`{"limit":999}`),
				},
				Then: mockserver.Response(http.StatusOK, responseIssuesEmpty),
			}.Server(),
			Comparator:   testconn.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 0, NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Since alone filters on updated_at with a lower bound",
			Input: common.ReadParams{
				ObjectName: "issues",
				Fields:     connectors.Fields("id", "title"),
				Since:      time.Unix(1754518014, 0),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/issues/search"),
					mockcond.MethodPOST(),
					mockcond.Body(`{
						"limit":999,
						"filter":{
							"operator":"and",
							"subfilters":[
								{"field":"updated_at","operator":"time_is_after","value":"2025-08-06T22:06:54Z"}
							]
						}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseIssuesEmpty),
			}.Server(),
			Comparator:   testconn.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 0, NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Since and Until filter on updated_at with both bounds",
			Input: common.ReadParams{
				ObjectName: "issues",
				Fields:     connectors.Fields("id", "title"),
				Since:      time.Unix(1754518014, 0),
				Until:      time.Unix(1754518016, 0),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/issues/search"),
					mockcond.MethodPOST(),
					mockcond.Body(`{
						"limit":999,
						"filter":{
							"operator":"and",
							"subfilters":[
								{"field":"updated_at","operator":"time_is_after","value":"2025-08-06T22:06:54Z"},
								{"field":"updated_at","operator":"time_is_before","value":"2025-08-06T22:06:56Z"}
							]
						}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseIssuesEmpty),
			}.Server(),
			Comparator:   testconn.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 0, NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Next page is the last page",
			Input: common.ReadParams{
				ObjectName: "accounts",
				Fields:     connectors.Fields("id", "name", "domain"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/accounts"),
				Then:  mockserver.Response(http.StatusOK, responseAccountsSecond),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
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

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
