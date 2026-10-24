package circleback

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	responseRateLimited := testutils.DataFromFile(t, "error-rate-limited.json")
	responseMeetings := testutils.DataFromFile(t, "meetings-first-page.json")
	responseTranscript := testutils.DataFromFile(t, "transcript.json")
	responseActionItems := testutils.DataFromFile(t, "action-items.json")
	responsePeople := testutils.DataFromFile(t, "people-first-page.json")
	responseCompanies := testutils.DataFromFile(t, "companies.json")
	responseActionItemsDone := testutils.DataFromFile(t, "action-items-done.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "meetings"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.ReadParams{ObjectName: "transcripts", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Error envelope is parsed",
			Input: common.ReadParams{ObjectName: "tag", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusTooManyRequests, responseRateLimited),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrLimitExceeded,
				testutils.StringError("Ratelimit exceeded for the Free plan. (RATE_LIMITED)"),
			},
		},
		{
			Name: "Meetings are read as every accessible meeting and link to the next cursor",
			Input: common.ReadParams{
				ObjectName: "meetings",
				Fields:     connectors.Fields("name", "updatedAt", "notes"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ResponseChainedFuncs(
					mockserver.ContentJSON(),
					mockserver.Header("Link", `</api/meetings?cursor=eyJwYWdlIjoxfQ&ownership=All>; rel="next"`),
				),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/meetings"),
					mockcond.QueryParam("ownership", "All"),
					mockcond.QueryParamsMissing("cursor"),
				},
				Then: mockserver.Response(http.StatusOK, responseMeetings),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 20,
				Data: []common.ReadResultRow{{
					Id: "IeS3Ok4ymVG3JfD2wYwli",
					Fields: map[string]any{
						"name":      "Test Co 60 – Expansion discussion (connie-sample)",
						"updatedat": "2026-10-22T06:27:43.828Z",
						"notes": "## Summary\nSample expansion discussion with Test Co 60. Fake data seeded " +
							"for connector testing.\n\n## Next steps\n- Send follow-up deck to Alex Sample 60\n" +
							"- Schedule next call",
					},
					Raw: map[string]any{
						"createdAt": "2026-10-31T15:00:00.000Z",
					},
				}},
				NextPage: "eyJwYWdlIjoxfQ",
				Done:     false,
			},
		},
		{
			Name: "Meetings keep those updated within Since and Until, both inclusive, and the cursor is sent back",
			Input: common.ReadParams{
				ObjectName: "meetings",
				Fields:     connectors.Fields("name"),
				Since:      time.Date(2026, 10, 22, 6, 27, 40, 873000000, time.UTC),
				Until:      time.Date(2026, 10, 22, 6, 27, 42, 389000000, time.UTC),
				NextPage:   "eyJwYWdlIjoxfQ",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/meetings"),
					mockcond.QueryParam("ownership", "All"),
					mockcond.QueryParam("cursor", "eyJwYWdlIjoxfQ"),
				},
				Then: mockserver.Response(http.StatusOK, responseMeetings),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id:     "wYpvAY2teOlEXx7pCAhlR",
					Fields: map[string]any{"name": "Test Co 59 – Procurement sync (connie-sample)"},
					Raw:    map[string]any{"updatedAt": "2026-10-22T06:27:42.389Z"},
				}, {
					Id:     "XKfLnBRa58nQK6AocVnpE",
					Fields: map[string]any{"name": "Test Co 58 – Pilot check-in (connie-sample)"},
					Raw:    map[string]any{"updatedAt": "2026-10-22T06:27:40.873Z"},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Selecting transcript fetches it per meeting, and a meeting without one reads as nil",
			Input: common.ReadParams{
				ObjectName: "meetings",
				Fields:     connectors.Fields("transcript"),
				Since:      time.Date(2026, 10, 22, 6, 27, 42, 389000000, time.UTC),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/api/meetings"),
					Then: mockserver.Response(http.StatusOK, responseMeetings),
				}, {
					If:   mockcond.Path("/api/meeting/IeS3Ok4ymVG3JfD2wYwli/transcript"),
					Then: mockserver.Response(http.StatusOK, responseTranscript),
				}, {
					If:   mockcond.Path("/api/meeting/wYpvAY2teOlEXx7pCAhlR/transcript"),
					Then: mockserver.Response(http.StatusNotFound),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id: "IeS3Ok4ymVG3JfD2wYwli",
					Fields: map[string]any{
						"transcript": []any{
							map[string]any{
								"text":      "Thanks for joining the expansion discussion.",
								"speaker":   "Alex Sample 60",
								"timestamp": float64(2),
							},
							map[string]any{
								"text":      "Happy to be here. This is sample data for testing.",
								"speaker":   "Jordan Demo 60",
								"timestamp": 9.5,
							},
							map[string]any{
								"text":      "Great, let's walk through next steps.",
								"speaker":   "Alex Sample 60",
								"timestamp": 17.25,
							},
						},
					},
					Raw: map[string]any{"id": "IeS3Ok4ymVG3JfD2wYwli"},
				}, {
					Id:     "wYpvAY2teOlEXx7pCAhlR",
					Fields: map[string]any{"transcript": nil},
					Raw:    map[string]any{"id": "wYpvAY2teOlEXx7pCAhlR"},
				}},
				Done: true,
			},
		},
		{
			Name: "A transcript request that times out fails the read rather than reading as no transcript",
			Input: common.ReadParams{
				ObjectName: "meetings",
				Fields:     connectors.Fields("transcript"),
				Since:      time.Date(2026, 10, 22, 6, 27, 43, 828000000, time.UTC),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/api/meetings"),
					Then: mockserver.Response(http.StatusOK, responseMeetings),
				}, {
					If:   mockcond.Path("/api/meeting/IeS3Ok4ymVG3JfD2wYwli/transcript"),
					Then: mockserver.Response(http.StatusRequestTimeout),
				}},
			}.Server(),
			ExpectedErrs: []error{common.ErrBadRequest},
		},
		{
			Name: "A rate-limited transcript is retried after its Retry-After",
			Input: common.ReadParams{
				ObjectName: "meetings",
				Fields:     connectors.Fields("transcript"),
				Since:      time.Date(2026, 10, 22, 6, 27, 43, 828000000, time.UTC),
			},
			Server:     rateLimitedOnceServer(responseMeetings, responseRateLimited, responseTranscript),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id: "IeS3Ok4ymVG3JfD2wYwli",
					Fields: map[string]any{
						"transcript": []any{
							map[string]any{
								"text":      "Thanks for joining the expansion discussion.",
								"speaker":   "Alex Sample 60",
								"timestamp": float64(2),
							},
							map[string]any{
								"text":      "Happy to be here. This is sample data for testing.",
								"speaker":   "Jordan Demo 60",
								"timestamp": 9.5,
							},
							map[string]any{
								"text":      "Great, let's walk through next steps.",
								"speaker":   "Alex Sample 60",
								"timestamp": 17.25,
							},
						},
					},
					Raw: map[string]any{"id": "IeS3Ok4ymVG3JfD2wYwli"},
				}},
				Done: true,
			},
		},
		{
			Name: "Action items start with pending items of every assignee, ignoring Since",
			Input: common.ReadParams{
				ObjectName: "action-items",
				Fields:     connectors.Fields("title", "status"),
				Since:      time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ResponseChainedFuncs(
					mockserver.ContentJSON(),
					mockserver.Header("Link",
						`</api/action-items?assigneeType=Anyone&cursor=eyJwYWdlIjoxfQ&status=PENDING>; rel="next"`),
				),
				If: mockcond.And{
					mockcond.Path("/api/action-items"),
					mockcond.QueryParam("assigneeType", "Anyone"),
					mockcond.QueryParam("status", "PENDING"),
					mockcond.QueryParamsMissing("cursor"),
				},
				Then: mockserver.Response(http.StatusOK, responseActionItems),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 25,
				Data: []common.ReadResultRow{{
					Id: "30728947",
					Fields: map[string]any{
						"title":  "Follow up with Test Co 60 (connie-sample)",
						"status": "PENDING",
					},
					Raw: map[string]any{"completedAt": nil},
				}},
				NextPage: "PENDING:eyJwYWdlIjoxfQ",
				Done:     false,
			},
		},
		{
			Name: "The last page of pending action items moves on to done ones",
			Input: common.ReadParams{
				ObjectName: "action-items",
				Fields:     connectors.Fields("title"),
				NextPage:   "PENDING:eyJwYWdlIjoxfQ",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/action-items"),
					mockcond.QueryParam("status", "PENDING"),
					mockcond.QueryParam("cursor", "eyJwYWdlIjoxfQ"),
				},
				Then: mockserver.Response(http.StatusOK, responseActionItems),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 25, NextPage: "DONE:", Done: false},
		},
		{
			Name: "The last page of done action items ends the read",
			Input: common.ReadParams{
				ObjectName: "action-items",
				Fields:     connectors.Fields("title"),
				NextPage:   "DONE:",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/action-items"),
					mockcond.QueryParam("assigneeType", "Anyone"),
					mockcond.QueryParam("status", "DONE"),
					mockcond.QueryParamsMissing("cursor"),
				},
				Then: mockserver.Response(http.StatusOK, responseActionItemsDone),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 0, NextPage: "", Done: true},
		},
		{
			Name:         "An unknown action items page token is rejected",
			Input:        common.ReadParams{ObjectName: "action-items", Fields: connectors.Fields("id"), NextPage: "All:x"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrNextPageInvalid},
		},
		{
			Name: "People send the page size as limit",
			Input: common.ReadParams{
				ObjectName: "people",
				Fields:     connectors.Fields("id"),
				PageSize:   2,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/people"),
					mockcond.QueryParam("limit", "2"),
				},
				Then: mockserver.Response(http.StatusOK, responsePeople),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id:     "63602481",
					Fields: map[string]any{"id": float64(63602481)},
					Raw:    map[string]any{"companyId": float64(8276573)},
				}},
				Done: true,
			},
		},
		{
			Name: "Companies are identified by their domain",
			Input: common.ReadParams{
				ObjectName: "companies",
				Fields:     connectors.Fields("name", "domain"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/companies"),
				Then:  mockserver.Response(http.StatusOK, responseCompanies),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 40,
				Data: []common.ReadResultRow{{
					Id: "testco60.example.com",
					Fields: map[string]any{
						"name":   "Internet Assigned Numbers Authority (IANA)",
						"domain": "testco60.example.com",
					},
					Raw: map[string]any{"domain": "testco60.example.com"},
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

// rateLimitedOnceServer serves the meetings page, and answers the first transcript request with a
// 429 (Retry-After: 0) and every later one with the transcript.
func rateLimitedOnceServer(meetings, rateLimited, transcript []byte) *httptest.Server {
	var transcriptCalls atomic.Int32

	return mockserver.NewServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.URL.Path == "/api/meetings":
			mockserver.Response(http.StatusOK, meetings)(w, r)
		case strings.HasSuffix(r.URL.Path, "/transcript") && transcriptCalls.Add(1) == 1:
			w.Header().Set("Retry-After", "0")
			mockserver.Response(http.StatusTooManyRequests, rateLimited)(w, r)
		default:
			mockserver.Response(http.StatusOK, transcript)(w, r)
		}
	})
}
