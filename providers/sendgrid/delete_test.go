package sendgrid

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	tests := []testconn.TestCaseDelete{
		{
			Name:         "Object name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Record id is required",
			Input:        common.DeleteParams{ObjectName: objectLists},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name: "Unsupported object",
			Input: common.DeleteParams{
				ObjectName: objectBounces,
				RecordId:   "bounce@example.com",
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Delete list",
			Input: common.DeleteParams{
				ObjectName: objectLists,
				RecordId:   "ca7a3796-e8a8-4029-9ccb-df8937940562",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/v3/marketing/lists/ca7a3796-e8a8-4029-9ccb-df8937940562"),
				},
				Then: mockserver.Response(http.StatusNoContent, nil),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name: "Delete template",
			Input: common.DeleteParams{
				ObjectName: objectTemplates,
				RecordId:   "733ba07f-ead1-41fc-933a-3976baa23716",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/v3/templates/733ba07f-ead1-41fc-933a-3976baa23716"),
				},
				Then: mockserver.Response(http.StatusNoContent, nil),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name: "Delete ASM group",
			Input: common.DeleteParams{
				ObjectName: objectASMGroups,
				RecordId:   "12345",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/v3/asm/groups/12345"),
				},
				Then: mockserver.Response(http.StatusNoContent, nil),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableDeleter, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
