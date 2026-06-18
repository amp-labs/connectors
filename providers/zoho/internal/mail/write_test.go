package mail

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestWrite(t *testing.T) { // nolint:funlen
	t.Parallel()

	tests := []testroutines.Write{
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
			Name: "Update is not supported",
			Input: common.WriteParams{
				ObjectName: "notes",
				RecordId:   "1781549700389120100",
				RecordData: map[string]any{"title": "edited"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
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
					Then: mockserver.ResponseString(http.StatusCreated, `{
						"status": {"code": 201, "description": "Created"},
						"data": {
							"entityId": "1711974988431110001",
							"URI": "https://mail.zoho.com/api/notes/me/1711974988431110001"
						}
					}`),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1711974988431110001",
				Data:     map[string]any{"entityId": "1711974988431110001"},
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
					Then: mockserver.ResponseString(http.StatusCreated, `{
						"status": {"code": 201, "description": "Created"},
						"data": {"folderId": "2560636000000076001", "folderName": "new"}
					}`),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "2560636000000076001",
				Data:     map[string]any{"folderName": "new"},
				Errors:   nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Account-scoped write needs the account id",
			Input: common.WriteParams{
				ObjectName: "accounts/folders",
				RecordData: map[string]any{"folderName": "new"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrMissingAccountID},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tc := testroutines.TestCase[common.WriteParams, *common.WriteResult](tt)
			t.Cleanup(tc.Close)

			// The last case exercises the missing-account-id path, so its adapter
			// is built without an account id.
			accountID := testAccountID
			if tt.Name == "Account-scoped write needs the account id" {
				accountID = ""
			}

			adapter := constructTestAdapter(t, tt.Server.URL, accountID)

			output, err := adapter.Write(t.Context(), tc.Input)
			tc.Validate(t, err, output)
		})
	}
}
