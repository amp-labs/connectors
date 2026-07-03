package flatfile

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseEventsFirst := testutils.DataFromFile(t, "events-first-page.json")
	responseEventsSecond := testutils.DataFromFile(t, "events-second-page.json")

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
			Name:  "Successful read first page with chosen fields",
			Input: common.ReadParams{ObjectName: "events", Fields: connectors.Fields("id", "domain", "topic")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/events"),
				Then:  mockserver.Response(http.StatusOK, responseEventsFirst),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":     "us_evt_qgpXn6x9veAtkSZ2",
						"domain": "job",
						"topic":  "job:created",
					},
					Raw: map[string]any{
						"id":        "us_evt_qgpXn6x9veAtkSZ2",
						"domain":    "job",
						"topic":     "job:created",
						"createdAt": "2025-07-03T11:52:35.564Z",
						"dataUrl":   "",
					},
				}},
				NextPage: testconn.URLTestServer + "/v1/events?pageNumber=2&pageSize=100", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},

		{
			Name:  "Successful read second page with chosen fields",
			Input: common.ReadParams{ObjectName: "events", Fields: connectors.Fields("id", "domain", "topic"), NextPage: testconn.URLTestServer + "/v1/events?pageNumber=2&pageSize=100"}, // nolint:lll
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/events"),
				Then:  mockserver.Response(http.StatusOK, responseEventsSecond),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":     "us_evt_qhsbyjuv89lf5n9e",
						"domain": "job",
						"topic":  "job:created",
					},
					Raw: map[string]any{
						"id":        "us_evt_qhsbyjuv89lf5n9e",
						"domain":    "job",
						"topic":     "job:created",
						"createdAt": "2025-07-03T11:52:08.818Z",
					},
				}},
				NextPage: testconn.URLTestServer + "/v1/events?pageNumber=3&pageSize=100", // nolint:lll
				Done:     false,
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
