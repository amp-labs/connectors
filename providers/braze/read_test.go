package braze

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

	scheduledBroadcasts := []byte(`{
		"scheduled_broadcasts": [
			{"id": "b-1", "name": "Weekly digest", "type": "Campaign", "next_send_time": "2026-12-01T10:00:00Z"}
		]
	}`)

	tests := []testconn.TestCaseRead{
		{
			Name:         "Scheduled broadcasts read without Until is rejected",
			Input:        common.ReadParams{ObjectName: "messages/scheduled_broadcasts", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrMissingUntilTimestamp},
		},
		{
			Name: "Scheduled broadcasts read sends Until as end_time in UTC",
			Input: common.ReadParams{
				ObjectName: "messages/scheduled_broadcasts",
				Fields:     connectors.Fields("id", "name"),
				Until:      time.Date(2026, 12, 31, 2, 0, 0, 0, time.FixedZone("UTC+2", 2*60*60)),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/messages/scheduled_broadcasts"),
					mockcond.QueryParam("end_time", "2026-12-31T00:00:00Z"),
				},
				Then: mockserver.Response(http.StatusOK, scheduledBroadcasts),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": "b-1", "name": "Weekly digest"},
					Raw:    map[string]any{"type": "Campaign"},
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
