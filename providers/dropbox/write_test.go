package dropbox

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) {
	t.Parallel()

	createResp := testutils.DataFromFile(t, "write/file-request-create.json")
	updateResp := testutils.DataFromFile(t, "write/file-request-update.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: objectFileRequests},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Files are not writable",
			Input: common.WriteParams{
				ObjectName: objectFiles,
				RecordData: map[string]any{"name": "notes.txt"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create file request",
			Input: common.WriteParams{
				ObjectName: objectFileRequests,
				RecordData: map[string]any{
					"title":       "Upload receipts",
					"destination": "/FileRequests",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/2/file_requests/create"),
					mockcond.Body(`{"destination":"/FileRequests","title":"Upload receipts"}`),
				},
				Then: mockserver.Response(http.StatusOK, createResp),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "id:fr-new",
				Data: map[string]any{
					"id":    "id:fr-new",
					"title": "Upload receipts",
				},
			},
		},
		{
			Name: "Update file request",
			Input: common.WriteParams{
				ObjectName: objectFileRequests,
				RecordId:   "id:fr-1",
				RecordData: map[string]any{
					"title": "Upload receipts (closed)",
					"open":  false,
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/2/file_requests/update"),
					mockcond.Body(`{"id":"id:fr-1","open":false,"title":"Upload receipts (closed)"}`),
				},
				Then: mockserver.Response(http.StatusOK, updateResp),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "id:fr-1",
				Data: map[string]any{
					"id":      "id:fr-1",
					"is_open": false,
				},
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
