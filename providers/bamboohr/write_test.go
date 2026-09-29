package bamboohr

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseWebhookCreate := testutils.DataFromFile(t, "write/webhook-create.json")
	responseWebhookUpdate := testutils.DataFromFile(t, "write/webhook-update.json")
	responseProjectCreate := testutils.DataFromFile(t, "write/time-tracking-project-create.json")
	responseTimeOffCreate := testutils.DataFromFile(t, "write/time-off-request-create.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Unsupported object is rejected",
			Input:        common.WriteParams{ObjectName: objectEmployees, RecordData: map[string]any{"x": "y"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create webhook",
			Input: common.WriteParams{
				ObjectName: objectWebhooks,
				RecordData: map[string]any{
					"name":   "Example Webhook",
					"url":    "https://www.example.com/hook",
					"format": "json",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/v1/webhooks"),
				},
				Then: mockserver.Response(http.StatusCreated, responseWebhookCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "4",
				Data:     map[string]any{"id": "4", "name": "Example Webhook"},
			},
		},
		{
			Name: "Update webhook",
			Input: common.WriteParams{
				ObjectName: objectWebhooks,
				RecordId:   "4",
				RecordData: map[string]any{
					"name":   "Updated Webhook",
					"url":    "https://www.example.com/hook",
					"format": "json",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/api/v1/webhooks/4"),
				},
				Then: mockserver.Response(http.StatusOK, responseWebhookUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "4",
				Data:     map[string]any{"id": "4", "name": "Updated Webhook"},
			},
		},
		{
			Name: "Create time tracking project",
			Input: common.WriteParams{
				ObjectName: objectTimeTrackingProjects,
				RecordData: map[string]any{"name": "Client Work", "billable": true},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/v1/time-tracking/projects"),
				},
				Then: mockserver.Response(http.StatusCreated, responseProjectCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "42",
				Data:     map[string]any{"id": float64(42), "name": "Client Work"},
			},
		},
		{
			Name: "Update time tracking project",
			Input: common.WriteParams{
				ObjectName: objectTimeTrackingProjects,
				RecordId:   "42",
				RecordData: map[string]any{"name": "Client Work Renamed"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/api/v1/time-tracking/projects/42"),
				},
				Then: mockserver.Response(http.StatusOK, responseProjectCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "42",
				Data:     map[string]any{"id": float64(42), "name": "Client Work"},
			},
		},
		{
			Name: "Create time off request",
			Input: common.WriteParams{
				ObjectName: objectTimeOffRequests,
				RecordData: map[string]any{
					"employeeId": 123,
					"startDate":  "2026-06-01",
					"endDate":    "2026-06-03",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/v1/time-off/requests"),
				},
				Then: mockserver.Response(http.StatusCreated, responseTimeOffCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "601",
				Data:     map[string]any{"id": float64(601), "employeeId": float64(123)},
			},
		},
		{
			Name: "Update time off request",
			Input: common.WriteParams{
				ObjectName: objectTimeOffRequests,
				RecordId:   "601",
				RecordData: map[string]any{"employeeNote": "Updated note"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/api/v1/time-off/requests/601"),
				},
				Then: mockserver.Response(http.StatusOK, responseTimeOffCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "601",
				Data:     map[string]any{"id": float64(601)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWriter, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
