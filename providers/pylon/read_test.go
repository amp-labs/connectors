package pylon

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

	responseAccountsFirst := testutils.DataFromFile(t, "accounts-first-page.json")
	responseAccountsSecond := testutils.DataFromFile(t, "accounts-second-page.json")

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
			Input: common.ReadParams{ObjectName: "accounts", Fields: connectors.Fields("id", "name", "domain")},
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
