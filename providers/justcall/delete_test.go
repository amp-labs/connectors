package justcall

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

func TestDelete(t *testing.T) { //nolint:funlen
	t.Parallel()

	deleteSuccessResponse := testutils.DataFromFile(t, "delete/success.json")

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Delete record ID must be included",
			Input:        common.DeleteParams{ObjectName: "contacts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:  "Delete contact with query param",
			Input: common.DeleteParams{ObjectName: "contacts", RecordId: "12345"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/contacts"),
					mockcond.MethodDELETE(),
					mockcond.QueryParam("id", "12345"),
				},
				Then: mockserver.Response(http.StatusOK, deleteSuccessResponse),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Delete tag with path ID",
			Input: common.DeleteParams{ObjectName: "tags", RecordId: "67890"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/texts/tags/67890"),
					mockcond.MethodDELETE(),
				},
				Then: mockserver.Response(http.StatusOK, deleteSuccessResponse),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Delete sales dialer contact with path ID",
			Input: common.DeleteParams{ObjectName: "sales_dialer/contacts", RecordId: "11111"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/sales_dialer/contacts/11111"),
					mockcond.MethodDELETE(),
				},
				Then: mockserver.Response(http.StatusOK, deleteSuccessResponse),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Delete webhook with path ID",
			Input: common.DeleteParams{ObjectName: "webhooks", RecordId: "22222"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/webhooks/url/22222"),
					mockcond.MethodDELETE(),
				},
				Then: mockserver.Response(http.StatusOK, deleteSuccessResponse),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:         "Delete unsupported object",
			Input:        common.DeleteParams{ObjectName: "calls", RecordId: "12345"},
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
