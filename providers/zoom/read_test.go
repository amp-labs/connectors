package zoom

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestReadModuleUser(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseUsersFirstPage := testutils.DataFromFile(t, "read-users-first-page.json")
	responseUsersSecondPage := testutils.DataFromFile(t, "read-users-second-page.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "users"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown objects are not supported",
			Input:        common.ReadParams{ObjectName: "tiger", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Read Users first page",
			Input: common.ReadParams{
				ObjectName: "users", Fields: connectors.Fields("id", "email"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseUsersFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    "KDcuGIm1QgePTO8WbOqwIQ",
						"email": "jchill@example.com",
					},
					Raw: map[string]any{
						"id":              "KDcuGIm1QgePTO8WbOqwIQ",
						"email":           "jchill@example.com",
						"user_created_at": "2019-06-01T07:58:03Z",
						"status":          "active",
					},
				}},
				NextPage: testroutines.URLTestServer + "/v2/users?next_page_token=8V8HigQkzm2O5r9RUn31D9ZyJHgrmFfbLa2&page_size=300", //nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read Users second page without next page token",
			Input: common.ReadParams{
				ObjectName: "users", Fields: connectors.Fields("id", "email"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseUsersSecondPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    "KfDcuGIdfdlfdQgePTO8WbOqwIQ",
						"email": "john@example.com",
					},
					Raw: map[string]any{
						"id":       "KfDcuGIdfdlfdQgePTO8WbOqwIQ",
						"email":    "john@example.com",
						"host_key": "2994fd2849",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL, common.ModuleID(providers.ModuleZoomUser))
			})
		})
	}
}

func TestReadModuleMeeting(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseArchiveFilesFirstPage := testutils.DataFromFile(t, "archive-files-first-page.json")
	responseArchiveFilesSecondPage := testutils.DataFromFile(t, "archive-files-second-page.json")

	tests := []testroutines.Read{
		{
			Name: "Read Archive List first page",
			Input: common.ReadParams{
				ObjectName: "archive_files", Fields: connectors.Fields("id", "topic"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseArchiveFilesFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    float64(553068284),
						"topic": "My Personal Meeting Room",
					},
					Raw: map[string]any{
						"id":                float64(553068284),
						"topic":             "My Personal Meeting Room",
						"timezone":          "Asia/Shanghai",
						"parent_meeting_id": "atsXxhSEQWit9t+U02HXNQ==",
					},
				}},
				NextPage: testroutines.URLTestServer + "/v2/archive_files?next_page_token=At6eWnFZ1FB3arCXnRxqHLXKhbDW18yz2i2&page_size=300", //nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},

		{
			Name: "Read Archive List Next page without next page token",
			Input: common.ReadParams{
				ObjectName: "archive_files", Fields: connectors.Fields("id", "topic"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseArchiveFilesSecondPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    float64(553032284),
						"topic": "My Personal Meeting Room",
					},
					Raw: map[string]any{
						"id":                float64(553032284),
						"topic":             "My Personal Meeting Room",
						"timezone":          "Asia/Shanghai",
						"parent_meeting_id": "atdfasXxhSEQWit9t+U02HXNQ==",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL, common.ModuleID(providers.ModuleZoomMeeting))
			})
		})
	}
}
