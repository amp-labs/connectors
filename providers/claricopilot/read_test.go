package claricopilot

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

	responseCallFirst := testutils.DataFromFile(t, "calls-first-page.json")
	responseCallLast := testutils.DataFromFile(t, "calls-second-page.json")

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
			Input: common.ReadParams{ObjectName: "calls", Fields: connectors.Fields("id", "title", "type")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/calls"),
				Then:  mockserver.Response(http.StatusOK, responseCallFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    "7549847e-4e97-4650-a20d-87791520e529",
						"title": "Ampersand promo",
						"type":  "RECORDING",
					},
					Raw: map[string]any{
						"id":                   "7549847e-4e97-4650-a20d-87791520e529",
						"source_id":            "REQUESTED_POST",
						"title":                "Ampersand promo",
						"status":               "POST_PROCESSING_DONE",
						"type":                 "RECORDING",
						"time":                 "2025-06-11T14:25:00.000Z",
						"last_modified_time":   "2025-06-11T14:53:06.310Z",
						"disposition":          "CALL_DID_NOT_CONNECT_WITH_PROSPECT",
						"call_review_page_url": "https://copilot.clari.com/call/7549847e-4e97-4650-a20d-87791520e529",
					},
				}},
				NextPage: testroutines.URLTestServer + "/calls?limit=100&skip=1", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Next page is the last page",
			Input: common.ReadParams{
				ObjectName: "calls",
				Fields:     connectors.Fields("id", "title", "type"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/calls"),
				Then:  mockserver.Response(http.StatusOK, responseCallLast),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     2,
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
