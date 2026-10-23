package circleback

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

func TestRead(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	responseRateLimited := testutils.DataFromFile(t, "error-rate-limited.json")
	responseMeetings := testutils.DataFromFile(t, "meetings-first-page.json")
	responseTranscript := testutils.DataFromFile(t, "transcript.json")
	responseActionItems := testutils.DataFromFile(t, "action-items.json")

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
			Name: "Meetings since a time keep only those updated after it, and the cursor is sent back",
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
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id:     "wYpvAY2teOlEXx7pCAhlR",
					Fields: map[string]any{"name": "Test Co 59 – Procurement sync (connie-sample)"},
					Raw:    map[string]any{"updatedAt": "2026-10-22T06:27:42.389Z"},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Selecting transcript fetches it per meeting",
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
				}},
			}.Server(),
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
			Name: "Action items are read for every assignee and status, ignoring Since",
			Input: common.ReadParams{
				ObjectName: "action-items",
				Fields:     connectors.Fields("title", "status"),
				Since:      time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/action-items"),
					mockcond.QueryParam("assigneeType", "Anyone"),
					mockcond.QueryParam("status", "All"),
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
				Done: true,
			},
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
				Then: mockserver.ResponseString(http.StatusOK, `[]`),
			}.Server(),
			Expected: &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
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
