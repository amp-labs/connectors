package getresponse

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	// Sample contact response data (matches actual GetResponse API structure)
	contactsResponse := `[
		{
			"contactId": "pV3r",
			"name": "John Doe",
			"origin": "import",
			"timeZone": "Europe/Warsaw",
			"activities": "https://api.getresponse.com/v3/contacts/pV3r/activities",
			"changedOn": "2024-01-16T11:00:00+0000",
			"createdOn": "2024-01-15T10:00:00+0000",
			"campaign": {
				"campaignId": "C",
				"href": "https://api.getresponse.com/v3/campaigns/C",
				"name": "Promo campaign"
			},
			"email": "john.doe@example.com",
			"dayOfCycle": "42",
			"scoring": 8,
			"engagementScore": 3,
			"href": "https://api.getresponse.com/v3/contacts/pV3r",
			"note": "Test note",
			"ipAddress": "1.2.3.4"
		},
		{
			"contactId": "pV4s",
			"name": "Jane Smith",
			"origin": "api",
			"timeZone": "America/New_York",
			"activities": "https://api.getresponse.com/v3/contacts/pV4s/activities",
			"changedOn": "2024-01-21T13:00:00+0000",
			"createdOn": "2024-01-20T12:00:00+0000",
			"campaign": {
				"campaignId": "C",
				"href": "https://api.getresponse.com/v3/campaigns/C",
				"name": "Promo campaign"
			},
			"email": "jane.smith@example.com",
			"dayOfCycle": "10",
			"scoring": 5,
			"engagementScore": 2,
			"href": "https://api.getresponse.com/v3/contacts/pV4s",
			"note": "Another note",
			"ipAddress": "5.6.7.8"
		}
	]`

	// Sample campaign response data (matches actual GetResponse API structure)
	campaignsResponse := `[
		{
			"campaignId": "f4PSi",
			"href": "https://api.getresponse.com/v3/campaigns/f4PSi",
			"name": "Test Campaign",
			"techName": "e5fe416f5a5cddc226e876e76257822b",
			"description": "Test Campaign Description",
			"languageCode": "EN",
			"isDefault": "false",
			"createdOn": "2024-01-10T08:00:00+0000"
		}
	]`

	// Sample custom-events response (connector-side filtering only)
	customEventsResponse := `[
		{
			"eventId": "evt1",
			"name": "Test Event",
			"createdOn": "2024-01-18T09:00:00Z"
		},
		{
			"eventId": "evt2",
			"name": "Test Event 2",
			"createdOn": "2024-01-25T10:00:00Z"
		}
	]`

	// Empty response for pagination test
	emptyResponse := `[]`

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Read contacts with provider-side filtering",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email", "name", "createdOn"),
				Since:      time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2024, time.January, 31, 23, 59, 59, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v3/contacts"),
					mockcond.QueryParam("query[createdOn][from]", "2024-01-01T00:00:00Z"),
					mockcond.QueryParam("query[createdOn][to]", "2024-01-31T23:59:59Z"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(contactsResponse)),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"email":     "john.doe@example.com",
							"name":      "John Doe",
							"createdon": "2024-01-15T10:00:00+0000",
						},
						Raw: map[string]any{
							"contactId":       "pV3r",
							"name":            "John Doe",
							"origin":          "import",
							"timeZone":        "Europe/Warsaw",
							"activities":      "https://api.getresponse.com/v3/contacts/pV3r/activities",
							"changedOn":       "2024-01-16T11:00:00+0000",
							"createdOn":       "2024-01-15T10:00:00+0000",
							"campaign":        map[string]any{"campaignId": "C", "href": "https://api.getresponse.com/v3/campaigns/C", "name": "Promo campaign"},
							"email":           "john.doe@example.com",
							"dayOfCycle":      "42",
							"scoring":         8.0,
							"engagementScore": 3.0,
							"href":            "https://api.getresponse.com/v3/contacts/pV3r",
							"note":            "Test note",
							"ipAddress":       "1.2.3.4",
						},
					},
					{
						Fields: map[string]any{
							"email":     "jane.smith@example.com",
							"name":      "Jane Smith",
							"createdon": "2024-01-20T12:00:00+0000",
						},
						Raw: map[string]any{
							"contactId":       "pV4s",
							"name":            "Jane Smith",
							"origin":          "api",
							"timeZone":        "America/New_York",
							"activities":      "https://api.getresponse.com/v3/contacts/pV4s/activities",
							"changedOn":       "2024-01-21T13:00:00+0000",
							"createdOn":       "2024-01-20T12:00:00+0000",
							"campaign":        map[string]any{"campaignId": "C", "href": "https://api.getresponse.com/v3/campaigns/C", "name": "Promo campaign"},
							"email":           "jane.smith@example.com",
							"dayOfCycle":      "10",
							"scoring":         5.0,
							"engagementScore": 2.0,
							"href":            "https://api.getresponse.com/v3/contacts/pV4s",
							"note":            "Another note",
							"ipAddress":       "5.6.7.8",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read campaigns without time filters",
			Input: common.ReadParams{
				ObjectName: "campaigns",
				Fields:     connectors.Fields("name", "createdOn"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/campaigns"),
				Then:  mockserver.Response(http.StatusOK, []byte(campaignsResponse)),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"name":      "Test Campaign",
							"createdon": "2024-01-10T08:00:00+0000",
						},
						Raw: map[string]any{
							"campaignId":   "f4PSi",
							"href":         "https://api.getresponse.com/v3/campaigns/f4PSi",
							"name":         "Test Campaign",
							"techName":     "e5fe416f5a5cddc226e876e76257822b",
							"description":  "Test Campaign Description",
							"languageCode": "EN",
							"isDefault":    "false",
							"createdOn":    "2024-01-10T08:00:00+0000",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read custom-events with connector-side filtering (no provider-side support)",
			Input: common.ReadParams{
				ObjectName: "custom-events",
				Fields:     connectors.Fields("name", "createdOn"),
				Since:      time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2024, time.January, 31, 23, 59, 59, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v3/custom-events"),
					// Should NOT have query[createdOn] parameters (connector-side filtering only)
					mockcond.QueryParamsMissing("query[createdOn][from]", "query[createdOn][to]"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(customEventsResponse)),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 2, // Both events should pass the filter (both within date range)
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"name":      "Test Event",
							"createdon": "2024-01-18T09:00:00Z",
						},
						Raw: map[string]any{
							"eventId":   "evt1",
							"name":      "Test Event",
							"createdOn": "2024-01-18T09:00:00Z",
						},
					},
					{
						Fields: map[string]any{
							"name":      "Test Event 2",
							"createdon": "2024-01-25T10:00:00Z",
						},
						Raw: map[string]any{
							"eventId":   "evt2",
							"name":      "Test Event 2",
							"createdOn": "2024-01-25T10:00:00Z",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read contacts with pagination (empty response indicates last page)",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/contacts"),
				Then:  mockserver.Response(http.StatusOK, []byte(emptyResponse)),
			}.Server(),
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
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
