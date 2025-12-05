package keap

import (
	"errors"
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

	errorContactBadRequest := testutils.DataFromFile(t, "write-contacts-missing-fields.json")
	responseContacts := testutils.DataFromFile(t, "write-contacts.json")

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
			Name:  "Error on invalid create payload",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorContactBadRequest),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New(
					"Contact must contain an e-mail address or phone number",
				),
			},
		},
		{
			Name: "Companies are updated using PATCH",
			Input: common.WriteParams{
				ObjectName: "companies",
				RecordId:   "123",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPATCH(),
				Then:  mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Automation categories are updated using PUT",
			Input: common.WriteParams{
				ObjectName: "automationCategory",
				RecordId:   "123",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Valid creation of a contacts", // update is the same
			Input: common.WriteParams{ObjectName: "contacts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/crm/rest/v2/contacts"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "30",
				Errors:   nil,
				Data: map[string]any{
					"email_status": "NonMarketable",
					"date_created": "2024-11-12T21:45:54.000+0000",
					"last_updated": "2024-11-12T21:46:01.341+0000",
					"email_addresses": []any{map[string]any{
						"email": "john.doe@gmail.com",
						"field": "EMAIL1",
					}},
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
