package twenty

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

	responseDelete := testutils.DataFromFile(t, "delete-vendor-contract.json")
	responseNotFound := testutils.DataFromFile(t, "error-record-not-found.json")

	tests := []testconn.TestCaseDelete{
		{
			Name:         "Record id must be included",
			Input:        common.DeleteParams{ObjectName: "vendorContracts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:  "Delete asks for a soft delete explicitly",
			Input: common.DeleteParams{ObjectName: "vendorContracts", RecordId: vendorContractID},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/rest/vendorContracts/" + vendorContractID),
					mockcond.QueryParam("soft_delete", "true"),
				},
				Then: mockserver.Response(http.StatusOK, responseDelete),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "A 2xx without the deleted record's id is an error",
			Input: common.DeleteParams{ObjectName: "vendorContracts", RecordId: vendorContractID},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, []byte(`{"data":{}}`)),
			}.Server(),
			ExpectedErrs: []error{errDeleteResponse},
		},
		{
			Name:  "A record already deleted or missing is a non-retryable not found",
			Input: common.DeleteParams{ObjectName: "vendorContracts", RecordId: vendorContractID},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseNotFound),
			}.Server(),
			ExpectedErrs: []error{common.ErrNotFound, testutils.StringError("Record not found")},
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
