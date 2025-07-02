package teamleader

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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	createContactsResponse := testutils.DataFromFile(t, "create-contacts.json")
	createCompaniesResponse := testutils.DataFromFile(t, "create-companies.json")

	tests := []testroutines.Write{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "leads"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},

		{
			Name: "Successfully creation of an contact",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: map[string]any{
				"first_name": "Johntest",
				"last_name":  "Run",
				"email":      "johntest.run@example.com",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, createContactsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "7c1d8672-f502-4333-9ea4-7a45add15115",
				Data: map[string]any{
					"id":   "7c1d8672-f502-4333-9ea4-7a45add15115",
					"type": "contact",
				},
			},
			ExpectedErrs: nil,
		},

		{
			Name: "Successfully creation of a companies",
			Input: common.WriteParams{ObjectName: "companies", RecordData: map[string]any{
				"name": "pied piper",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, createCompaniesResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "4784189d-610b-4488-b3a5-5f324f752417",
				Data: map[string]any{
					"id":   "4784189d-610b-4488-b3a5-5f324f752417",
					"type": "company",
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
