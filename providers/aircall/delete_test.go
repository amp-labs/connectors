package aircall

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestDelete(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []testroutines.Delete{
		// --- Contacts ---
		{
			Name: "Delete contact successfully",
			Input: common.DeleteParams{
				ObjectName: "contacts",
				RecordId:   "12345",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodDelete),
					mockcond.Path("/v1/contacts/12345"),
				},
				Then: mockserver.Response(http.StatusNoContent, nil),
			}.Server(),
			Expected: &common.DeleteResult{
				Success: true,
			},
		},
		// --- Users ---
		{
			Name: "Delete user successfully",
			Input: common.DeleteParams{
				ObjectName: "users",
				RecordId:   "555",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodDelete),
					mockcond.Path("/v1/users/555"),
				},
				Then: mockserver.Response(http.StatusNoContent, nil),
			}.Server(),
			Expected: &common.DeleteResult{
				Success: true,
			},
		},
		// --- Tags ---
		{
			Name: "Delete tag successfully",
			Input: common.DeleteParams{
				ObjectName: "tags",
				RecordId:   "10",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodDelete),
					mockcond.Path("/v1/tags/10"),
				},
				Then: mockserver.Response(http.StatusNoContent, nil),
			}.Server(),
			Expected: &common.DeleteResult{
				Success: true,
			},
		},
		// --- Teams ---
		{
			Name: "Delete team successfully",
			Input: common.DeleteParams{
				ObjectName: "teams",
				RecordId:   "88",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodDelete),
					mockcond.Path("/v1/teams/88"),
				},
				Then: mockserver.Response(http.StatusNoContent, nil),
			}.Server(),
			Expected: &common.DeleteResult{
				Success: true,
			},
		},
		// --- Error Cases ---
		{
			Name: "Delete not found error",
			Input: common.DeleteParams{
				ObjectName: "contacts",
				RecordId:   "99999",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodDelete),
					mockcond.Path("/v1/contacts/99999"),
				},
				Then: mockserver.Response(http.StatusNotFound, []byte(`{"error": "Not Found"}`)),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrNotFound,
			},
		},
		{
			Name: "Delete unsupported object",
			Input: common.DeleteParams{
				ObjectName: "numbers",
				RecordId:   "123",
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Delete unsupported object - calls",
			Input: common.DeleteParams{
				ObjectName: "calls",
				RecordId:   "999",
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
