package calendly

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseScheduledEvents := testutils.DataFromFile(t, "scheduled_events.json")
	responseUsersMe := testutils.DataFromFile(t, "users_me.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.Path("/users/me"),
						Then: mockserver.Response(http.StatusOK, responseUsersMe),
					},
				},
				Default: mockserver.Response(http.StatusOK, responseScheduledEvents),
			}.Server(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "scheduled_events"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.Path("/users/me"),
						Then: mockserver.Response(http.StatusOK, responseUsersMe),
					},
				},
				Default: mockserver.Response(http.StatusOK, responseScheduledEvents),
			}.Server(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown objects are not supported",
			Input:        common.ReadParams{ObjectName: "unknown_object", Fields: connectors.Fields("id")},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.Path("/users/me"),
						Then: mockserver.Response(http.StatusOK, responseUsersMe),
					},
				},
				Default: mockserver.Response(http.StatusOK, responseScheduledEvents),
			}.Server(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Read list of scheduled events",
			Input: common.ReadParams{ObjectName: "scheduled_events", Fields: connectors.Fields("")},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.Path("/users/me"),
						Then: mockserver.Response(http.StatusOK, responseUsersMe),
					},
					{
						If:   mockcond.Path("/scheduled_events"),
						Then: mockserver.Response(http.StatusOK, responseScheduledEvents),
					},
				},
				Default: mockserver.Response(http.StatusOK, responseScheduledEvents),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"uri":        "https://api.calendly.com/scheduled_events/GBGBDCAADAEDCRZ2",
						"name":       "15 Minute Meeting",
						"status":     "active",
						"start_time": "2024-01-15T10:00:00.000000Z",
						"end_time":   "2024-01-15T10:15:00.000000Z",
						"event_type": "https://api.calendly.com/event_types/GBGBDCAADAEDCRZ2",
						"created_at": "2024-01-01T09:00:00.000000Z",
						"updated_at": "2024-01-01T09:00:00.000000Z",
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read scheduled events with pagination",
			Input: common.ReadParams{ObjectName: "scheduled_events", Fields: connectors.Fields("")},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.Path("/users/me"),
						Then: mockserver.Response(http.StatusOK, responseUsersMe),
					},
					{
						If:   mockcond.Path("/scheduled_events"),
						Then: mockserver.ResponseString(http.StatusOK, `{
							"collection": [
								{
									"uri": "https://api.calendly.com/scheduled_events/EVENT1",
									"name": "Meeting 1",
									"status": "active",
									"start_time": "2024-01-15T10:00:00.000000Z",
									"end_time": "2024-01-15T10:15:00.000000Z"
								}
							],
							"pagination": {
								"next_page": "https://api.calendly.com/scheduled_events?page_token=next123"
							}
						}`),
					},
				},
				Default: mockserver.Response(http.StatusOK, responseScheduledEvents),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"uri":        "https://api.calendly.com/scheduled_events/EVENT1",
						"name":       "Meeting 1",
						"status":     "active",
						"start_time": "2024-01-15T10:00:00.000000Z",
						"end_time":   "2024-01-15T10:15:00.000000Z",
					},
				}},
				NextPage: "https://api.calendly.com/scheduled_events?page_token=next123",
				Done:     false,
			},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
