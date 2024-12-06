package constantcontact

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

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseCreateContact := testutils.DataFromFile(t, "write-contact.json")
	responseCreateEmailCampaign := testutils.DataFromFile(t, "write-email-campaign.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "contacts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "Unsupported object name",
			Input:        common.WriteParams{ObjectName: "butterflies", RecordData: "dummy"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Creation of a contacts",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.PathSuffix("/contacts"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, responseCreateContact),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "af73e650-96f0-11ef-b2a0-fa163eafb85e",
				Errors:   nil,
				Data: map[string]any{
					"first_name": "Debora",
					"last_name":  "Lang",
					"job_title":  "Musician",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update of a contact via PUT",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "af73e650-96f0-11ef-b2a0-fa163eafb85e",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.PathSuffix("/contacts/af73e650-96f0-11ef-b2a0-fa163eafb85e"),
					mockcond.MethodPUT(),
				},
				Then: mockserver.Response(http.StatusOK, responseCreateContact),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "af73e650-96f0-11ef-b2a0-fa163eafb85e",
				Errors:   nil,
				Data: map[string]any{
					"first_name": "Debora",
					"last_name":  "Lang",
					"job_title":  "Musician",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Creation of email campaign",
			Input: common.WriteParams{
				ObjectName: "email_campaigns",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.PathSuffix("/emails"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, responseCreateEmailCampaign),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "987cb9ab-ad0c-4087-bd46-2e2ad241221f",
				Errors:   nil,
				Data: map[string]any{
					"current_status": "Draft",
					"name":           "December Newsletter for Dog Lovers",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update of email campaign is the only to use PATCH",
			Input: common.WriteParams{
				ObjectName: "email_campaigns",
				RecordId:   "987cb9ab-ad0c-4087-bd46-2e2ad241221f",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.PathSuffix("/emails/987cb9ab-ad0c-4087-bd46-2e2ad241221f"),
					mockcond.MethodPATCH(),
				},
				Then: mockserver.Response(http.StatusOK, responseCreateEmailCampaign),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "987cb9ab-ad0c-4087-bd46-2e2ad241221f",
				Errors:   nil,
				Data: map[string]any{
					"current_status": "Draft",
					"name":           "December Newsletter for Dog Lovers",
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
