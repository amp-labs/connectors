package microsoft

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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	errorUnknownResource := testutils.DataFromFile(t, "read/unknown-resource.json")
	responseUsersFirst := testutils.DataFromFile(t, "read/users/1-first-page.json")
	responseUsersLast := testutils.DataFromFile(t, "read/users/2-second-page.json")
	responseCalendarEvents := testutils.DataFromFile(t, "read/events/list.json")
	responseMessagesEvents := testutils.DataFromFile(t, "read/messages/list.json")

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
			Name:  "Correct error message is understood from JSON response",
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorUnknownResource),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest, testutils.StringError("Resource not found for the segment 'user'."),
			},
		},
		{
			Name:  "Successfully read users",
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields("displayName")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1.0/users"),
				Then:  mockserver.Response(http.StatusOK, responseUsersFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"displayname": "Integration User",
					},
					Raw: map[string]any{
						"surname":           "User",
						"userPrincipalName": "integration.user_withampersand.com#EXT#@integrationuserwithampersan.onmicrosoft.com",
						"id":                "12151ea6-6d86-4afd-a68d-88ab34f5170a",
					},
				}},
				NextPage: "https://graph.microsoft.com/v1.0/users?$top=1&$skiptoken=RFNwdAIAAQAAACM6aW50ZWdyYXRpb24udXNlckB3aXRoYW1wZXJzYW5kLmNvbSlVc2VyXzEyMTUxZWE2LTZkODYtNGFmZC1hNjhkLTg4YWIzNGY1MTcwYbkAAAAAAAAAAAAA", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Next page is the last page",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("displayName"),
				NextPage:   testroutines.URLTestServer + "/v1.0/users?$top=1&$skiptoken=RFNwdAIAAQAAACM6aW50ZWdyYXRpb24udXNlckB3aXRoYW1wZXJzYW5kLmNvbSlVc2VyXzEyMTUxZWE2LTZkODYtNGFmZC1hNjhkLTg4YWIzNGY1MTcwYbkAAAAAAAAAAAAA", // nolint:lll
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1.0/users"),
				Then:  mockserver.Response(http.StatusOK, responseUsersLast),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read events",
			Input: common.ReadParams{
				ObjectName: "me/events",
				Fields:     connectors.Fields("subject", "bodyPreview"),
				Since: time.Date(2024, 9, 19, 4, 30, 45, 600,
					time.FixedZone("UTC-8", -8*60*60)),
				PageSize: 28,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1.0/me/events"),
					// Pacific time to UTC is achieved by adding 8 hours
					mockcond.QueryParam("$filter", "lastModifiedDateTime ge 2024-09-19T12:30:45.000Z"),
					mockcond.QueryParam("$top", "28"),
				},
				Then: mockserver.Response(http.StatusOK, responseCalendarEvents),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"subject":     "Movies night",
						"bodypreview": "Gather to watch cinema.",
					},
					Raw: map[string]any{
						"id":                         "AAMkAGY0YzAwY2ViLWQyODktNDI3NS1iNmY4LTE5YzU0MjI5ZTA4OQBGAAAAAABeMJSlO8qLToz2i2IQ1wsqBwB8hj1Rtd60SKTngNs3if9RAAB-ie_oAAB8hj1Rtd60SKTngNs3if9RAAEMsx4rAAA=",
						"reminderMinutesBeforeStart": float64(15),
						"isReminderOn":               true,
						"hasAttachments":             false,
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read messages",
			Input: common.ReadParams{
				ObjectName: "me/messages",
				Fields:     connectors.Fields("subject", "bodyPreview", "importance"),
				Since: time.Date(2024, 9, 19, 4, 30, 45, 600,
					time.FixedZone("UTC-8", -8*60*60)),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1.0/me/messages"),
					// Pacific time to UTC is achieved by adding 8 hours
					mockcond.QueryParam("$filter", "lastModifiedDateTime ge 2024-09-19T12:30:45.000Z"),
					mockcond.QueryParam("$top", "100"), // default pagination
				},
				Then: mockserver.Response(http.StatusOK, responseMessagesEvents),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"subject":     "Timmy Wehner",
						"bodypreview": "Forrest Von",
						"importance":  "normal",
					},
					Raw: map[string]any{
						"id": "AAMkAGY0YzAwY2ViLWQyODktNDI3NS1iNmY4LTE5YzU0MjI5ZTA4OQBGAAAAAABeMJSlO8qLToz2i2IQ1wsqBwB8hj1Rtd60SKTngNs3if9RAAAAAAEKAAB8hj1Rtd60SKTngNs3if9RAAEMs4IlAAA=",
					},
					Id: "AAMkAGY0YzAwY2ViLWQyODktNDI3NS1iNmY4LTE5YzU0MjI5ZTA4OQBGAAAAAABeMJSlO8qLToz2i2IQ1wsqBwB8hj1Rtd60SKTngNs3if9RAAAAAAEKAAB8hj1Rtd60SKTngNs3if9RAAEMs4IlAAA=",
				}, {
					Fields: map[string]any{
						"subject":     "Gail Waelchi",
						"bodypreview": "Eleonore Kutch",
						"importance":  "normal",
					},
					Raw: map[string]any{
						"id": "AAMkAGY0YzAwY2ViLWQyODktNDI3NS1iNmY4LTE5YzU0MjI5ZTA4OQBGAAAAAABeMJSlO8qLToz2i2IQ1wsqBwB8hj1Rtd60SKTngNs3if9RAAAAAAEKAAB8hj1Rtd60SKTngNs3if9RAAEMs4ImAAA=",
					},
					Id: "AAMkAGY0YzAwY2ViLWQyODktNDI3NS1iNmY4LTE5YzU0MjI5ZTA4OQBGAAAAAABeMJSlO8qLToz2i2IQ1wsqBwB8hj1Rtd60SKTngNs3if9RAAAAAAEKAAB8hj1Rtd60SKTngNs3if9RAAEMs4ImAAA=",
				}},
				NextPage: "https://graph.microsoft.com/v1.0/me/messages?%24top=10&%24skip=10",
				Done:     false,
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
