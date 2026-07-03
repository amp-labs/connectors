package acculynx

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestDelete(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []testroutines.TestCaseDelete{
		{
			Name:         "Delete param object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Delete param record id must be included",
			Input:        common.DeleteParams{ObjectName: "jobs/representatives/ar-owner"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name: "Unsupported object returns ErrOperationNotSupportedForObject",
			Input: common.DeleteParams{
				ObjectName: "jobs",
				RecordId:   "9ecc68c2-9beb-4b8f-a4b5-6f4e52a41d75",
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Successful AR-owner deletion",
			Input: common.DeleteParams{
				ObjectName: "jobs/representatives/ar-owner",
				RecordId:   "9ecc68c2-9beb-4b8f-a4b5-6f4e52a41d75",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/jobs/9ecc68c2-9beb-4b8f-a4b5-6f4e52a41d75/representatives/ar-owner"),
					mockcond.MethodDELETE(),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful sales-owner deletion (204 No Content)",
			Input: common.DeleteParams{
				ObjectName: "jobs/representatives/sales-owner",
				RecordId:   "9ecc68c2-9beb-4b8f-a4b5-6f4e52a41d75",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/jobs/9ecc68c2-9beb-4b8f-a4b5-6f4e52a41d75/representatives/sales-owner"),
					mockcond.MethodDELETE(),
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

			tt.Run(t, func() (testroutines.TestableDeleter, error) {
				return constructTestReadConnector(tt.Server.URL)
			})
		})
	}
}
