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
	responseCreate := testutils.DataFromFile(t, "write/applications/create.json")
	responseUpdate := testutils.DataFromFile(t, "write/applications/update.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Error response on bad request",
			Input: common.WriteParams{ObjectName: "applications", RecordData: map[string]any{}},
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
			Name:  "Create application via POST",
			Input: common.WriteParams{ObjectName: "applications", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/applications"),
				},
				Then: mockserver.Response(http.StatusOK, responseCreate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "12345",
				Data: map[string]any{
					"candidate_id": float64(101),
					"status":       "in_process",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update application via PATCH",
			Input: common.WriteParams{ObjectName: "applications", RecordData: "dummy", RecordId: "12345"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v3/applications/12345"),
				},
				Then: mockserver.Response(http.StatusOK, responseUpdate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "12345",
				Data: map[string]any{
					"candidate_id": float64(101),
					"status":       "hired",
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
