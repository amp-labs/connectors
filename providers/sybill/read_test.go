package sybill

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

	responseForbidden := testutils.DataFromFile(t, "error-forbidden.json")
	responseConversationsFirstPage := testutils.DataFromFile(t, "conversations-first-page.json")
	responseConversationsEmpty := testutils.DataFromFile(t, "conversations-empty.json")
	responseConversationDetail1 := testutils.DataFromFile(t, "conversation-detail-1.json")
	responseConversationDetail2 := testutils.DataFromFile(t, "conversation-detail-2.json")
	responseAccountsEmpty := testutils.DataFromFile(t, "accounts-empty.json")

	nextCursor := "Q042bXBCc2FCeGlBMXFTUm0yZ2FGa0lVQ0FRU0VCUk9Hb1QwMTFXTXYyZktnbUtwZzlJYUZEb1NSRUZUU0VKUFFWSkVYMDFG" +
		"VkVGRVFWUkJJaFpDRkFnRUVoQVVUaHFFOU5kVmpMOW55b0ppcVlQUw=="

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{Fields: connectors.Fields("conversationId")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "conversations"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.ReadParams{ObjectName: "messages", Fields: connectors.Fields("messageId")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Error detail is parsed",
			Input: common.ReadParams{ObjectName: "deals", Fields: connectors.Fields("dealId")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusForbidden, responseForbidden),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrForbidden,
				testutils.StringError("API key missing required scope: read"),
			},
		},
		{
			Name: "List fields of conversations are read without detail requests",
			Input: common.ReadParams{
				ObjectName: "conversations",
				Fields:     connectors.Fields("conversationId", "title", "startTime"),
				PageSize:   2,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/conversations"),
					mockcond.QueryParam("limit", "2"),
					mockcond.QueryParamsMissing("startedAfter", "startedBefore", "cursor"),
				},
				Then: mockserver.Response(http.StatusOK, responseConversationsFirstPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id: "f119701c-3ff7-5e8c-8e01-2feea07b015b",
					Fields: map[string]any{
						"conversationid": "f119701c-3ff7-5e8c-8e01-2feea07b015b",
						"title":          "Test Co 25 – QBR (connie-sample)",
						"starttime":      "2026-09-26T15:00:00+00:00",
					},
					Raw: map[string]any{"sourceId": "c1dc1ed3-bd87-4d71-8dd1-3da00bacb8b3"},
				}, {
					Id: "144e1a84-f4d7-558c-bf67-ca8262a983d2",
					Fields: map[string]any{
						"conversationid": "144e1a84-f4d7-558c-bf67-ca8262a983d2",
						"title":          "Test Co 24 – Onboarding kickoff (connie-sample)",
						"starttime":      "2026-09-25T15:00:00+00:00",
					},
				}},
				NextPage: common.NextPageToken(nextCursor),
				Done:     false,
			},
		},
		{
			Name: "Transcript and summary come from each conversation's detail",
			Input: common.ReadParams{
				ObjectName: "conversations",
				Fields:     connectors.Fields("conversationId", "category", "summary", "transcript"),
				PageSize:   2,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/v1/conversations"),
						mockcond.QueryParam("limit", "2"),
					},
					Then: mockserver.Response(http.StatusOK, responseConversationsFirstPage),
				}, {
					If:   mockcond.Path("/v1/conversations/f119701c-3ff7-5e8c-8e01-2feea07b015b"),
					Then: mockserver.Response(http.StatusOK, responseConversationDetail1),
				}, {
					If:   mockcond.Path("/v1/conversations/144e1a84-f4d7-558c-bf67-ca8262a983d2"),
					Then: mockserver.Response(http.StatusOK, responseConversationDetail2),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id: "f119701c-3ff7-5e8c-8e01-2feea07b015b",
					Fields: map[string]any{
						"conversationid": "f119701c-3ff7-5e8c-8e01-2feea07b015b",
						"category":       "customer_checkin",
						"summary": map[string]any{
							"Outcome": "Alex Sample 25 identified a follow-up deck and a pricing review as the next steps.",
							"Key Takeaways": []any{
								map[string]any{
									"topic":        "Meeting context",
									"key_takeaway": "Alex Sample 25 opened the meeting as a QBR.",
								},
								map[string]any{
									"topic":        "Connector testing",
									"key_takeaway": "Jordan Demo 25 stated that this was sample data for connector testing.",
								},
								map[string]any{
									"topic": "Follow-up",
									"key_takeaway": "The stated next steps were a follow-up deck and a pricing review; " +
										"no owners or deadlines were specified.",
								},
							},
						},
					},
					Raw: map[string]any{"category": "customer_checkin"},
				}, {
					Id: "144e1a84-f4d7-558c-bf67-ca8262a983d2",
					Fields: map[string]any{
						"category": "customer_onboarding",
					},
					Raw: map[string]any{"category": "customer_onboarding"},
				}},
				NextPage: common.NextPageToken(nextCursor),
				Done:     false,
			},
		},
		{
			Name: "A failed detail request fails the read",
			Input: common.ReadParams{
				ObjectName: "conversations",
				Fields:     connectors.Fields("transcript"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/v1/conversations"),
						mockcond.QueryParam("limit", "10"),
					},
					Then: mockserver.Response(http.StatusOK, responseConversationsFirstPage),
				}, {
					If:   mockcond.Path("/v1/conversations/f119701c-3ff7-5e8c-8e01-2feea07b015b"),
					Then: mockserver.Response(http.StatusForbidden, responseForbidden),
				}},
			}.Server(),
			ExpectedErrs: []error{
				common.ErrForbidden,
				testutils.StringError("API key missing required scope: read"),
			},
		},
		{
			Name: "Incremental read of conversations sends a widened start-time window and the cursor",
			Input: common.ReadParams{
				ObjectName: "conversations",
				Fields:     connectors.Fields("conversationId"),
				Since:      time.Date(2026, 10, 23, 2, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 10, 23, 3, 0, 0, 0, time.UTC),
				NextPage:   "cursor-2",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/conversations"),
					mockcond.QueryParam("limit", "50"),
					mockcond.QueryParam("startedAfter", "2026-10-22T20:00:00Z"),
					mockcond.QueryParam("startedBefore", "2026-10-23T03:00:01Z"),
					mockcond.QueryParam("cursor", "cursor-2"),
				},
				Then: mockserver.Response(http.StatusOK, responseConversationsEmpty),
			}.Server(),
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Accounts are read in full: Since and Until are not sent",
			Input: common.ReadParams{
				ObjectName: "accounts",
				Fields:     connectors.Fields("accountId"),
				Since:      time.Date(2026, 10, 23, 2, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 10, 23, 3, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/accounts"),
					mockcond.QueryParam("limit", "50"),
					mockcond.QueryParamsMissing("startedAfter", "startedBefore",
						"createdAfter", "createdBefore", "lastActivityAfter", "lastActivityBefore"),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountsEmpty),
			}.Server(),
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
				NextPage: "",
				Done:     true,
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
