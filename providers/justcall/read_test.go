package justcall

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

func TestRead(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	responseUsers := testutils.DataFromFile(t, "read/users/list.json")
	responseCallsFirstPage := testutils.DataFromFile(t, "read/calls/first-page.json")
	responseCallsLastPage := testutils.DataFromFile(t, "read/calls/last-page.json")
	responseContacts := testutils.DataFromFile(t, "read/contacts/list.json")
	responseTexts := testutils.DataFromFile(t, "read/texts/list.json")
	responseWebhooks := testutils.DataFromFile(t, "read/webhooks/list.json")
	responseSalesDialerContacts := testutils.DataFromFile(t, "read/sales_dialer_contacts/first-page.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
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
			Name: "Read users",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id", "name", "email"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/users"),
					mockcond.QueryParam("per_page", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseUsers),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":    float64(12345),
							"name":  "John Doe",
							"email": "john@example.com",
						},
						Raw: map[string]any{
							"role":      "Admin",
							"extension": float64(101),
						},
					},
					{
						Fields: map[string]any{
							"id":    float64(12346),
							"name":  "Jane Smith",
							"email": "jane@example.com",
						},
						Raw: map[string]any{
							"role":      "Agent",
							"extension": float64(102),
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read calls first page with pagination",
			Input: common.ReadParams{
				ObjectName: "calls",
				Fields:     connectors.Fields("id", "contact_name", "agent_name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/calls"),
					mockcond.QueryParam("per_page", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseCallsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":           float64(1001),
							"contact_name": "Alice Johnson",
							"agent_name":   "John Doe",
						},
						Raw: map[string]any{
							"call_sid": "CA123456789",
						},
					},
					{
						Fields: map[string]any{
							"id":           float64(1002),
							"contact_name": "Bob Williams",
							"agent_name":   "Jane Smith",
						},
						Raw: map[string]any{
							"call_sid": "CA987654321",
						},
					},
				},
				NextPage: "https://api.justcall.io/v2.1/calls?page=2&per_page=2",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read calls last page using NextPage token",
			Input: common.ReadParams{
				ObjectName: "calls",
				Fields:     connectors.Fields("id", "contact_name"),
				NextPage:   testroutines.URLTestServer + "/v2.1/calls?page=3&per_page=2",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/calls"),
					mockcond.QueryParam("page", "3"),
					mockcond.QueryParam("per_page", "2"),
				},
				Then: mockserver.Response(http.StatusOK, responseCallsLastPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":           float64(1005),
							"contact_name": "Eve Davis",
						},
						Raw: map[string]any{
							"call_sid": "CA555555555",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{ //nolint:dupl
			Name: "Read contacts",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "name", "email", "company"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/contacts"),
					mockcond.QueryParam("per_page", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":      float64(5001),
							"name":    "Alice Johnson",
							"email":   "alice@example.com",
							"company": "Acme Corp",
						},
						Raw: map[string]any{
							"first_name":     "Alice",
							"contact_number": "+14155551111",
						},
					},
					{
						Fields: map[string]any{
							"id":      float64(5002),
							"name":    "Bob Williams",
							"email":   "bob@example.com",
							"company": "Tech Inc",
						},
						Raw: map[string]any{
							"first_name":     "Bob",
							"contact_number": "+14155552222",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{ //nolint:dupl
			Name: "Read texts",
			Input: common.ReadParams{
				ObjectName: "texts",
				Fields:     connectors.Fields("id", "body", "direction", "contact_name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/texts"),
					mockcond.QueryParam("per_page", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseTexts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":           float64(8001),
							"body":         "Hello, this is a test message",
							"direction":    "outbound",
							"contact_name": "Alice Johnson",
						},
						Raw: map[string]any{
							"delivery_status": "delivered",
							"medium":          "sms",
						},
					},
					{
						Fields: map[string]any{
							"id":           float64(8002),
							"body":         "Thanks for your response",
							"direction":    "inbound",
							"contact_name": "Alice Johnson",
						},
						Raw: map[string]any{
							"delivery_status": "delivered",
							"medium":          "sms",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read webhooks without pagination",
			Input: common.ReadParams{
				ObjectName: "webhooks",
				Fields:     connectors.Fields("id", "topic", "url", "status"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/webhooks"),
					mockcond.QueryParamsMissing("per_page"),
				},
				Then: mockserver.Response(http.StatusOK, responseWebhooks),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":     float64(9001),
							"topic":  "call.completed",
							"url":    "https://example.com/webhooks/call-completed",
							"status": "active",
						},
						Raw: map[string]any{
							"created_at": "2024-01-15 10:00:00",
						},
					},
					{
						Fields: map[string]any{
							"id":     float64(9002),
							"topic":  "sms.received",
							"url":    "https://example.com/webhooks/sms-received",
							"status": "active",
						},
						Raw: map[string]any{
							"created_at": "2024-02-20 15:30:00",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read calls with incremental sync (Since/Until)",
			Input: common.ReadParams{
				ObjectName: "calls",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/calls"),
					mockcond.QueryParam("from_datetime", "2024-01-01 00:00:00"),
					mockcond.QueryParam("to_datetime", "2024-12-31 23:59:59"),
				},
				Then: mockserver.Response(http.StatusOK, responseCallsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"id": float64(1001)},
						Raw:    map[string]any{"call_sid": "CA123456789"},
					},
					{
						Fields: map[string]any{"id": float64(1002)},
						Raw:    map[string]any{"call_sid": "CA987654321"},
					},
				},
				NextPage: "https://api.justcall.io/v2.1/calls?page=2&per_page=2",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read returns error on 400 Bad Request",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id", "name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/users"),
					mockcond.QueryParam("per_page", "100"),
				},
				Then: mockserver.Response(http.StatusBadRequest, testutils.DataFromFile(t, "read/error-bad-request.json")),
			}.Server(),
			ExpectedErrs: []error{common.ErrBadRequest},
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

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			Module:              common.ModuleRoot,
			AuthenticatedClient: &http.Client{},
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestBaseURL(serverURL)

	return connector, nil
}
