package dynamicsbusiness

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

	responseErrorFormat := testutils.DataFromFile(t, "write/contacts/not-found.json")
	responseContacts := testutils.DataFromFile(t, "write/contacts/new.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "Contacts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:  "Correct error message is understood from JSON response",
			Input: common.WriteParams{ObjectName: "Contacts", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseErrorFormat),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Resource not found for the segment 'contact'."),
			},
		},
		{
			Name:  "Create Contact",
			Input: common.WriteParams{ObjectName: "Contacts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path(
						"/v2.0/test-workspace/test-environment/api/v2.0/companies(test-company-id)/Contacts"),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "9dc64b0b-4630-f011-9a4a-7ced8d1df4b8",
				Errors:   nil,
				Data: map[string]any{
					"type":        "Person",
					"displayName": "Monique Peters",
					"number":      "CT000025",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update Contacts",
			Input: common.WriteParams{
				ObjectName: "Contacts",
				RecordId:   "9dc64b0b-4630-f011-9a4a-7ced8d1df4b8",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v2.0/test-workspace/test-environment/api/v2.0/companies(test-company-id)" +
						"/Contacts(9dc64b0b-4630-f011-9a4a-7ced8d1df4b8)"),
					mockcond.Header(http.Header{"If-Match": []string{"*"}}),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "9dc64b0b-4630-f011-9a4a-7ced8d1df4b8",
				Errors:   nil,
				Data: map[string]any{
					"type":        "Person",
					"displayName": "Monique Peters",
					"number":      "CT000025",
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
