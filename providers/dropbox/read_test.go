package dropbox

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseFilesFirstPage := testutils.DataFromFile(t, "read/files-first-page.json")
	responseFilesLastPage := testutils.DataFromFile(t, "read/files-last-page.json")
	responseSharedFolders := testutils.DataFromFile(t, "read/shared-folders.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: objectFiles},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.ReadParams{ObjectName: "search", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "List folder root posts path and limit",
			Input: common.ReadParams{
				ObjectName: objectFiles,
				Fields:     connectors.Fields("name", "server_modified"),
				PageSize:   50,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/2/files/list_folder"),
					mockcond.Body(`{"limit":50,"path":"","recursive":false}`),
				},
				Then: mockserver.Response(http.StatusOK, responseFilesFirstPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id: "id:abc123",
					Fields: map[string]any{
						"name":            "notes.txt",
						"server_modified": "2024-06-01T12:00:00Z",
					},
					Raw: map[string]any{".tag": "file"},
				}},
				NextPage: "cursor-page-2",
				Done:     false,
			},
		},
		{
			Name: "List folder continue sends cursor only",
			Input: common.ReadParams{
				ObjectName: objectFiles,
				Fields:     connectors.Fields("name"),
				NextPage:   "cursor-page-2",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/2/files/list_folder/continue"),
					mockcond.Body(`{"cursor":"cursor-page-2"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseFilesLastPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id:     "id:folder1",
					Fields: map[string]any{"name": "Photos"},
					Raw:    map[string]any{".tag": "folder"},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Shared folders use shared_folder_id as record id",
			Input: common.ReadParams{
				ObjectName: objectSharedFolders,
				Fields:     connectors.Fields("name", "time_invited"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/2/sharing/list_folders"),
					mockcond.Body(`{"limit":100}`),
				},
				Then: mockserver.Response(http.StatusOK, responseSharedFolders),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id: "84528196421",
					Fields: map[string]any{
						"name":         "Team Share",
						"time_invited": "2024-03-10T08:30:00Z",
					},
					Raw: map[string]any{"name": "Team Share"},
				}},
				Done: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
