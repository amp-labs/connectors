package sellsy

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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseBadRequest := testutils.DataFromFile(t, "write/contacts/err-bad-request.json")
	responseContacts := testutils.DataFromFile(t, "write/contacts/new.json")

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
				Always: mockserver.Response(http.StatusBadRequest, responseBadRequest),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Le contenu de la requête est invalid: le champ 'last_name' est manquant."),
			},
		},
		{
			Name:  "Create contact via POST",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/contacts"),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "39",
				Errors:   nil,
				Data: map[string]any{
					"first_name": "Waldo",
					"last_name":  "Vazquez",
					"created":    "2025-09-16T23:47:09+02:00",
					"updated":    "2025-09-16T23:47:09+02:00",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update contact via PUT",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: "dummy", RecordId: "39"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v2/contacts/39"),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "39",
				Errors:   nil,
				Data: map[string]any{
					"first_name": "Waldo",
					"last_name":  "Vazquez",
					"created":    "2025-09-16T23:47:09+02:00",
					"updated":    "2025-09-16T23:47:09+02:00",
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
