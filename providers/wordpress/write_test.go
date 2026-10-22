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

func TestWrite(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	errorEditing := testutils.DataFromFile(t, "write/cannot-edit-record.json")
	responsePosts := testutils.DataFromFile(t, "write/posts.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Error response on posts PATCH is understood",
			Input: common.WriteParams{ObjectName: "posts", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusForbidden, errorEditing),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrForbidden,
				testutils.StringError("rest_cannot_edit: Sorry, you are not allowed to edit this post."),
			},
		},
		{
			Name:  "Create a post",
			Input: common.WriteParams{ObjectName: "posts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/wp/v2/sites/test-workspace/posts"),
				},
				Then: mockserver.Response(http.StatusOK, responsePosts),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "17",
				Data: map[string]any{
					"title": map[string]any{
						"raw":      "Created via Postman",
						"rendered": "Created via Postman",
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update a post",
			Input: common.WriteParams{ObjectName: "posts", RecordData: "dummy", RecordId: "17"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/wp/v2/sites/test-workspace/posts/17"),
				},
				Then: mockserver.Response(http.StatusOK, responsePosts),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "17",
				Data: map[string]any{
					"title": map[string]any{
						"raw":      "Created via Postman",
						"rendered": "Created via Postman",
					},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWriter, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}
