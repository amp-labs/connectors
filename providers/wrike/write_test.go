package wrike

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	responseTaskUpdate := testutils.DataFromFile(t, "task-update.json")
	responseCommentUpdate := testutils.DataFromFile(t, "comment-update.json")
	responseCustomFieldCreate := testutils.DataFromFile(t, "customfield-create.json")
	responseTimesheetCreate := testutils.DataFromFile(t, "timesheet-create.json")
	responseNotAllowed := testutils.DataFromFile(t, "error-not-allowed.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: objectTasks},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.WriteParams{ObjectName: "account", RecordData: map[string]any{"name": "x"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Read-only object is not writable",
			Input:        common.WriteParams{ObjectName: objectAuditLog, RecordData: map[string]any{"x": "y"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Update-only object rejects create",
			Input:        common.WriteParams{ObjectName: objectContacts, RecordData: map[string]any{"firstName": "x"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Task create is not supported (created under a folder in Wrike)",
			Input:        common.WriteParams{ObjectName: objectTasks, RecordData: map[string]any{"title": "orphan"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Update task strips the id and uses PUT",
			Input: common.WriteParams{
				ObjectName: objectTasks,
				RecordId:   "MAAAAAEQ_kSH",
				RecordData: map[string]any{"id": "MAAAAAEQ_kSH", "title": "Connector JSON probe 2", "status": "Completed"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/api/v4/tasks/MAAAAAEQ_kSH"),
					mockcond.Body(`{"status":"Completed","title":"Connector JSON probe 2"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseTaskUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "MAAAAAEQ_kSH",
				Data:     map[string]any{"id": "MAAAAAEQ_kSH", "title": "Connector JSON probe 2", "status": "Completed"},
			},
		},
		{
			Name: "Update comment uses its own item path",
			Input: common.WriteParams{
				ObjectName: objectComments,
				RecordId:   "IEAG5DBKIMCUAHY2",
				RecordData: map[string]any{"text": "probe comment (updated)"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/api/v4/comments/IEAG5DBKIMCUAHY2"),
					mockcond.Body(`{"text":"probe comment (updated)"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseCommentUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "IEAG5DBKIMCUAHY2",
				Data:     map[string]any{"id": "IEAG5DBKIMCUAHY2", "text": "probe comment (updated)"},
			},
		},
		{
			Name: "Create custom field at account level",
			Input: common.WriteParams{
				ObjectName: objectCustomFields,
				RecordData: map[string]any{"title": "Connector Priority", "type": "Text"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/v4/customfields"),
					mockcond.Body(`{"title":"Connector Priority","type":"Text"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomFieldCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "IEAG5DBKJUAAAAAA",
				Data:     map[string]any{"id": "IEAG5DBKJUAAAAAA", "title": "Connector Priority", "type": "Text"},
			},
		},
		{
			Name: "Create timesheet returns timesheetId",
			Input: common.WriteParams{
				ObjectName: objectTimesheets,
				RecordData: map[string]any{"periodStartDate": "2026-10-06", "timeframe": "Week"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/v4/timesheets"),
				},
				Then: mockserver.Response(http.StatusOK, responseTimesheetCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "TS-NEW",
				Data:     map[string]any{"timesheetId": "TS-NEW", "timeframe": "Week"},
			},
		},
		{
			Name: "Disallowed parameter surfaces as caller error",
			Input: common.WriteParams{
				ObjectName: objectTasks,
				RecordId:   "MAAAAAEQ_kSH",
				RecordData: map[string]any{"createdDate": "2026-01-01T00:00:00Z"},
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseNotAllowed),
			}.Server(),
			ExpectedErrs: []error{common.ErrCaller},
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
