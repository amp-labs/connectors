package fathom

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

	responseMeetingsFirst := testutils.DataFromFile(t, "meetings-first-page.json")
	responseMeetingsSecond := testutils.DataFromFile(t, "meetings-second-page.json")

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
			Input: common.ReadParams{ObjectName: "meetings", Fields: connectors.Fields("title", "url", "created_at")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/external/v1/meetings"),
				Then:  mockserver.Response(http.StatusOK, responseMeetingsFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"title":      "Nebo / ayan catchup",
						"url":        "https://fathom.video/calls/342490368",
						"created_at": "2025-07-03T18:31:53Z",
					},
					Raw: map[string]any{
						"title":                "Nebo / ayan catchup",
						"url":                  "https://fathom.video/calls/342490368",
						"created_at":           "2025-07-03T18:31:53Z",
						"scheduled_end_time":   "2025-07-03T18:30:00Z",
						"recording_start_time": "2025-07-03T18:01:30Z",
						"recording_end_time":   "2025-07-03T18:31:47Z",
					},
				}},
				NextPage: "eyJ0ZWFtX2NhbGxzIjp7InJlY29yZGluZ19zdGFydGVkX2F0IjoiMjAyNS0wNi0zMFQyMDowMjozMi4wNTQ4MzBaIiwiaWQiOjMzODY0NTU0MH0sImNvbXBsZXRlZF9zb3VyY2VzIjpbInRlYW1fcm9sZV9jYWxscyIsImhvc3Rfc2hhcmVkX3RlYW1fcm9sZV9jYWxscyIsImNvbnRhY3RfdGVhbV9tZW1iZXJfY2FsbHMiLCJmb2xkZXJfdGVhbV9jYWxscyIsImZvbGRlcl90ZWFtX3JvbGVfY2FsbHMiLCJmb2xkZXJfaG9zdF9zaGFyZWRfdGVhbV9yb2xlX2NhbGxzIiwiZm9sZGVyX2NvbnRhY3RfdGVhbV9tZW1iZXJfY2FsbHMiXX0=", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Next page is the last page",
			Input: common.ReadParams{
				ObjectName: "meetings",
				Fields:     connectors.Fields("title", "url", "created_at"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/external/v1/meetings"),
				Then:  mockserver.Response(http.StatusOK, responseMeetingsSecond),
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
