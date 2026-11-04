package notion

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestWrite(t *testing.T) {
	t.Parallel()

	createResp := []byte(`{"object":"file_upload","id":"file-1","status":"pending","filename":"notes.txt"}`)

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: objectUploads},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Users are not writable",
			Input: common.WriteParams{
				ObjectName: objectUsers,
				RecordData: map[string]any{"name": "Ada"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "File uploads do not support update",
			Input: common.WriteParams{
				ObjectName: objectUploads,
				RecordId:   "file-1",
				RecordData: map[string]any{"filename": "notes.txt"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create file upload",
			Input: common.WriteParams{
				ObjectName: objectUploads,
				RecordData: map[string]any{
					"mode":         "single_part",
					"filename":     "notes.txt",
					"content_type": "text/plain",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/file_uploads"),
					mockcond.Header(http.Header{"Notion-Version": []string{notionVersion}}),
					mockcond.Body(`{"content_type":"text/plain","filename":"notes.txt","mode":"single_part"}`),
				},
				Then: mockserver.Response(http.StatusOK, createResp),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "file-1",
				Data: map[string]any{
					"id":       "file-1",
					"status":   "pending",
					"filename": "notes.txt",
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
