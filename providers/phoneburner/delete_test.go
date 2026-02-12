package phoneburner

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

func TestDelete(t *testing.T) { //nolint:funlen,cyclop
	t.Parallel()

	respUnauthorized := testutils.DataFromFile(t, "read/error-unauthorized.json")

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write object and its ID must be included",
			Input:        common.DeleteParams{ObjectName: "contacts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:         "Unknown objects are not supported",
			Input:        common.DeleteParams{ObjectName: "tiger", RecordId: "1"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Delete contact",
			Input: common.DeleteParams{ObjectName: "contacts", RecordId: "30919347"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/rest/1/contacts/30919347"),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:  "Delete folder",
			Input: common.DeleteParams{ObjectName: "folders", RecordId: "11888"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/rest/1/folders/11888"),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:  "Delete member",
			Input: common.DeleteParams{ObjectName: "members", RecordId: "25381104"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/rest/1/members/25381104"),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:  "Delete tag",
			Input: common.DeleteParams{ObjectName: "tags", RecordId: "1"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/rest/1/tags/1"),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:  "Provider envelope error is mapped for delete (200 with http_status=401)",
			Input: common.DeleteParams{ObjectName: "contacts", RecordId: "30919347"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/rest/1/contacts/30919347"),
				},
				Then: mockserver.Response(http.StatusOK, respUnauthorized),
			}.Server(),
			ExpectedErrs: []error{common.ErrAccessToken},
		},
		{
			Name:         "Unsupported delete object returns not supported",
			Input:        common.DeleteParams{ObjectName: "dialsession", RecordId: "1"},
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
