package mailgun

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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseDomainsFirst := testutils.DataFromFile(t, "read/domains-first-page.json")
	responseDomainsSecond := testutils.DataFromFile(t, "read/domains-second-page.json")
	responseWebhooks := testutils.DataFromFile(t, "read/webhooks.json")
	responseTemplatesFirst := testutils.DataFromFile(t, "read/templates-first-page.json")
	responseUsersFirst := testutils.DataFromFile(t, "read/users-first-page.json")
	responseUsersSecond := testutils.DataFromFile(t, "read/users-second-page.json")
	responseDynamicPoolsHistory := testutils.DataFromFile(t, "read/dynamic-pools-history.json")
	responseDynamicPoolsHistorySecond := testutils.DataFromFile(t, "read/dynamic-pools-history-second-page.json")
	responseDynamicPoolsDomainsFirst := testutils.DataFromFile(t, "read/dynamic-pools-domains-first-page.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "domains"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.ReadParams{ObjectName: "unknown", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "total_count_skip: read domains first page",
			Input: common.ReadParams{
				ObjectName: "domains",
				Fields:     connectors.Fields("id", "name", "state"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v4/domains"),
					mockcond.QueryParam("limit", "1"),
					mockcond.QueryParam("skip", "0"),
				},
				Then: mockserver.Response(http.StatusOK, responseDomainsFirst),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    "1",
						"name":  "example.com",
						"state": "active",
					},
					Raw: map[string]any{
						"id":         "1",
						"name":       "example.com",
						"created_at": "Mon, 02 Jan 2006 15:04:05 MST",
						"state":      "active",
						"type":       "custom",
					},
				}},
				NextPage: testconn.URLTestServer + "/v4/domains?limit=1&skip=1",
				Done:     false,
			},
		},
		{
			Name: "total_count_skip: read domains second page",
			Input: common.ReadParams{
				ObjectName: "domains",
				Fields:     connectors.Fields("id", "name"),
				NextPage:   testconn.URLTestServer + "/v4/domains?limit=1&skip=1",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v4/domains"),
					mockcond.QueryParam("limit", "1"),
					mockcond.QueryParam("skip", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseDomainsSecond),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "2",
						"name": "other.com",
					},
					Raw: map[string]any{
						"id":         "2",
						"name":       "other.com",
						"created_at": "Tue, 03 Jan 2006 15:04:05 MST",
						"state":      "unverified",
						"type":       "sandbox",
					},
				}},
				Done: true,
			},
		},
		{
			Name:  "none: read webhooks without pagination",
			Input: common.ReadParams{ObjectName: "webhooks", Fields: connectors.Fields("webhook_id", "url")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/webhooks"),
				},
				Then: mockserver.Response(http.StatusOK, responseWebhooks),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"webhook_id": "507f1f77bcf86cd799439011",
						"url":        "https://api.example.com/webhooks/mailgun",
					},
					Raw: map[string]any{
						"webhook_id":  "507f1f77bcf86cd799439011",
						"description": "Production alerts webhook",
						"url":         "https://api.example.com/webhooks/mailgun",
						"event_types": []any{"delivered", "permanent_fail"},
						"created_at":  "2026-02-16T19:06:01Z",
					},
				}},
				Done: true,
			},
		},
		{
			Name: "paging_next: read templates first page",
			Input: common.ReadParams{
				ObjectName: "templates",
				Fields:     connectors.Fields("id", "name", "description"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v4/templates"),
					mockcond.QueryParam("limit", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseTemplatesFirst),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":          "019f6f77-2ecc-7677-9408-f3018a86915c",
						"name":        "ampersand",
						"description": "test",
					},
					Raw: map[string]any{
						"name":        "ampersand",
						"description": "test",
						"createdAt":   "Fri, 17 Jul 2026 09:45:09 UTC",
						"createdBy":   "html",
						"id":          "019f6f77-2ecc-7677-9408-f3018a86915c",
					},
				}},
				NextPage: testconn.URLTestServer + "/v4/templates?page=next&p=ampersand&limit=1",
				Done:     false,
			},
		},
		{
			Name: "total_skip: read users first page",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id", "name", "email"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v5/users"),
					mockcond.QueryParam("limit", "1"),
					mockcond.QueryParam("skip", "0"),
				},
				Then: mockserver.Response(http.StatusOK, responseUsersFirst),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    "123",
						"name":  "John Doe",
						"email": "johndoe@example.com",
					},
					Raw: map[string]any{
						"id":        "123",
						"name":      "John Doe",
						"email":     "johndoe@example.com",
						"role":      "basic",
						"activated": true,
					},
				}},
				NextPage: testconn.URLTestServer + "/v5/users?limit=1&skip=1",
				Done:     false,
			},
		},
		{
			Name: "total_skip: read users second page",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id", "name"),
				NextPage:   testconn.URLTestServer + "/v5/users?limit=1&skip=1",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v5/users"),
					mockcond.QueryParam("limit", "1"),
					mockcond.QueryParam("skip", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseUsersSecond),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "456",
						"name": "Jane Doe",
					},
					Raw: map[string]any{
						"id":        "456",
						"name":      "Jane Doe",
						"email":     "janedoe@example.com",
						"role":      "admin",
						"activated": true,
					},
				}},
				Done: true,
			},
		},
		{
			Name: "paging_next: read dynamic pools history first page (capital Limit + Next)",
			Input: common.ReadParams{
				ObjectName: "dynamic_pools/history",
				Fields:     connectors.Fields("id", "domain_name", "timestamp"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/dynamic_pools/history"),
					mockcond.QueryParam("Limit", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseDynamicPoolsHistory),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":          "hist-1",
						"domain_name": "example.com",
						"timestamp":   "2026-01-15T10:00:00Z",
					},
					Raw: map[string]any{
						"id":          "hist-1",
						"timestamp":   "2026-01-15T10:00:00Z",
						"domain_name": "example.com",
						"reason":      "band_change",
						"account_id":  "acc-1",
					},
				}},
				NextPage: testconn.URLTestServer + "/v1/dynamic_pools/history?Limit=1&page=next",
				Done:     false,
			},
		},
		{
			Name: "paging_next: read dynamic pools history second page",
			Input: common.ReadParams{
				ObjectName: "dynamic_pools/history",
				Fields:     connectors.Fields("id", "domain_name"),
				NextPage:   testconn.URLTestServer + "/v1/dynamic_pools/history?Limit=1&page=next",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/dynamic_pools/history"),
					mockcond.QueryParam("Limit", "1"),
					mockcond.QueryParam("page", "next"),
				},
				Then: mockserver.Response(http.StatusOK, responseDynamicPoolsHistorySecond),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":          "hist-2",
						"domain_name": "other.com",
					},
					Raw: map[string]any{
						"id":          "hist-2",
						"timestamp":   "2026-01-16T10:00:00Z",
						"domain_name": "other.com",
						"reason":      "band_change",
						"account_id":  "acc-1",
					},
				}},
				Done: true,
			},
		},
		{
			Name: "paging_next: read dynamic pools domains first page",
			Input: common.ReadParams{
				ObjectName: "dynamic_pools/domains",
				Fields:     connectors.Fields("id", "name", "pool"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/dynamic_pools/domains"),
					mockcond.QueryParam("limit", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseDynamicPoolsDomainsFirst),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "dom-1",
						"name": "example.com",
						"pool": "dynamic_good",
					},
					Raw: map[string]any{
						"id":         "dom-1",
						"account_id": "acc-1",
						"name":       "example.com",
						"pool":       "dynamic_good",
					},
				}},
				NextPage: testconn.URLTestServer + "/v1/dynamic_pools/domains?limit=1&page=next",
				Done:     false,
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
