package supersend

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

func TestWrite(t *testing.T) {
	t.Parallel()

	createTeamResponse := testutils.DataFromFile(t, "write/create-team.json")
	createLabelResponse := testutils.DataFromFile(t, "write/create-label.json")
	updateLabelResponse := testutils.DataFromFile(t, "write/update-label.json")
	createSenderResponse := testutils.DataFromFile(t, "write/create-sender.json")
	updateSenderResponse := testutils.DataFromFile(t, "write/update-sender.json")
	createCampaignResponse := testutils.DataFromFile(t, "write/create-campaign.json")
	updateCampaignResponse := testutils.DataFromFile(t, "write/update-campaign.json")
	createContactResponse := testutils.DataFromFile(t, "write/create-contact.json")
	updateContactResponse := testutils.DataFromFile(t, "write/update-contact.json")
	createSenderProfileResponse := testutils.DataFromFile(t, "write/create-sender-profile.json")
	updateSenderProfileResponse := testutils.DataFromFile(t, "write/update-sender-profile.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "teams"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "Unsupported object returns error",
			Input:        common.WriteParams{ObjectName: "unsupported", RecordData: map[string]any{"foo": "bar"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Update teams fails (no update endpoint)",
			Input: common.WriteParams{
				ObjectName: "teams",
				RecordId:   "team-001",
				RecordData: map[string]any{"name": "Updated"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		// Teams - create only (uses /v2/teams)
		{
			Name: "Create team successfully",
			Input: common.WriteParams{
				ObjectName: "teams",
				RecordData: map[string]any{
					"name":   "New Team",
					"domain": "newteam.example.com",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/teams"),
				},
				Then: mockserver.Response(http.StatusOK, createTeamResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "team-123",
				Data: map[string]any{
					"id":   "team-123",
					"name": "New Team",
				},
			},
			ExpectedErrs: nil,
		},
		// Labels - create and update
		{
			Name: "Create label successfully",
			Input: common.WriteParams{
				ObjectName: "labels",
				RecordData: map[string]any{
					"name":  "Important",
					"color": "#FF5733",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/labels"),
				},
				Then: mockserver.Response(http.StatusOK, createLabelResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "label-456",
				Data: map[string]any{
					"id":    "label-456",
					"name":  "Important",
					"color": "#FF5733",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update label successfully",
			Input: common.WriteParams{
				ObjectName: "labels",
				RecordId:   "label-456",
				RecordData: map[string]any{
					"name":  "Very Important",
					"color": "#FF0000",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v1/labels/label-456"),
				},
				Then: mockserver.Response(http.StatusOK, updateLabelResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "label-456",
				Data: map[string]any{
					"id":    "label-456",
					"name":  "Very Important",
					"color": "#FF0000",
				},
			},
			ExpectedErrs: nil,
		},
		// Senders - create and update (no delete)
		{
			Name: "Create sender successfully",
			Input: common.WriteParams{
				ObjectName: "senders",
				RecordData: map[string]any{
					"email": "sender@example.com",
					"name":  "Test Sender",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/sender"),
				},
				Then: mockserver.Response(http.StatusOK, createSenderResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "sender-789",
				Data: map[string]any{
					"id":    "sender-789",
					"email": "sender@example.com",
					"name":  "Test Sender",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update sender successfully",
			Input: common.WriteParams{
				ObjectName: "senders",
				RecordId:   "sender-789",
				RecordData: map[string]any{
					"name": "Updated Sender",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v1/sender/sender-789"),
				},
				Then: mockserver.Response(http.StatusOK, updateSenderResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "sender-789",
				Data: map[string]any{
					"id":   "sender-789",
					"name": "Updated Sender",
				},
			},
			ExpectedErrs: nil,
		},
		// Campaigns - create and update
		{
			Name: "Create campaign successfully",
			Input: common.WriteParams{
				ObjectName: "campaigns",
				RecordData: map[string]any{
					"name":   "New Campaign",
					"teamId": "team-001",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/auto/campaign"),
				},
				Then: mockserver.Response(http.StatusOK, createCampaignResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "campaign-001",
				Data: map[string]any{
					"id":   "campaign-001",
					"name": "New Campaign",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update campaign successfully",
			Input: common.WriteParams{
				ObjectName: "campaigns",
				RecordId:   "campaign-001",
				RecordData: map[string]any{
					"name": "Updated Campaign",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v1/campaign/campaign-001"),
				},
				Then: mockserver.Response(http.StatusOK, updateCampaignResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "campaign-001",
				Data: map[string]any{
					"id":   "campaign-001",
					"name": "Updated Campaign",
				},
			},
			ExpectedErrs: nil,
		},
		// Contacts - create (POST) and update (PATCH)
		{
			Name: "Create contact successfully",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordData: map[string]any{
					"email":     "contact@example.com",
					"firstName": "John",
					"lastName":  "Doe",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/contacts"),
				},
				Then: mockserver.Response(http.StatusOK, createContactResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "contact-001",
				Data: map[string]any{
					"id":        "contact-001",
					"email":     "contact@example.com",
					"firstName": "John",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update contact uses PATCH",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "contact-001",
				RecordData: map[string]any{
					"firstName": "Jane",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v2/contacts/contact-001"),
				},
				Then: mockserver.Response(http.StatusOK, updateContactResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "contact-001",
				Data: map[string]any{
					"id":        "contact-001",
					"firstName": "Jane",
				},
			},
			ExpectedErrs: nil,
		},
		// Sender Profiles - create and update
		{
			Name: "Create sender profile successfully",
			Input: common.WriteParams{
				ObjectName: "sender-profiles",
				RecordData: map[string]any{
					"name":  "Marketing Profile",
					"email": "marketing@example.com",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/sender-profile"),
				},
				Then: mockserver.Response(http.StatusOK, createSenderProfileResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "sp-001",
				Data: map[string]any{
					"id":   "sp-001",
					"name": "Marketing Profile",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update sender profile successfully",
			Input: common.WriteParams{
				ObjectName: "sender-profiles",
				RecordId:   "sp-001",
				RecordData: map[string]any{
					"name": "Updated Marketing Profile",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v1/sender-profile/sp-001"),
				},
				Then: mockserver.Response(http.StatusOK, updateSenderProfileResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "sp-001",
				Data: map[string]any{
					"id":   "sp-001",
					"name": "Updated Marketing Profile",
				},
			},
			ExpectedErrs: nil,
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
