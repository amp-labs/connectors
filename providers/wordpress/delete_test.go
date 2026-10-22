package wordpress

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestDelete(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	errorRemoving := testutils.DataFromFile(t, "write/already-deleted.json")

	tests := []testconn.TestCaseDelete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write object and its ID must be included",
			Input:        common.DeleteParams{ObjectName: "posts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:  "Error response on repeated removal is understood",
			Input: common.DeleteParams{ObjectName: "posts", RecordId: "17"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusGone, errorRemoving),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrNotFound,
				testutils.StringError("rest_already_trashed: The post has already been deleted."),
			},
		},
		{
			Name:  "Remove a post",
			Input: common.DeleteParams{ObjectName: "posts", RecordId: "17"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/wp/v2/sites/test-workspace/posts/17"),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableDeleter, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}
