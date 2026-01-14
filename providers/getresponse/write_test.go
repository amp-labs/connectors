package getresponse

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	// Sample contact response for update (200 OK with body)
	contactUpdateResponse := `{
		"contactId": "pV3r",
		"name": "John Doe Updated",
		"origin": "api",
		"timeZone": "Europe/Warsaw",
		"activities": "https://api.getresponse.com/v3/contacts/pV3r/activities",
		"changedOn": "2024-01-20T12:00:00+0000",
		"createdOn": "2024-01-15T10:00:00+0000",
		"campaign": {
			"campaignId": "C",
			"href": "https://api.getresponse.com/v3/campaigns/C",
			"name": "Promo campaign"
		},
		"email": "john.doe.updated@example.com",
		"dayOfCycle": "42",
		"scoring": 8,
		"engagementScore": 3,
		"href": "https://api.getresponse.com/v3/contacts/pV3r",
		"note": "Updated note",
		"ipAddress": "1.2.3.4"
	}`

	// Sample campaign response for update (200 OK with body)
	campaignUpdateResponse := `{
		"campaignId": "f4PSi",
		"href": "https://api.getresponse.com/v3/campaigns/f4PSi",
		"name": "Updated Campaign Name",
		"techName": "e5fe416f5a5cddc226e876e76257822b",
		"description": "Updated Campaign Description",
		"languageCode": "EN",
		"isDefault": "false",
		"createdOn": "2024-01-10T08:00:00+0000"
	}`

	// Error response for bad request
	errorResponse := `{
		"httpStatus": 400,
		"message": "Validation failed",
		"context": [
			{
				"field": "email",
				"message": "Email is required"
			}
		]
	}`

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Error invalid payload",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: map[string]any{}},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, []byte(errorResponse)),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrCaller,
				errors.New("Validation failed"),
			},
		},
		{
			Name:  "Create contact via POST (202 Accepted, no body)",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: map[string]any{"email": "new@example.com", "campaign": map[string]any{"campaignId": "C"}}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/contacts"),
				},
				Then: mockserver.Response(http.StatusAccepted, nil),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "",
				Errors:   nil,
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update contact via POST with RecordId (200 OK, with body)",
			Input: common.WriteParams{ObjectName: "contacts", RecordId: "pV3r", RecordData: map[string]any{"name": "John Doe Updated", "email": "john.doe.updated@example.com"}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/contacts/pV3r"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(contactUpdateResponse)),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "pV3r",
				Errors:   nil,
				Data: map[string]any{
					"contactId":       "pV3r",
					"name":            "John Doe Updated",
					"email":           "john.doe.updated@example.com",
					"origin":          "api",
					"timeZone":        "Europe/Warsaw",
					"activities":      "https://api.getresponse.com/v3/contacts/pV3r/activities",
					"changedOn":       "2024-01-20T12:00:00+0000",
					"createdOn":       "2024-01-15T10:00:00+0000",
					"campaign":        map[string]any{"campaignId": "C", "href": "https://api.getresponse.com/v3/campaigns/C", "name": "Promo campaign"},
					"dayOfCycle":      "42",
					"scoring":         8.0,
					"engagementScore": 3.0,
					"href":            "https://api.getresponse.com/v3/contacts/pV3r",
					"note":            "Updated note",
					"ipAddress":       "1.2.3.4",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create campaign via POST (202 Accepted, no body)",
			Input: common.WriteParams{ObjectName: "campaigns", RecordData: map[string]any{"name": "New Campaign", "languageCode": "EN"}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/campaigns"),
				},
				Then: mockserver.Response(http.StatusAccepted, nil),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "",
				Errors:   nil,
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update campaign via POST with RecordId (200 OK, with body)",
			Input: common.WriteParams{ObjectName: "campaigns", RecordId: "f4PSi", RecordData: map[string]any{"name": "Updated Campaign Name", "description": "Updated Campaign Description"}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/campaigns/f4PSi"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(campaignUpdateResponse)),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "f4PSi",
				Errors:   nil,
				Data: map[string]any{
					"campaignId":   "f4PSi",
					"href":         "https://api.getresponse.com/v3/campaigns/f4PSi",
					"name":         "Updated Campaign Name",
					"techName":     "e5fe416f5a5cddc226e876e76257822b",
					"description":  "Updated Campaign Description",
					"languageCode": "EN",
					"isDefault":    "false",
					"createdOn":    "2024-01-10T08:00:00+0000",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Error 401 Unauthorized",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: map[string]any{"email": "test@example.com"}},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnauthorized, []byte(`{"message": "Unauthorized"}`)),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrAccessToken,
			},
		},
		{
			Name:  "Error 500 Internal Server Error",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: map[string]any{"email": "test@example.com"}},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusInternalServerError, []byte(`{"message": "Internal Server Error"}`)),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrServer,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
