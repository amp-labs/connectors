package reply

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestDelete(t *testing.T) { // nolint:funlen
	t.Parallel()

	tests := []testconn.TestCaseDelete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Record id must be included",
			Input:        common.DeleteParams{ObjectName: "contacts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:         "Read-only object is rejected by the registry",
			Input:        common.DeleteParams{ObjectName: "inbox/threads", RecordId: "1"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Successful contact deletion returns 204",
			Input: common.DeleteParams{ObjectName: "contacts", RecordId: "761575530"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/v3/contacts/761575530"),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
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
