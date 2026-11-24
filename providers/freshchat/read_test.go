package freshchat

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestRead(t *testing.T) { //nolint:funlen
	t.Parallel()

	users := []byte(`{
		"users": [
			{"id": "u-1", "first_name": "Ada", "email": "ada@example.com"},
			{"id": "u-2", "first_name": "Linus", "email": "linus@example.com"}
		],
		"pagination": {"total_items": 2, "total_pages": 1, "current_page": 1, "items_per_page": 100}
	}`)

	tests := []testconn.TestCaseRead{
		{
			// Freshchat refuses to list users without a filter, so a read with no time
			// range falls back to created_from at the epoch, sent as a timestamp.
			Name:  "Users read without a time range filters from the epoch",
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields("id", "first_name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/users"),
					mockcond.QueryParam("created_from", "1970-01-01T00:00:00Z"),
					mockcond.QueryParam("items_per_page", "100"),
				},
				Then: mockserver.Response(http.StatusOK, users),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": "u-1", "first_name": "Ada"},
					Raw:    map[string]any{"email": "ada@example.com"},
				}, {
					Fields: map[string]any{"id": "u-2", "first_name": "Linus"},
					Raw:    map[string]any{"email": "linus@example.com"},
				}},
				Done: true,
			},
		},
		{
			// A given time range is the filter, so the epoch fallback is not sent.
			Name: "Users read with Since filters by updated_from alone",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/users"),
					mockcond.QueryParam("updated_from", "2026-01-02T03:04:05Z"),
					mockcond.QueryParamsMissing("created_from"),
				},
				Then: mockserver.Response(http.StatusOK, users),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 2, Done: true},
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

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "test",
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
