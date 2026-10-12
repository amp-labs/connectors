package notion

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen
	t.Parallel()

	usersBody := testutils.DataFromFile(t, "read-users.json")
	usersLastBody := testutils.DataFromFile(t, "read-users-last.json")
	uploadsBody := testutils.DataFromFile(t, "read-file-uploads.json")

	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: objectUsers},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown object name is not supported",
			Input:        common.ReadParams{ObjectName: "pages", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Read users first page",
			Input: common.ReadParams{
				ObjectName: objectUsers,
				Fields:     connectors.Fields("name", "type"),
				Since:      since,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/users"),
					mockcond.QueryParam(pageSizeKey, "100"),
					mockcond.QueryParamsMissing(startCursorKey),
					mockcond.Header(http.Header{"Notion-Version": []string{notionVersion}}),
				},
				Then: mockserver.Response(http.StatusOK, usersBody),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows:     1,
				Done:     false,
				NextPage: testconn.URLTestServer + "/v1/users?page_size=100&start_cursor=cursor-2",
				Data: []common.ReadResultRow{
					{
						Id: "user-1",
						Fields: map[string]any{
							"name": "Ada",
							"type": "person",
						},
						Raw: map[string]any{
							"id":   "user-1",
							"name": "Ada",
							"type": "person",
						},
					},
				},
			},
		},
		{
			Name: "Read users last page",
			Input: common.ReadParams{
				ObjectName: objectUsers,
				Fields:     connectors.Fields("name"),
				NextPage:   testconn.URLTestServer + "/v1/users?page_size=100&start_cursor=cursor-2",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/users"),
					mockcond.QueryParam(startCursorKey, "cursor-2"),
					mockcond.Header(http.Header{"Notion-Version": []string{notionVersion}}),
				},
				Then: mockserver.Response(http.StatusOK, usersLastBody),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Done: true,
				Data: []common.ReadResultRow{
					{
						Id: "user-2",
						Fields: map[string]any{
							"name": "Bot",
						},
						Raw: map[string]any{
							"id":   "user-2",
							"name": "Bot",
						},
					},
				},
			},
		},
		{
			Name: "Read file uploads filters last_edited_time",
			Input: common.ReadParams{
				ObjectName: objectUploads,
				Fields:     connectors.Fields("status"),
				Since:      since,
				Until:      until,
				PageSize:   25,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/file_uploads"),
					mockcond.QueryParam(pageSizeKey, "25"),
					mockcond.Header(http.Header{"Notion-Version": []string{notionVersion}}),
				},
				Then: mockserver.Response(http.StatusOK, uploadsBody),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Done: true,
				Data: []common.ReadResultRow{
					{
						Id: "file-new",
						Fields: map[string]any{
							"status": "uploaded",
						},
						Raw: map[string]any{
							"id":     "file-new",
							"status": "uploaded",
						},
					},
				},
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
