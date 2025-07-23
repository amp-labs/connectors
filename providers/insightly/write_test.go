package insightly

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

// nolint
func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	writeContactsResponse := testutils.DataFromFile(t, "write/contacts.json")
	writeLeadsResponse := testutils.DataFromFile(t, "write/leads.json")

	tests := []testroutines.Write{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "Contacts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Create contact",
			Input: common.WriteParams{ObjectName: "Contacts", RecordData: map[string]any{
				"EMAIL_ADDRESS": "pamela@mail.com",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3.1/Contacts"),
				},
				Then: mockserver.Response(http.StatusOK, writeContactsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "366638973",
				Data: map[string]any{
					"FIRST_NAME":    "Pamela",
					"LAST_NAME":     "Huber",
					"EMAIL_ADDRESS": "pamela@mail.com",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update Contact",
			Input: common.WriteParams{
				ObjectName: "Contacts",
				RecordId:   "366638973",
				RecordData: map[string]any{
					"EMAIL_ADDRESS": "pamela@mail.com",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v3.1/Contacts/366638973"),
					// Identifier is inserted.
					mockcond.Body(`{
						"EMAIL_ADDRESS": "pamela@mail.com",
						"CONTACT_ID": "366638973"
					}`),
				},
				Then: mockserver.Response(http.StatusOK, writeContactsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "366638973",
				Data: map[string]any{
					"FIRST_NAME":    "Pamela",
					"LAST_NAME":     "Huber",
					"EMAIL_ADDRESS": "pamela@mail.com",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update Leads",
			Input: common.WriteParams{
				ObjectName: "Leads",
				RecordId:   "78572651",
				RecordData: map[string]any{
					"FIRST_NAME": "Henrietta",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v3.1/Leads/78572651"),
					// Identifier is inserted.
					mockcond.Body(`{
  						"FIRST_NAME": "Henrietta",
						"LEAD_ID": "78572651"
					}`),
				},
				Then: mockserver.Response(http.StatusOK, writeLeadsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "78572651",
				Data: map[string]any{
					"SALUTATION": "Ms",
					"FIRST_NAME": "Henrietta",
					"LAST_NAME":  "Sloan",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
