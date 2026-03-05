package greenhouse

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

	responseError := testutils.DataFromFile(t, "error.json")
	responseCreate := testutils.DataFromFile(t, "write/candidates/create.json")
	responseUpdate := testutils.DataFromFile(t, "write/candidates/update.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Error response on bad request",
			Input: common.WriteParams{ObjectName: "candidates", RecordData: map[string]any{}},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnprocessableEntity, responseError),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("Your request included invalid JSON."),
			},
		},
		{
			Name:  "Create candidate via POST",
			Input: common.WriteParams{ObjectName: "candidates", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/candidates"),
				},
				Then: mockserver.Response(http.StatusOK, responseCreate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "12345",
				Data: map[string]any{
					"first_name": "John",
					"last_name":  "Doe",
					"company":    "Acme Corp",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update candidate via PATCH",
			Input: common.WriteParams{ObjectName: "candidates", RecordData: "dummy", RecordId: "12345"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v3/candidates/12345"),
				},
				Then: mockserver.Response(http.StatusOK, responseUpdate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "12345",
				Data: map[string]any{
					"first_name": "John",
					"last_name":  "Smith",
					"company":    "Acme Corp",
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
