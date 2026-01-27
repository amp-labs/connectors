package capsule

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseUnprocessablePayload := testutils.DataFromFile(t, "write/unprocessable-payload.json")
	responseMethodNotAllowedPayload := testutils.DataFromFile(t, "write/method-not-allowed.json")
	responseTask := testutils.DataFromFile(t, "write/task/new.json")
	responseProject := testutils.DataFromFile(t, "write/project/new.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Error invalid payload",
			Input: common.WriteParams{ObjectName: "tasks", RecordData: map[string]any{}},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnprocessableEntity, responseUnprocessablePayload),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Validation Failed (dueOn is required)"),
			},
		},
		{
			Name:  "Error wrong http verb",
			Input: common.WriteParams{ObjectName: "tasks", RecordData: map[string]any{}},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusMethodNotAllowed, responseMethodNotAllowedPayload),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Method not allowed"),
			},
		},
		{
			Name: "Create task via POST",
			Input: common.WriteParams{
				ObjectName: "tasks",
				RecordData: map[string]any{
					"dueOn": "2025-05-20",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/v2/tasks"),
					mockcond.Body(`{"task": {"dueOn" : "2025-05-20"}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseTask),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "147248893",
				Errors:   nil,
				Data: map[string]any{
					"description":      "Brand new task from postman",
					"dueOn":            "2025-05-20",
					"taskDayDelayRule": "TRACK_DAYS",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update task via PUT",
			Input: common.WriteParams{
				ObjectName: "tasks",
				RecordId:   "147248893",
				RecordData: map[string]any{
					"dueOn": "2025-05-20",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/api/v2/tasks/147248893"),
					mockcond.Body(`{"task": {"dueOn" : "2025-05-20"}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseTask),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "147248893",
				Errors:   nil,
				Data: map[string]any{
					"description": "Brand new task from postman",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create project via POST",
			Input: common.WriteParams{
				ObjectName: "kases",
				RecordData: map[string]any{
					"name": "Research",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/v2/kases"),
					mockcond.Body(`{"kase": {"name" : "Research"}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseProject),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "5271516",
				Errors:   nil,
				Data: map[string]any{
					"name":        "Research",
					"description": "Project focused on business research",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update project via PUT",
			Input: common.WriteParams{
				ObjectName: "kases",
				RecordData: map[string]any{
					"name": "Research",
				},
				RecordId: "5271516",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/api/v2/kases/5271516"),
					mockcond.Body(`{"kase": {"name" : "Research"}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseProject),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "5271516",
				Errors:   nil,
				Data: map[string]any{
					"name":        "Research",
					"description": "Project focused on business research",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
