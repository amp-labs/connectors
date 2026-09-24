package mail

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

	errorForbidden := testutils.DataFromFile(t, "forbidden.json")
	errorNotFound := testutils.DataFromFile(t, "not-found.html")
	responseMessagesFirstPage := testutils.DataFromFile(t, "read/messages/1-first-page.json")
	responseMessagesLastPage := testutils.DataFromFile(t, "read/messages/2-last-page.json")
	responseMessagesList := testutils.DataFromFile(t, "read/messages/list.json")
	responseMessageItem := testutils.DataFromFile(t, "read/messages/message-item.json")
	responseDrafts := testutils.DataFromFile(t, "read/drafts/drafts.json")
	responseDraftMessageItem1 := testutils.DataFromFile(t, "read/drafts/message-item-1.json")
	responseDraftMessageItem2 := testutils.DataFromFile(t, "read/drafts/message-item-2.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "messages"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Error forbidden object",
			Input: common.ReadParams{ObjectName: "messages", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusForbidden, errorForbidden),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrForbidden,
				testutils.StringError("CSE is not enabled."),
			},
		},
		{
			Name:  "HTML Error for not found object",
			Input: common.ReadParams{ObjectName: "messages", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentHTML(),
				Always: mockserver.Response(http.StatusNotFound, errorNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				common.ErrNotFound,
				testutils.StringError("The requested URL /gmail/v1/users/me/butterfly was not found on this server. That’s all we know"), // nolint:lll
			},
		},
		{
			Name: "Read messages first page",
			Input: common.ReadParams{
				ObjectName: "messages",
				Fields:     connectors.Fields("id"),
				PageSize:   33,
				Since:      time.Unix(1726815600, 0),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/gmail/v1/users/me/messages"),
					mockcond.QueryParam("maxResults", "33"), // from params
					mockcond.QueryParam("q", "after:1726815600"),
				},
				Then: mockserver.Response(http.StatusOK, responseMessagesFirstPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": "1993fb4b539b5a1a",
					},
					Raw: map[string]any{
						"threadId": "1993fa50bd191f7b",
					},
				}, {
					Fields: map[string]any{
						"id": "1993fa50bd191f7b",
					},
					Raw: map[string]any{
						"threadId": "1993fa50bd191f7b",
					},
				}},
				NextPage: "08277485409175924556",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read messages second page without next cursor",
			Input: common.ReadParams{
				ObjectName: "messages",
				Fields:     connectors.Fields("id", "$['payload']['body']", "threadId"),
				NextPage:   "08277485409175924556",
				Since:      time.Unix(1727740800, 0),
				Until:      time.Unix(1767830400, 0),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					// Get collection of all messages. This is a list of identifiers.
					If: mockcond.And{
						mockcond.Path("/gmail/v1/users/me/messages"),
						mockcond.QueryParam("maxResults", "500"), // default page size
						mockcond.QueryParam("pageToken", "08277485409175924556"),
						mockcond.QueryParam("q", "after:1727740800 before:1767830400"),
					},
					Then: mockserver.Response(http.StatusOK, responseMessagesLastPage),
				}, {
					// Each message is fetched.
					// The last page is a collection of messages with just 1 message where id=19174f3eeda702ed.
					If:   mockcond.Path("/gmail/v1/users/me/messages/19174f3eeda702ed"),
					Then: mockserver.Response(http.StatusOK, responseMessageItem),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"payload": map[string]any{
							"body": map[string]any{
								"size": float64(0), // nested field was requested
							},
						},
						"id":       "19174f3eeda702ed",
						"threadid": "19174f3eeda702ed", // fields are lower case
					},
					Raw: map[string]any{
						"id":       "19174f3eeda702ed",
						"threadId": "19174f3eeda702ed",
						// Message content is embedded.
						"snippet":      "Restart your 14-day free trial now.",
						"sizeEstimate": float64(29817),
						"historyId":    "31772",
						"internalDate": "1769523000000",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read drafts with embedded messages",
			Input: common.ReadParams{
				ObjectName: "drafts",
				Fields: connectors.Fields(
					"$['message']['snippet']",
					"$['message']['payload']['body']",
					"$['message']['payload']['mimeType']",
				),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.Path("/gmail/v1/users/me/drafts"),
						mockcond.QueryParam("maxResults", "500"),
					},
					Then: mockserver.Response(http.StatusOK, responseDrafts),
				}, {
					If:   mockcond.Path("/gmail/v1/users/me/messages/19c5a57884cc0fc0"),
					Then: mockserver.Response(http.StatusOK, responseDraftMessageItem1),
				}, {
					If:   mockcond.Path("/gmail/v1/users/me/messages/19c5a576fbc9a5ec"),
					Then: mockserver.Response(http.StatusOK, responseDraftMessageItem2),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id: "r5482888295841858934",
					Fields: map[string]any{
						"message": map[string]any{
							"snippet": "(DRAFT) Good news. You are being upgraded from Basic",
							"payload": map[string]any{
								"mimetype": "multipart/alternative",
								"body": map[string]any{
									"size": float64(0),
								},
							},
						},
					},
					Raw: map[string]any{
						"id": "r5482888295841858934",
						"message": map[string]any{
							"id":       "19c5a57884cc0fc0",
							"threadId": "19c2643cde3cab88",
							"labelIds": []any{"DRAFT"},
							"snippet":  "(DRAFT) Good news. You are being upgraded from Basic",
							"payload": map[string]any{
								"partId":   "",
								"mimeType": "multipart/alternative",
								"filename": "",
								"headers": []any{
									map[string]any{
										"name":  "Subject",
										"value": "(DRAFT) Good news",
									},
								},
								"body": map[string]any{
									"size": float64(0),
								},
								"parts": []any{},
							},
							"sizeEstimate": float64(1924),
							"historyId":    "39895",
							"internalDate": "1771042211000",
						},
					},
				}, {
					Id: "r9115380863326367925",
					Fields: map[string]any{
						"message": map[string]any{
							"snippet": "(DRAFT) You have had Linux installed for about a month now",
							"payload": map[string]any{
								"mimetype": "multipart/alternative",
								"body": map[string]any{
									"size": float64(0),
								},
							},
						},
					},
					Raw: map[string]any{
						"id": "r9115380863326367925",
						"message": map[string]any{
							"id":       "19c5a576fbc9a5ec",
							"threadId": "19c5a5623749ebbb",
							"labelIds": []any{"DRAFT"},
							"snippet":  "(DRAFT) You have had Linux installed for about a month now",
							"payload": map[string]any{
								"partId":   "",
								"mimeType": "multipart/alternative",
								"filename": "",
								"headers": []any{
									map[string]any{
										"name":  "Subject",
										"value": "(DRAFT) Linux installed",
									},
								},
								"body": map[string]any{
									"size": float64(0),
								},
								"parts": []any{},
							},
							"sizeEstimate": float64(1677),
							"historyId":    "39887",
							"internalDate": "1771042205000",
						},
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Incremental read of virtual object sentMessages",
			Input: common.ReadParams{
				ObjectName: "AMPERSAND-sentMessages",
				Fields:     connectors.Fields("snippet", "labelIds"),
				Since: time.Date(2024, 9, 19, 4, 30, 45, 600,
					time.FixedZone("UTC-8", -8*60*60)),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.Path("/gmail/v1/users/me/messages"),
						mockcond.QueryParam("maxResults", "500"),
						mockcond.QueryParam("q", "after:1726749045 in:SENT"),
					},
					Then: mockserver.Response(http.StatusOK, responseMessagesList),
				}, {
					If:   mockcond.Path("/gmail/v1/users/me/messages/1a0cae6d02f3bbe3"),
					Then: mockserver.ResponseString(http.StatusOK, `{"id": "1a0cae6d02f3bbe3","labelIds": ["SENT"]}`),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id:     "1a0cae6d02f3bbe3",
					Fields: map[string]any{"labelids": []any{"SENT"}},
					Raw:    map[string]any{"id": "1a0cae6d02f3bbe3"},
				}},
				NextPage: "04710164667041394880",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Incremental read of virtual object inbox messages",
			Input: common.ReadParams{
				ObjectName: "AMPERSAND-messages",
				Fields:     connectors.Fields("snippet", "labelIds"),
				Since: time.Date(2024, 9, 19, 4, 30, 45, 600,
					time.FixedZone("UTC-8", -8*60*60)),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.Path("/gmail/v1/users/me/messages"),
						mockcond.QueryParam("maxResults", "500"),
						mockcond.QueryParam("q", "after:1726749045 in:INBOX"),
					},
					Then: mockserver.Response(http.StatusOK, responseMessagesList),
				}, {
					If:   mockcond.Path("/gmail/v1/users/me/messages/1a0cae6d02f3bbe3"),
					Then: mockserver.ResponseString(http.StatusOK, `{"id": "1a0cae6d02f3bbe3","labelIds": ["INBOX"]}`),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id:     "1a0cae6d02f3bbe3",
					Fields: map[string]any{"labelids": []any{"INBOX"}},
					Raw:    map[string]any{"id": "1a0cae6d02f3bbe3"},
				}},
				NextPage: "04710164667041394880",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestAdapter(tt.Server)
			})
		})
	}
}
