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
	groupsResponse := testutils.DataFromFile(t, "groups.json")
	licensesResponse := testutils.DataFromFile(t, "licenses.json")

	tests := []testroutines.TestCaseRead{
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
					mockcond.QueryParam("size", "200"),
					mockcond.QueryParam("page", "0"),
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
				NextPage: "1",
				Done:     false,
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
		{
			Name: "Groups (SCIM) read pulls records from resources envelope without pagination params",
			Input: common.ReadParams{
				ObjectName: "groups",
				Fields:     connectors.Fields("id", "displayName"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/identity/v1/Groups"),
				},
				Then: mockserver.Response(http.StatusOK, groupsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":          "string",
						"displayname": "string",
					},
					Raw: map[string]any{
						"id":          "string",
						"displayName": "string",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Licenses (Admin) read pulls records from results envelope with offset+pageSize pagination",
			Input: common.ReadParams{
				ObjectName: "licenses",
				Fields:     connectors.Fields("key", "type", "enabled"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/admin/rest/v1/accounts/" + testAccountKey + "/licenses"),
					mockcond.QueryParam("pageSize", "200"),
					mockcond.QueryParam("offset", "0"),
				},
				Then: mockserver.Response(http.StatusOK, licensesResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"key":     float64(0),
						"type":    "string",
						"enabled": true,
					},
					Raw: map[string]any{
						"key":     float64(0),
						"type":    "string",
						"enabled": true,
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

			tt.Run(t, func() (testroutines.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
