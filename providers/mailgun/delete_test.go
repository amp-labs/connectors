package mailgun

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestDelete(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseBounceDelete := testutils.DataFromFile(t, "delete/bounce-delete.json")
	responseNotFound := testutils.DataFromFile(t, "delete/not-found.json")

	tests := []testconn.TestCaseDelete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Record id is required",
			Input:        common.DeleteParams{ObjectName: "lists"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			// messages is send-only; lists/members needs the parent list as a
			// second identifier — neither supports delete.
			Name: "Unsupported object returns operation-not-supported",
			Input: common.DeleteParams{
				ObjectName: "messages",
				RecordId:   "some-id",
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Domain-scoped delete without workspace errors",
			Input: common.DeleteParams{
				ObjectName: "bounces",
				RecordId:   "bounced@example.com",
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingDomain},
		},
		{
			Name: "Delete bounce substitutes domain into the item path",
			Input: common.DeleteParams{
				ObjectName: "bounces",
				RecordId:   "bounced@example.com",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/v3/example.com/bounces/bounced@example.com"),
				},
				Then: mockserver.Response(http.StatusOK, responseBounceDelete),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name: "Delete route",
			Input: common.DeleteParams{
				ObjectName: "routes",
				RecordId:   "4f3bad2335335426750048c6",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/v3/routes/4f3bad2335335426750048c6"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{"message":"Route has been deleted"}`)),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			// Deleting a missing record surfaces Mailgun's {"message": ...} error
			// (real captured response). common.InterpretError classes 404 as
			// retryable.
			Name: "Delete of a missing record maps to an error",
			Input: common.DeleteParams{
				ObjectName: "bounces",
				RecordId:   "ghost@example.com",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodDELETE(),
				Then:  mockserver.Response(http.StatusNotFound, responseNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrRetryable,
				testutils.StringError("Address not found in bounces table"),
			},
		},
	}

	// Domain-scoped objects require a workspace (Mailgun sending domain).
	workspaceByTest := map[string]string{
		"Delete bounce substitutes domain into the item path": "example.com",
		"Delete of a missing record maps to an error":         "example.com",
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableDeleter, error) {
				return constructReadTestConnector(tt.Server.URL, workspaceByTest[tt.Name])
			})
		})
	}
}
