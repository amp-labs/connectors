package mail

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { // nolint:funlen,maintidx
	t.Parallel()

	noteCreateResponse := testutils.DataFromFile(t, "write_note_create.json")
	taskCreateResponse := testutils.DataFromFile(t, "write_task_create.json")
	folderCreateResponse := testutils.DataFromFile(t, "write_folder_create.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Object name is required",
			Input:        common.WriteParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Record data is required",
			Input:        common.WriteParams{ObjectName: "notes"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "Unsupported object",
			Input:        common.WriteParams{ObjectName: "accounts", RecordData: map[string]any{}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},
		{
			Name: "Create a note",
			Input: common.WriteParams{
				ObjectName: "notes",
				RecordData: map[string]any{"title": "note title", "content": "note desc"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/api/notes/me"),
					},
					Then: mockserver.Response(http.StatusCreated, noteCreateResponse),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1711974988431110001",
				Data:     map[string]any{"entityId": "1711974988431110001"},
				Errors:   nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create a task",
			Input: common.WriteParams{
				ObjectName: "tasks",
				RecordData: map[string]any{"title": "Ampersand task"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/api/tasks/me"),
					},
					Then: mockserver.Response(http.StatusOK, taskCreateResponse),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "48184000000093001",
				Data:     map[string]any{"title": "Ampersand task"},
				Errors:   nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create an account-scoped folder",
			Input: common.WriteParams{
				ObjectName: "accounts/folders",
				RecordData: map[string]any{"folderName": "new"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/api/accounts/" + testAccountID + "/folders"),
					},
					Then: mockserver.Response(http.StatusCreated, folderCreateResponse),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "2560636000000076007",
				Data:     map[string]any{"folderName": "new"},
				Errors:   nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update account folder",
			Input: common.WriteParams{
				ObjectName: "accounts/folders",
				RecordId:   "2560636000000076007",
				RecordData: map[string]any{"mode": "rename", "folderName": "renamed"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPUT(),
						mockcond.Path("/api/accounts/" + testAccountID + "/folders/2560636000000076007"),
					},
					Then: mockserver.ResponseString(http.StatusOK,
						`{"status": {"code": 200, "description": "success"}}`),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			// The rename endpoint returns status only (no data object), so we report
			// success without a record id.
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tc := testconn.TestCase[common.WriteParams, *common.WriteResult](tt)
			t.Cleanup(tc.Close)

			adapter := constructTestAdapter(t, tt.Server.URL, testAccountID)

			output, err := adapter.Write(t.Context(), tc.Input)
			tc.Validate(t, err, output)
		})
	}
}
