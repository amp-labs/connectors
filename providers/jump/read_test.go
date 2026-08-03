package jump

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

func TestRead(t *testing.T) {
	t.Parallel()

	tasksResponse := testutils.DataFromFile(t, "read-tasks.json")
	requestTasks := testutils.DataFromFile(t, "read/request/tasks.json")
	tasksPage2Response := testutils.DataFromFile(t, "read-tasks-page2.json")
	requestTasksPage2 := testutils.DataFromFile(t, "read/request/tasks-page2.json")
	meetingsResponse := testutils.DataFromFile(t, "read-meetings.json")
	requestMeetings := testutils.DataFromFile(t, "read/request/meetings.json")
	readErrorResponse := testutils.DataFromFile(t, "read-error.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "tasks"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "Successfully read tasks with assignee",
			Input: common.ReadParams{
				ObjectName: "tasks",
				Fields:     connectors.Fields("id", "title", "assignee"),
				PageSize:   2,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Body(string(requestTasks)),
				},
				Then: mockserver.Response(http.StatusOK, tasksResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":    "tsk_9e17843922b94782832",
							"title": "TEST",
						},
						Raw: map[string]any{
							"id":          "tsk_9e17843922b94782832",
							"title":       "TEST",
							"status":      "Todo",
							"meetingId":   "mtg_fbff3264957a48b4888",
							"description": "TEST",
							"assignee": map[string]any{
								"email":  "ada@example.com",
								"name":   "Amperasnd Developer",
								"source": "USER",
							},
						},
					},
					{
						Fields: map[string]any{
							"id":    "tsk_bde1c8b314e8459893e",
							"title": "Send follow-up email",
						},
						Raw: map[string]any{
							"id":        "tsk_bde1c8b314e8459893e",
							"title":     "Send follow-up email",
							"status":    "Todo",
							"meetingId": "mtg_fbff3264957a48b4888",
						},
					},
				},
				NextPage: "g3QAAAACdwJpZG0AAAAXdHNrX2JkZTFjOGIzMTRlODQ1OTg5M2V3C2luc2VydGVkX2F0dAAAAA13C21pY3Jvc2Vjb25kaAJiAAJEgWEGdwZzZWNvbmRhCXcIY2FsZW5kYXJ3E0VsaXhpci5DYWxlbmRhci5JU093BW1vbnRoYQZ3Cl9fc3RydWN0X193D0VsaXhpci5EYXRlVGltZXcKdXRjX29mZnNldGEAdwpzdGRfb2Zmc2V0YQB3BHllYXJiAAAH6ncEaG91cmEXdwNkYXlhF3cJem9uZV9hYmJybQAAAANVVEN3Bm1pbnV0ZWESdwl0aW1lX3pvbmVtAAAAB0V0Yy9VVEM=",
				Done:     false,
			},
		},
		{
			Name: "Read tasks second page using NextPage cursor",
			Input: common.ReadParams{
				ObjectName: "tasks",
				Fields:     connectors.Fields("id", "title"),
				PageSize:   2,
				NextPage:   common.NextPageToken("g3QAAAACdwJpZG0AAAAXdHNrX2JkZTFjOGIzMTRlODQ1OTg5M2V3C2luc2VydGVkX2F0dAAAAA13C21pY3Jvc2Vjb25kaAJiAAJEgWEGdwZzZWNvbmRhCXcIY2FsZW5kYXJ3E0VsaXhpci5DYWxlbmRhci5JU093BW1vbnRoYQZ3Cl9fc3RydWN0X193D0VsaXhpci5EYXRlVGltZXcKdXRjX29mZnNldGEAdwpzdGRfb2Zmc2V0YQB3BHllYXJiAAAH6ncEaG91cmEXdwNkYXlhF3cJem9uZV9hYmJybQAAAANVVEN3Bm1pbnV0ZWESdwl0aW1lX3pvbmVtAAAAB0V0Yy9VVEM="),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Body(string(requestTasksPage2)),
				},
				Then: mockserver.Response(http.StatusOK, tasksPage2Response),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":    "tsk_page2_example001",
							"title": "Review notes",
						},
						Raw: map[string]any{
							"id":        "tsk_page2_example001",
							"title":     "Review notes",
							"status":    "Done",
							"meetingId": "mtg_fbff3264957a48b4888",
						},
					},
				},
				Done: true,
			},
		},
		{
			Name: "GraphQL errors in response return error",
			Input: common.ReadParams{
				ObjectName: "tasks",
				Fields:     connectors.Fields("id"),
				PageSize:   2,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
				},
				Then: mockserver.Response(http.StatusOK, readErrorResponse),
			}.Server(),
			ExpectedErrs: []error{
				testutils.StringError("VALIDATION_FAILED: invalid cursor"),
			},
		},
		{
			Name: "Read tasks with case-insensitive nested field includes assignee in query",
			Input: common.ReadParams{
				ObjectName: "tasks",
				Fields:     connectors.Fields("id", "title", "Assignee"),
				PageSize:   2,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Body(string(requestTasks)),
				},
				Then: mockserver.Response(http.StatusOK, tasksResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":    "tsk_9e17843922b94782832",
							"title": "TEST",
						},
						Raw: map[string]any{
							"id":          "tsk_9e17843922b94782832",
							"title":       "TEST",
							"status":      "Todo",
							"meetingId":   "mtg_fbff3264957a48b4888",
							"description": "TEST",
							"assignee": map[string]any{
								"email":  "ada@example.com",
								"name":   "Amperasnd Developer",
								"source": "USER",
							},
						},
					},
					{
						Fields: map[string]any{
							"id":    "tsk_bde1c8b314e8459893e",
							"title": "Send follow-up email",
						},
						Raw: map[string]any{
							"id":        "tsk_bde1c8b314e8459893e",
							"title":     "Send follow-up email",
							"status":    "Todo",
							"meetingId": "mtg_fbff3264957a48b4888",
						},
					},
				},
				NextPage: "g3QAAAACdwJpZG0AAAAXdHNrX2JkZTFjOGIzMTRlODQ1OTg5M2V3C2luc2VydGVkX2F0dAAAAA13C21pY3Jvc2Vjb25kaAJiAAJEgWEGdwZzZWNvbmRhCXcIY2FsZW5kYXJ3E0VsaXhpci5DYWxlbmRhci5JU093BW1vbnRoYQZ3Cl9fc3RydWN0X193D0VsaXhpci5EYXRlVGltZXcKdXRjX29mZnNldGEAdwpzdGRfb2Zmc2V0YQB3BHllYXJiAAAH6ncEaG91cmEXdwNkYXlhF3cJem9uZV9hYmJybQAAAANVVEN3Bm1pbnV0ZWESdwl0aW1lX3pvbmVtAAAAB0V0Yy9VVEM=",
				Done:     false,
			},
		},
		{
			Name: "Successfully read meetings",
			Input: common.ReadParams{
				ObjectName: "meetings",
				Fields:     connectors.Fields("id", "status", "source", "startedAt"),
				PageSize:   2,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Body(string(requestMeetings)),
				},
				Then: mockserver.Response(http.StatusOK, meetingsResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":        "mtg_2df34b22211d4f248f8",
							"status":    "FAILED",
							"source":    "WEB_RECORDER",
							"startedat": "2026-06-27T23:06:44.246638Z",
						},
						Raw: map[string]any{
							"id":        "mtg_2df34b22211d4f248f8",
							"status":    "FAILED",
							"source":    "WEB_RECORDER",
							"startedAt": "2026-06-27T23:06:44.246638Z",
						},
					},
					{
						Fields: map[string]any{
							"id":        "mtg_fbff3264957a48b4888",
							"status":    "COMPLETED",
							"source":    nil,
							"startedat": "2026-06-23T23:18:09.022872Z",
						},
						Raw: map[string]any{
							"id":        "mtg_fbff3264957a48b4888",
							"status":    "COMPLETED",
							"source":    nil,
							"startedAt": "2026-06-23T23:18:09.022872Z",
						},
					},
				},
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

func TestQueryTemplateFields(t *testing.T) {
	t.Parallel()

	fields := queryTemplateFields("tasks", connectors.Fields("id", "Assignee"))

	if !fields["assignee"] {
		t.Fatal("expected canonical assignee key for case-insensitive request")
	}

	if !fields["Assignee"] {
		t.Fatal("expected original requested field key to be preserved")
	}
}
