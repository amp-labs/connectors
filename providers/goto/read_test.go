package gotoconn

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

func TestRead(t *testing.T) { //nolint:funlen
	t.Parallel()

	webinarsResponse := testutils.DataFromFile(t, "webinars.json")
	sessionsResponse := testutils.DataFromFile(t, "sessions.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "webinars"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "Webinars read pulls records from _embedded envelope",
			Input: common.ReadParams{
				ObjectName: "webinars",
				Fields:     connectors.Fields("webinarKey", "subject", "numberOfRegistrants"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/G2W/rest/v2/organizers/" + testAccountKey + "/webinars"),
					mockcond.QueryParam("size", "100"),
				},
				Then: mockserver.Response(http.StatusOK, webinarsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"webinarkey":          "7878787878787878",
						"subject":             "Introduction to GoToWebinar",
						"numberofregistrants": float64(42),
					},
					Raw: map[string]any{
						"webinarKey": "7878787878787878",
						"subject":    "Introduction to GoToWebinar",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Sessions read pulls records from top-level objectName envelope",
			Input: common.ReadParams{
				ObjectName: "sessions",
				Fields:     connectors.Fields("sessionId", "status", "expertName"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/G2A/rest/v1/extendedsessions"),
					mockcond.QueryParam("size", "100"),
				},
				Then: mockserver.Response(http.StatusOK, sessionsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"sessionid":  "SS-87555",
						"status":     "complete",
						"expertname": "John Doe",
					},
					Raw: map[string]any{
						"sessionId":  "SS-87555",
						"status":     "complete",
						"expertName": "John Doe",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		//nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
