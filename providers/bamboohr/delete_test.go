package bamboohr

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
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Unsupported object is rejected",
			Input:        common.DeleteParams{ObjectName: objectTimeOffRequests, RecordId: "1"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Alert configurations are write-only",
			Input:        common.DeleteParams{ObjectName: objectAlertConfigurations, RecordId: "1"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Delete webhook",
			Input: common.DeleteParams{ObjectName: objectWebhooks, RecordId: "4"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/api/v1/webhooks/4"),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:  "Delete time tracking project",
			Input: common.DeleteParams{ObjectName: objectTimeTrackingProjects, RecordId: "42"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/api/v1/time-tracking/projects/42"),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:  "Delete holiday",
			Input: common.DeleteParams{ObjectName: objectHolidays, RecordId: "12"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/api/v1/holidays/12"),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:  "Delete employee",
			Input: common.DeleteParams{ObjectName: objectEmployees, RecordId: "15"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/api/v1/employees/15"),
				},
				Then: mockserver.Response(http.StatusNoContent),
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
