package zoom

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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	errorBadRequest := testutils.DataFromFile(t, "bad-request.json")
	responseUsersFirstPage := testutils.DataFromFile(t, "read/users/1-first-page.json")
	responseUsersSecondPage := testutils.DataFromFile(t, "read/users/2-second-page.json")
	responseArchiveFilesFirstPage := testutils.DataFromFile(t, "read/archive_files/1-first-page.json")
	responseArchiveFilesSecondPage := testutils.DataFromFile(t, "read/archive_files/2-second-page.json")
	responseRecordingsFirstPage := testutils.DataFromFile(t, "read/recordings/1-first-page.json")
	responseRecordingsSecondPage := testutils.DataFromFile(t, "read/recordings/2-second-page.json")
	responseUser1MeetingsFirstPage := testutils.DataFromFile(t, "read/account_meetings/user-a/1-first-page.json")
	responseUser2MeetingsFirstPage := testutils.DataFromFile(t, "read/account_meetings/user-b/1-first-page.json")
	responseUser2MeetingsSecondPage := testutils.DataFromFile(t, "read/account_meetings/user-b/2-second-page.json")

	tests := []testconn.TestCaseRead{
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
			Name:  "Error response is understood",
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorBadRequest),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("error code (4711): " +
					"Invalid access token, does not contain scopes:[device:read:list_zdm_devices:admin"),
			},
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
			Comparator: testconn.ComparatorSubsetRead,
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
				NextPage: testconn.URLTestServer + "/v2/users?next_page_token=8V8HigQkzm2O5r9RUn31D9ZyJHgrmFfbLa2&page_size=300", //nolint:lll
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
			Comparator: testconn.ComparatorSubsetRead,
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
			Comparator: testconn.ComparatorSubsetRead,
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
				NextPage: testconn.URLTestServer + "/v2/archive_files?next_page_token=At6eWnFZ1FB3arCXnRxqHLXKhbDW18yz2i2&page_size=300", //nolint:lll
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
			Comparator: testconn.ComparatorSubsetRead,
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
		// TODO cannot use exact matches for time.Now() in unit test.
		//{
		//	Name: "Read Recordings first page",
		//	Input: common.ReadParams{
		//		ObjectName: "recordings",
		//		Fields:     connectors.Fields("id", "topic"),
		//	},
		//	Server: mockserver.Conditional{
		//		Setup: mockserver.ContentJSON(),
		//		If: mockcond.And{
		//			mockcond.Path("/v2/users/me/recordings"),
		//			mockcond.QueryParam("from", time.Now().AddDate(0, 0, -29).Format("2006-01-02")),
		//			mockcond.QueryParam("to", time.Now().Format("2006-01-02")),
		//		},
		//		Then: mockserver.Response(http.StatusOK, responseRecordingsFirstPage),
		//	}.Server(),
		//	Comparator: testconn.ComparatorSubsetRead,
		//	Expected: &common.ReadResult{
		//		Rows: 1,
		//		Data: []common.ReadResultRow{{
		//			Fields: map[string]any{
		//				"id":    float64(6840331990),
		//				"topic": "My Personal Meeting",
		//			},
		//			Raw: map[string]any{
		//				"id":              float64(6840331990),
		//				"topic":           "My Personal Meeting",
		//				"account_id":      "Cx3wERazSgup7ZWRHQM8-w",
		//				"host_id":         "_0ctZtY0REqWalTmwvrdIw",
		//				"duration":        float64(20),
		//				"recording_count": float64(22),
		//			},
		//		}},
		//		NextPage: common.NextPageToken(testconn.URLTestServer + "/v2/users/me/recordings?" +
		//			"from=" + time.Now().AddDate(0, 0, -29).Format("2006-01-02") + "&next_page_token=Tva2CuIdTgsv8wAnhyAdU3m06Y2HuLQtlh3&page_size=300&to=" + time.Now().Format("2006-01-02")), //nolint:lll
		//		Done: false,
		//	},
		//	ExpectedErrs: nil,
		//},
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
			Comparator: testconn.ComparatorSubsetRead,
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
			Comparator: testconn.ComparatorSubsetRead,
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
				NextPage: testconn.URLTestServer + "/v2/users/me/recordings?" +
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
			Comparator: testconn.ComparatorSubsetRead,
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
				NextPage: testconn.URLTestServer + "/v2/archive_files?" +
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
			Comparator: testconn.ComparatorSubsetRead,
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
				NextPage: testconn.URLTestServer + "/v2/users?next_page_token=8V8HigQkzm2O5r9RUn31D9ZyJHgrmFfbLa2&page_size=300", //nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read Meetings first page",
			Input: common.ReadParams{
				ObjectName: "meetings",
				Fields:     connectors.Fields("topic"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/users/me/meetings"),
				},
				Then: mockserver.Response(http.StatusOK, responseUser2MeetingsFirstPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"topic": "Team Coffee Chat & Trivia"},
					Id:     "86535688426",
					Raw:    map[string]any{"created_at": "2026-10-30T20:04:50Z"},
				}, {
					Fields: map[string]any{"topic": "Virtual Happy Hour"},
					Id:     "84059314795",
					Raw:    map[string]any{"created_at": "2026-10-30T20:05:26Z"},
				}},
				NextPage: testconn.URLTestServer + "/v2/users/me/meetings?next_page_token=i2Gm6qVgwKn5yteiAwgyLVWEOBzA7Pb3TH2&page_size=300",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read Account Meetings first page across all users",
			Input: common.ReadParams{
				ObjectName: "account_meetings",
				Fields:     connectors.Fields("id", "topic"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.Path("/v2/users"),
						mockcond.QueryParam("page_size", "300"),
						mockcond.QueryParamsMissing("next_page_token"),
					},
					Then: mockserver.Response(http.StatusOK, responseUsersFirstPage),
				}, {
					If: mockcond.And{
						mockcond.Path("/v2/users"),
						mockcond.QueryParam("next_page_token", "8V8HigQkzm2O5r9RUn31D9ZyJHgrmFfbLa2"),
					},
					Then: mockserver.Response(http.StatusOK, responseUsersSecondPage),
				}, {
					If:   mockcond.Path("/v2/users/KDcuGIm1QgePTO8WbOqwIQ/meetings"),
					Then: mockserver.Response(http.StatusOK, responseUser1MeetingsFirstPage),
				}, {
					If:   mockcond.Path("/v2/users/KfDcuGIdfdlfdQgePTO8WbOqwIQ/meetings"),
					Then: mockserver.Response(http.StatusOK, responseUser2MeetingsFirstPage),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetReadSorted,
			Expected: &common.ReadResult{
				Rows: 4,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"topic": "Virtual Happy Hour"},
					Id:     "84059314795",
					Raw:    map[string]any{"created_at": "2026-10-30T20:05:26Z"},
				}, {
					Fields: map[string]any{"topic": "Q4 Departmental Alignment & Sync"},
					Id:     "86381714576",
					Raw:    map[string]any{"created_at": "2026-10-30T20:11:05Z"},
				}, {
					Fields: map[string]any{"topic": "Team Coffee Chat & Trivia"},
					Id:     "86535688426",
					Raw:    map[string]any{"created_at": "2026-10-30T20:04:50Z"},
				}, {
					Fields: map[string]any{"topic": "Quick sync"},
					Id:     "87414255009",
					Raw:    map[string]any{"created_at": "2026-10-30T20:11:45Z"},
				}},
				NextPage: `[{
					"context": "KfDcuGIdfdlfdQgePTO8WbOqwIQ",
					"value": "` + testconn.URLTestServer + `/v2/users/KfDcuGIdfdlfdQgePTO8WbOqwIQ/meetings?next_page_token=i2Gm6qVgwKn5yteiAwgyLVWEOBzA7Pb3TH2&page_size=300"
				}]`,
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Account meetings correctly navigates to the next page using aggregate page token",
			Input: common.ReadParams{
				ObjectName: "account_meetings",
				Fields:     connectors.Fields("topic"),
				NextPage: `[{
					"context": "KfDcuGIdfdlfdQgePTO8WbOqwIQ",
					"value": "` + testconn.URLTestServer + `/v2/users/KfDcuGIdfdlfdQgePTO8WbOqwIQ/meetings?next_page_token=i2Gm6qVgwKn5yteiAwgyLVWEOBzA7Pb3TH2&page_size=300"
				}]`,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/users/KfDcuGIdfdlfdQgePTO8WbOqwIQ/meetings"),
					mockcond.QueryParam("next_page_token", "i2Gm6qVgwKn5yteiAwgyLVWEOBzA7Pb3TH2"),
					mockcond.QueryParam("page_size", "300"),
				},
				Then: mockserver.Response(http.StatusOK, responseUser2MeetingsSecondPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetReadSorted,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"topic": "Database Migration & Architecture Sync"},
					Id:     "83625651564",
					Raw:    map[string]any{"created_at": "2026-10-30T20:06:17Z"},
				}, {
					Fields: map[string]any{"topic": "Project Alpha: Weekly Status Update"},
					Id:     "88033623689",
					Raw:    map[string]any{"created_at": "2026-10-30T20:06:17Z"},
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
			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}
