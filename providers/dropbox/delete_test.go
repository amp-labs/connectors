package dropbox

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	deleteResp := testutils.DataFromFile(t, "delete/file-request-delete.json")

	tests := []testconn.TestCaseDelete{
		{
			Name:         "Object name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Record id is required",
			Input:        common.DeleteParams{ObjectName: objectFileRequests},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name: "Shared links cannot be deleted",
			Input: common.DeleteParams{
				ObjectName: objectSharedLinks,
				RecordId:   "id:link1",
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Delete closed file request",
			Input: common.DeleteParams{
				ObjectName: objectFileRequests,
				RecordId:   "id:fr-closed",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/2/file_requests/delete"),
					mockcond.Body(`{"ids":["id:fr-closed"]}`),
				},
				Then: mockserver.Response(http.StatusOK, deleteResp),
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
