package zoom

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseUsersFirstPage := testutils.DataFromFile(t, "read-users-first-page.json")
	responseUsersSecondPage := testutils.DataFromFile(t, "read-users-second-page.json")
	responseArchiveFilesFirstPage := testutils.DataFromFile(t, "archive-files-first-page.json")
	responseArchiveFilesSecondPage := testutils.DataFromFile(t, "archive-files-second-page.json")
	responseRecordingsFirstPage := testutils.DataFromFile(t, "recordings-first-page.json")
	responseRecordingsSecondPage := testutils.DataFromFile(t, "recordings-second-page.json")

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
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/users"),
				Then:  mockserver.Response(http.StatusOK, responseUsersFirstPage),
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
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/users"),
				Then:  mockserver.Response(http.StatusOK, responseUsersSecondPage),
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
		{
			Name: "Read Archive List first page",
			Input: common.ReadParams{
				ObjectName: "archive_files", Fields: connectors.Fields("id", "topic"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/archive_files"),
				Then:  mockserver.Response(http.StatusOK, responseArchiveFilesFirstPage),
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
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/archive_files"),
				Then:  mockserver.Response(http.StatusOK, responseArchiveFilesSecondPage),
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
		{
			Name: "Read Recordings first page",
			Input: common.ReadParams{
				ObjectName: "recordings",
				Fields:     connectors.Fields("id", "topic"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/users/me/recordings"),
					//mockcond.QueryParam("from", time.Now().AddDate(0, 0, -29).Format("2006-01-02")),
					//mockcond.QueryParam("to", time.Now().Format("2006-01-02")),
					//mockcond.QueryParamTimeApprox("from", time.Now().AddDate(0, 0, -29), "2006-01-02"),
					//mockcond.QueryParamTimeApprox("to", time.Now(), "2006-01-02"),
				},
				Then: mockserver.Response(http.StatusOK, responseRecordingsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    float64(6840331990),
						"topic": "My Personal Meeting",
					},
					Raw: map[string]any{
						"id":              float64(6840331990),
						"topic":           "My Personal Meeting",
						"account_id":      "Cx3wERazSgup7ZWRHQM8-w",
						"host_id":         "_0ctZtY0REqWalTmwvrdIw",
						"duration":        float64(20),
						"recording_count": float64(22),
					},
				}},
				// TODO do not use time.Now
				// Use custom comparator to handle this special case.
				NextPage: common.NextPageToken(testroutines.URLTestServer + "/v2/users/me/recordings?" +
					"from=" + time.Now().AddDate(0, 0, -29).Format("2006-01-02") + "&next_page_token=Tva2CuIdTgsv8wAnhyAdU3m06Y2HuLQtlh3&page_size=300&to=" + time.Now().Format("2006-01-02")), //nolint:lll
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read Recordings second page without next page token",
			Input: common.ReadParams{
				ObjectName: "recordings", Fields: connectors.Fields("id", "topic"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/users/me/recordings"),
				Then:  mockserver.Response(http.StatusOK, responseRecordingsSecondPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    float64(7951442001),
						"topic": "Team Standup Recording",
					},
					Raw: map[string]any{
						"id":              float64(7951442001),
						"topic":           "Team Standup Recording",
						"account_id":      "Dx4xFSbzTgvp8XWSIHN9-x",
						"host_id":         "_1duAuZ1SFrXblUnxwseJx",
						"duration":        float64(45),
						"recording_count": float64(5),
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Incremental read of Recordings with Since and Until filters",
			Input: common.ReadParams{
				ObjectName: "recordings",
				Fields:     connectors.Fields("id", "topic"),
				Since:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/users/me/recordings"),
					mockcond.QueryParam("from", "2024-01-01"),
					mockcond.QueryParam("to", "2024-03-31"),
				},
				Then: mockserver.Response(http.StatusOK, responseRecordingsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    float64(6840331990),
						"topic": "My Personal Meeting",
					},
					Raw: map[string]any{
						"id":              float64(6840331990),
						"topic":           "My Personal Meeting",
						"account_id":      "Cx3wERazSgup7ZWRHQM8-w",
						"host_id":         "_0ctZtY0REqWalTmwvrdIw",
						"duration":        float64(20),
						"recording_count": float64(22),
					},
				}},
				NextPage: testroutines.URLTestServer + "/v2/users/me/recordings?" +
					"from=2024-01-01&next_page_token=Tva2CuIdTgsv8wAnhyAdU3m06Y2HuLQtlh3&page_size=300&to=2024-03-31",
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Incremental read of Archive Files with Since filter",
			Input: common.ReadParams{
				ObjectName: "archive_files",
				Fields:     connectors.Fields("id", "topic"),
				Since:      time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/archive_files"),
					mockcond.QueryParam("from", "2024-06-01"),
				},
				Then: mockserver.Response(http.StatusOK, responseArchiveFilesFirstPage),
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
				NextPage: testroutines.URLTestServer + "/v2/archive_files?" +
					"from=2024-06-01&next_page_token=At6eWnFZ1FB3arCXnRxqHLXKhbDW18yz2i2&page_size=300",
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Users read does not add time filter params",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id", "email"),
				Since:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/users"),
					mockcond.QueryParamsMissing("from"),
					mockcond.QueryParamsMissing("to"),
				},
				Then: mockserver.Response(http.StatusOK, responseUsersFirstPage),
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
						"id":    "KDcuGIm1QgePTO8WbOqwIQ",
						"email": "jchill@example.com",
					},
				}},
				NextPage: testroutines.URLTestServer + "/v2/users?next_page_token=8V8HigQkzm2O5r9RUn31D9ZyJHgrmFfbLa2&page_size=300", //nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
