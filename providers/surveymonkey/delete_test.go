package surveymonkey

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
			Input:        common.DeleteParams{ObjectName: objectContacts},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:         "Unsupported object",
			Input:        common.DeleteParams{ObjectName: objectGroups, RecordId: "1"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Delete contact",
			Input: common.DeleteParams{
				ObjectName: objectContacts,
				RecordId:   "1234",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/v3/contacts/1234"),
				},
				Then: mockserver.Response(http.StatusNoContent, nil),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name: "Delete contact list",
			Input: common.DeleteParams{
				ObjectName: objectContactLists,
				RecordId:   "5678",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/v3/contact_lists/5678"),
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
