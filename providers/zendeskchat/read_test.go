package zendeskchat

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

	agentsResponse := testutils.DataFromFile(t, "agents-read.json")
	routingSettingsAgentsResponse := testutils.DataFromFile(t, "routing-settings-agents-read.json")

	tests := []testconn.TestCaseRead{
		{
			Name:  "Successfully read agents",
			Input: common.ReadParams{ObjectName: "agents", Fields: connectors.Fields("id", "email")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/v2/chat/agents"),
				Then:  mockserver.Response(http.StatusOK, agentsResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    float64(26363624373267),
						"email": "agent@example.com",
					},
					Raw: map[string]any{
						"display_name": "Test Agent",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully read records keyed by agent id",
			Input: common.ReadParams{ObjectName: "routing_settings/agents", Fields: connectors.Fields("id", "skills")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/v2/chat/routing_settings/agents"),
				Then:  mockserver.Response(http.StatusOK, routingSettingsAgentsResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":     "26363624373267",
						"skills": []any{float64(101)},
					},
					Raw: map[string]any{
						"id":      "26363624373267",
						"routing": map[string]any{"chat_limit": float64(3)},
					},
				}, {
					Fields: map[string]any{
						"id":     "27774298335251",
						"skills": []any{},
					},
					Raw: map[string]any{
						"id":      "27774298335251",
						"routing": map[string]any{"chat_limit": float64(3)},
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Keyed records with empty data yield no records",
			Input: common.ReadParams{ObjectName: "routing_settings/agents", Fields: connectors.Fields("id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/v2/chat/routing_settings/agents"),
				Then:  mockserver.Response(http.StatusOK, []byte(`{"data": {}}`)),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     0,
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
