package mailgun

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

	responseWebhooks := testutils.DataFromFile(t, "read/webhooks.json")
	responseBounces := testutils.DataFromFile(t, "read/bounces.json")
	responseBouncesEmpty := testutils.DataFromFile(t, "read/bounces-empty.json")
	responseBouncesRelative := testutils.DataFromFile(t, "read/bounces-relative.json")
	responseBouncesWindow := testutils.DataFromFile(t, "read/bounces-window.json")
	responseDomainsPage1 := testutils.DataFromFile(t, "read/domains-page1.json")
	responseDomainsPage2 := testutils.DataFromFile(t, "read/domains-page2.json")
	responseDomainsNull := testutils.DataFromFile(t, "read/domains-null.json")
	responseDynamicPoolsDomains := testutils.DataFromFile(t, "read/dynamic-pools-domains.json")
	responseLists := testutils.DataFromFile(t, "read/lists.json")
	responseListMembers := testutils.DataFromFile(t, "read/list-members.json")
	responseListMembersPage2 := testutils.DataFromFile(t, "read/list-members-page2.json")
	responseLogs := testutils.DataFromFile(t, "read/logs.json")
	responseLogsEmpty := testutils.DataFromFile(t, "read/logs-empty.json")
	responseTagsPage1 := testutils.DataFromFile(t, "read/analytics-tags-page1.json")
	responseTagsPage2 := testutils.DataFromFile(t, "read/analytics-tags-page2.json")
	responseErrorBadRequest := testutils.DataFromFile(t, "read/error-bad-request.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Unsupported object returns operation-not-supported",
			Input: common.ReadParams{
				ObjectName: "nonexistent",
				Fields:     connectors.Fields("id"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Domain-scoped object without workspace errors",
			Input: common.ReadParams{
				ObjectName: "bounces",
				Fields:     connectors.Fields("address"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingDomain},
		},
		{
			// Mailgun surfaces provider errors as {"message": "..."} with a 4xx
			// status; common.InterpretError maps the 400 to ErrCaller (bad-request
			// class) and preserves the message. Mirrors a real live response
			// (ip_pools disabled on the account).
			Name: "Provider 400 maps to caller error with message",
			Input: common.ReadParams{
				ObjectName: "ip_pools",
				Fields:     connectors.Fields("pool_id"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.And{mockcond.MethodGET(), mockcond.Path("/v3/ip_pools")},
					Then: mockserver.Response(http.StatusBadRequest, responseErrorBadRequest),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"message":"unexpected"}`),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrCaller,
				testutils.StringError("this feature is disabled for the account"),
			},
		},
		{
			Name: "Single-shot webhooks returns all rows and is done",
			Input: common.ReadParams{
				ObjectName: "webhooks",
				Fields:     connectors.Fields("webhook_id", "url"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.And{mockcond.MethodGET(), mockcond.Path("/v1/webhooks")},
					Then: mockserver.Response(http.StatusOK, responseWebhooks),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"message":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"webhook_id": "5f1b1e2a3c4d5e6f7a8b9c0d"},
						Raw:    map[string]any{"webhook_id": "5f1b1e2a3c4d5e6f7a8b9c0d"},
					},
					{
						Fields: map[string]any{"webhook_id": "5f1b1e2a3c4d5e6f7a8b9c0e"},
						Raw:    map[string]any{"webhook_id": "5f1b1e2a3c4d5e6f7a8b9c0e"},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Domain-scoped bounces substitutes domain and follows cursor",
			Input: common.ReadParams{
				ObjectName: "bounces",
				Fields:     connectors.Fields("address"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/v3/example.com/bounces"),
						mockcond.QueryParam("limit", "100"),
					},
					Then: mockserver.Response(http.StatusOK, responseBounces),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"message":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"address": "foo@bar.com"},
					Raw:    map[string]any{"address": "foo@bar.com"},
				}},
				NextPage: "https://api.mailgun.net/v3/example.com/bounces?page=next&address=foo%40bar.com&limit=100",
				Done:     false,
			},
		},
		{
			Name: "Cursor terminates on empty page",
			Input: common.ReadParams{
				ObjectName: "bounces",
				Fields:     connectors.Fields("address"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.And{mockcond.MethodGET(), mockcond.Path("/v3/example.com/bounces")},
					Then: mockserver.Response(http.StatusOK, responseBouncesEmpty),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"message":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     0,
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Offset pagination advances skip on a full page",
			Input: common.ReadParams{
				ObjectName: "domains",
				Fields:     connectors.Fields("id", "name"),
				PageSize:   2,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/v4/domains"),
						mockcond.QueryParam("skip", "0"),
						mockcond.QueryParam("limit", "2"),
					},
					Then: mockserver.Response(http.StatusOK, responseDomainsPage1),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"message":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{Fields: map[string]any{"id": "1"}, Raw: map[string]any{"id": "1"}},
					{Fields: map[string]any{"id": "2"}, Raw: map[string]any{"id": "2"}},
				},
				NextPage: "{{testServerURL}}/v4/domains?limit=2&skip=2",
				Done:     false,
			},
		},
		{
			Name: "Offset pagination completes on a short page",
			Input: common.ReadParams{
				ObjectName: "domains",
				Fields:     connectors.Fields("id"),
				PageSize:   2,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.And{mockcond.MethodGET(), mockcond.Path("/v4/domains")},
					Then: mockserver.Response(http.StatusOK, responseDomainsPage2),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"message":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{Fields: map[string]any{"id": "3"}, Raw: map[string]any{"id": "3"}},
				},
				NextPage: "",
				Done:     true,
			},
		},
		{
			// Mailgun returns {"items": null} (not []) for a page past the last
			// record. The read must treat that as an empty, terminal page rather
			// than erroring on the null. Regression test for a live-caught bug.
			Name: "Offset pagination tolerates a null items page",
			Input: common.ReadParams{
				ObjectName: "domains",
				Fields:     connectors.Fields("id"),
				PageSize:   2,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.And{mockcond.MethodGET(), mockcond.Path("/v4/domains")},
					Then: mockserver.Response(http.StatusOK, responseDomainsNull),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"message":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     0,
				NextPage: "",
				Done:     true,
			},
		},
		{
			// Some v1 endpoints return relative paging.next paths; they must be
			// resolved against the connector base URL.
			Name: "Cursor resolves a relative next-page URL",
			Input: common.ReadParams{
				ObjectName: "bounces",
				Fields:     connectors.Fields("address"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.And{mockcond.MethodGET(), mockcond.Path("/v3/example.com/bounces")},
					Then: mockserver.Response(http.StatusOK, responseBouncesRelative),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"message":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				NextPage: "{{testServerURL}}/v3/example.com/bounces?page=next&limit=100",
				Done:     false,
			},
		},
		{
			// The OpenAPI spec declares GET /v1/dynamic_pools/domains with
			// capitalized paging keys ({Next, Previous, First, Last}); the cursor
			// must fall back to paging.Next when paging.next is absent.
			Name: "Cursor follows capitalized paging keys",
			Input: common.ReadParams{
				ObjectName: "dynamic_pools/domains",
				Fields:     connectors.Fields("domain"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.And{mockcond.MethodGET(), mockcond.Path("/v1/dynamic_pools/domains")},
					Then: mockserver.Response(http.StatusOK, responseDynamicPoolsDomains),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"message":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"domain": "pooled.example.com"},
					Raw:    map[string]any{"domain": "pooled.example.com"},
				}},
				NextPage: "https://api.mailgun.net/v1/dynamic_pools/domains?page=abc",
				Done:     false,
			},
		},
		{
			// Mailgun list endpoints accept no time parameters, so Since/Until are
			// applied connector-side against the record timestamp. The next-page
			// cursor comes from the unfiltered page, so filtering does not end
			// pagination early.
			Name: "Since filters records connector-side",
			Input: common.ReadParams{
				ObjectName: "bounces",
				Fields:     connectors.Fields("address"),
				Since:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.And{mockcond.MethodGET(), mockcond.Path("/v3/example.com/bounces")},
					Then: mockserver.Response(http.StatusOK, responseBouncesWindow),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"message":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"address": "new@bar.com"},
					Raw:    map[string]any{"address": "new@bar.com"},
				}},
				NextPage: "https://api.mailgun.net/v3/example.com/bounces?page=next&address=new%40bar.com&limit=100",
				Done:     false,
			},
		},
		{
			// analytics/logs is a POST read. With no Since/Until the body omits
			// start/end so the API's own in-retention defaults apply; the next-page
			// token is read from pagination.next in the response body.
			Name: "Logs POST omits start and end by default and pages by token",
			Input: common.ReadParams{
				ObjectName: "analytics/logs",
				Fields:     connectors.Fields("event"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1/analytics/logs"),
						mockcond.Body(`{"pagination":{"limit":100,"sort":"@timestamp:asc"}}`),
					},
					Then: mockserver.Response(http.StatusOK, responseLogs),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"message":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"event": "delivered"},
						Raw:    map[string]any{"event": "delivered"},
					},
					{
						Fields: map[string]any{"event": "opened"},
						Raw:    map[string]any{"event": "opened"},
					},
				},
				NextPage: "tok-2",
				Done:     false,
			},
		},
		{
			// A follow-up logs page carries Since as RFC 2822 start plus the token
			// in the body; a null items page ends pagination even though the
			// provider echoes another token.
			Name: "Logs next page sends Since and token in the body",
			Input: common.ReadParams{
				ObjectName: "analytics/logs",
				Fields:     connectors.Fields("event"),
				Since:      time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
				NextPage:   "tok-2",
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1/analytics/logs"),
						mockcond.Body(`{
							"start": "Sat, 01 Aug 2026 00:00:00 +0000",
							"pagination": {"limit": 100, "sort": "@timestamp:asc", "token": "tok-2"}
						}`),
					},
					Then: mockserver.Response(http.StatusOK, responseLogsEmpty),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"message":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     0,
				NextPage: "",
				Done:     true,
			},
		},
		{
			// analytics/tags (the modern Tags API) is a POST read whose skip/limit
			// pagination travels in the request body; a full page advances skip via
			// NextPage.
			Name: "Tags POST paginates by body-carried skip",
			Input: common.ReadParams{
				ObjectName: "analytics/tags",
				Fields:     connectors.Fields("tag"),
				PageSize:   2,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1/analytics/tags"),
						mockcond.Body(`{"pagination":{"skip":0,"limit":2}}`),
					},
					Then: mockserver.Response(http.StatusOK, responseTagsPage1),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"message":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{Fields: map[string]any{"tag": "newsletter"}, Raw: map[string]any{"tag": "newsletter"}},
					{Fields: map[string]any{"tag": "onboarding"}, Raw: map[string]any{"tag": "onboarding"}},
				},
				NextPage: "2",
				Done:     false,
			},
		},
		{
			Name: "Tags POST completes on a short page",
			Input: common.ReadParams{
				ObjectName: "analytics/tags",
				Fields:     connectors.Fields("tag"),
				PageSize:   2,
				NextPage:   "2",
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1/analytics/tags"),
						mockcond.Body(`{"pagination":{"skip":2,"limit":2}}`),
					},
					Then: mockserver.Response(http.StatusOK, responseTagsPage2),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"message":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{Fields: map[string]any{"tag": "receipts"}, Raw: map[string]any{"tag": "receipts"}},
				},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Nested lists/members fans out over lists",
			Input: common.ReadParams{
				ObjectName: "lists/members",
				Fields:     connectors.Fields("address", "subscribed"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.And{mockcond.MethodGET(), mockcond.Path("/v3/lists")},
						Then: mockserver.Response(http.StatusOK, responseLists),
					},
					{
						If:   mockcond.And{mockcond.MethodGET(), mockcond.Path("/v3/lists/developers/members")},
						Then: mockserver.Response(http.StatusOK, responseListMembers),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"message":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{Fields: map[string]any{"address": "alice@example.com"}, Raw: map[string]any{"address": "alice@example.com"}},
					{Fields: map[string]any{"address": "bob@example.com"}, Raw: map[string]any{"address": "bob@example.com"}},
				},
				NextPage: "",
				Done:     true,
			},
		},
		{
			// A full member page (records == limit) must trigger a follow-up fetch
			// with an advanced skip; the short second page ends the member walk.
			Name: "Nested lists/members walks every member page",
			Input: common.ReadParams{
				ObjectName: "lists/members",
				Fields:     connectors.Fields("address"),
				PageSize:   2,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.And{mockcond.MethodGET(), mockcond.Path("/v3/lists")},
						Then: mockserver.Response(http.StatusOK, responseLists),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/v3/lists/developers/members"),
							mockcond.QueryParam("skip", "0"),
						},
						Then: mockserver.Response(http.StatusOK, responseListMembers),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/v3/lists/developers/members"),
							mockcond.QueryParam("skip", "2"),
						},
						Then: mockserver.Response(http.StatusOK, responseListMembersPage2),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"message":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{
					{Fields: map[string]any{"address": "alice@example.com"}, Raw: map[string]any{"address": "alice@example.com"}},
					{Fields: map[string]any{"address": "bob@example.com"}, Raw: map[string]any{"address": "bob@example.com"}},
					{Fields: map[string]any{"address": "carol@example.com"}, Raw: map[string]any{"address": "carol@example.com"}},
				},
				NextPage: "",
				Done:     true,
			},
		},
	}

	// Domain-scoped objects require a workspace (Mailgun sending domain).
	workspaceByTest := map[string]string{
		"Domain-scoped bounces substitutes domain and follows cursor": "example.com",
		"Cursor terminates on empty page":                             "example.com",
		"Cursor resolves a relative next-page URL":                    "example.com",
		"Since filters records connector-side":                        "example.com",
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructReadTestConnector(tt.Server.URL, workspaceByTest[tt.Name])
			})
		})
	}
}

func constructReadTestConnector(serverURL, workspace string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: &http.Client{},
		Workspace:           workspace,
	})
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestBaseURL(serverURL)

	return connector, nil
}
