package mailgun

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	responseListCreate := testutils.DataFromFile(t, "write/list-create.json")
	responseBounceCreate := testutils.DataFromFile(t, "write/bounce-create.json")
	responseRouteUpdate := testutils.DataFromFile(t, "write/route-update.json")
	responseMessageSend := testutils.DataFromFile(t, "write/message-send.json")
	responseForwardCreate := testutils.DataFromFile(t, "write/forward-create.json")
	responseMemberUpsert := testutils.DataFromFile(t, "write/member-upsert.json")
	responseErrorBadRequest := testutils.DataFromFile(t, "read/error-bad-request.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "lists"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Unsupported object returns operation-not-supported",
			Input: common.WriteParams{
				ObjectName: "keys",
				RecordData: map[string]any{"description": "nope"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			// Suppression objects have no update endpoint (create/delete only).
			Name: "Update on a suppression object is rejected",
			Input: common.WriteParams{
				ObjectName: "bounces",
				RecordId:   "bounced@example.com",
				RecordData: map[string]any{"code": "550"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Domain-scoped write without workspace errors",
			Input: common.WriteParams{
				ObjectName: "bounces",
				RecordData: map[string]any{"address": "bounced@example.com"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingDomain},
		},
		{
			// Typical create: POST of a URL-encoded form (Mailgun accepts
			// URL-encoded wherever multipart is documented; verified live).
			Name: "Create list posts a URL-encoded form",
			Input: common.WriteParams{
				ObjectName: "lists",
				RecordData: map[string]any{
					"address": "developers@example.com",
					"name":    "Developers",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/lists"),
					mockcond.Header(http.Header{
						"Content-Type": []string{"application/x-www-form-urlencoded"},
					}),
					mockcond.Body(`address=developers%40example.com&name=Developers`),
				},
				Then: mockserver.Response(http.StatusOK, responseListCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "developers@example.com",
				Data: map[string]any{
					"address": "developers@example.com",
					"name":    "Developers",
				},
			},
		},
		{
			// Domain-scoped create substitutes the workspace into the path; the
			// suppression response is flat (no envelope object).
			Name: "Create bounce substitutes domain and reads flat response",
			Input: common.WriteParams{
				ObjectName: "bounces",
				RecordData: map[string]any{
					"address": "bounced@example.com",
					"code":    "550",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/example.com/bounces"),
					mockcond.Body(`address=bounced%40example.com&code=550`),
				},
				Then: mockserver.Response(http.StatusOK, responseBounceCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "bounced@example.com",
				Data: map[string]any{
					"address": "bounced@example.com",
				},
			},
		},
		{
			// Exception pattern proven live: route CREATE responds with a
			// {message, route: {...}} envelope, but route UPDATE responds flat
			// (route fields at the root beside "message"). The parser falls back
			// to the root when the envelope key is absent.
			Name: "Update route puts to the item path",
			Input: common.WriteParams{
				ObjectName: "routes",
				RecordId:   "4f3bad2335335426750048c6",
				RecordData: map[string]any{"description": "updated inbound route"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v3/routes/4f3bad2335335426750048c6"),
					mockcond.Body(`description=updated+inbound+route`),
				},
				Then: mockserver.Response(http.StatusOK, responseRouteUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "4f3bad2335335426750048c6",
				Data: map[string]any{
					"description": "updated inbound route",
				},
			},
		},
		{
			// messages is the write-only send endpoint; the queued message id is
			// the resulting RecordId.
			Name: "Send message returns the queued message id",
			Input: common.WriteParams{
				ObjectName: "messages",
				RecordData: map[string]any{
					"from":    "sender@example.com",
					"to":      "recipient@example.com",
					"subject": "Hello",
					"text":    "Hi there",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/example.com/messages"),
					mockcond.Body(`from=sender%40example.com&subject=Hello&text=Hi+there&to=recipient%40example.com`),
				},
				Then: mockserver.Response(http.StatusOK, responseMessageSend),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "<20260805171000.1.ABCDEF@example.com>",
				Data: map[string]any{
					"message": "Queued. Thank you.",
				},
			},
		},
		{
			// forwards carries the record as query parameters with an empty body.
			Name: "Create forward sends fields as query parameters",
			Input: common.WriteParams{
				ObjectName: "forwards",
				RecordData: map[string]any{
					"match":       "inbound@example.com",
					"forward.url": "https://example.com/inbound",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/forwards"),
					mockcond.QueryParam("match", "inbound@example.com"),
					mockcond.QueryParam("forward.url", "https://example.com/inbound"),
				},
				Then: mockserver.Response(http.StatusOK, responseForwardCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "6bd6c9dbd8b1c34c1bb1c34c",
				Data: map[string]any{
					"match": "inbound@example.com",
				},
			},
		},
		{
			// lists/members updates POST the collection with Mailgun's upsert flag
			// (its item path would need the parent list as a second identifier);
			// list_address is a sidecar stripped from the payload.
			Name: "Update list member upserts via POST",
			Input: common.WriteParams{
				ObjectName: "lists/members",
				RecordId:   "alice@example.com",
				RecordData: map[string]any{
					"list_address": "developers@example.com",
					"name":         "Alice",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/lists/developers@example.com/members"),
					mockcond.Body(`address=alice%40example.com&name=Alice&upsert=yes`),
				},
				Then: mockserver.Response(http.StatusOK, responseMemberUpsert),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "alice@example.com",
				Data: map[string]any{
					"address": "alice@example.com",
					"name":    "Alice",
				},
			},
		},
		{
			Name: "Provider 400 maps to caller error with message",
			Input: common.WriteParams{
				ObjectName: "lists",
				RecordData: map[string]any{"address": "bad"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.And{mockcond.MethodPOST(), mockcond.Path("/v3/lists")},
				Then:  mockserver.Response(http.StatusBadRequest, responseErrorBadRequest),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrCaller,
				testutils.StringError("this feature is disabled for the account"),
			},
		},
	}

	// Domain-scoped writes require a workspace (Mailgun sending domain).
	workspaceByTest := map[string]string{
		"Create bounce substitutes domain and reads flat response": "example.com",
		"Send message returns the queued message id":               "example.com",
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWriter, error) {
				return constructReadTestConnector(tt.Server.URL, workspaceByTest[tt.Name])
			})
		})
	}
}
