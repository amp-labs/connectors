package livestorm

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestRead(t *testing.T) { //nolint:funlen
	t.Parallel()

	since := time.Unix(1735689600, 0).UTC()
	until := time.Unix(1735776000, 0).UTC()

	eventsResponse := []byte(`{
		"data": [
			{
				"id": "evt_1",
				"type": "events",
				"attributes": {
					"title": "Launch Webinar",
					"updated_at": "2025-01-01T12:00:00Z"
				}
			}
		],
		"meta": {
			"current_page": 0,
			"page_count": 2
		}
	}`)

	chatMessagesResponse := []byte(`{
		"data": [
			{
				"id": "chat_1",
				"type": "session_chat_messages",
				"attributes": {
					"text": "hello"
				}
			}
		],
		"meta": {
			"current_page": 0,
			"page_count": 1
		}
	}`)

	jobResponse := []byte(`{
		"data": {
			"id": "job_1",
			"type": "jobs",
			"attributes": {
				"status": "done"
			}
		}
	}`)

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: objectEvents},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown object name is not supported",
			Input:        common.ReadParams{ObjectName: "unknown", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Session chat messages require session id in filter",
			Input:        common.ReadParams{ObjectName: objectSessionChatMessages, Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrSessionIDRequired},
		},
		{
			Name:         "Jobs require job id in filter",
			Input:        common.ReadParams{ObjectName: objectJobs, Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrJobIDRequired},
		},
		{
			Name: "Read events applies v1 path and incremental time filters",
			Input: common.ReadParams{
				ObjectName: objectEvents,
				Fields:     connectors.Fields("title"),
				Since:      since,
				Until:      until,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/events"),
					mockcond.QueryParam("page[number]", "0"),
					mockcond.QueryParam("page[size]", "100"),
					mockcond.QueryParam("filter[updated_since]", fmt.Sprintf("%d", since.Unix())),
					mockcond.QueryParam("filter[updated_until]", fmt.Sprintf("%d", until.Unix())),
				},
				Then: mockserver.Response(http.StatusOK, eventsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows:     1,
				Done:     false,
				NextPage: testroutines.URLTestServer + "/v1/events?page[number]=1&page[size]=100&filter[updated_since]=1735689600&filter[updated_until]=1735776000",
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"title": "Launch Webinar",
						},
						Raw: map[string]any{
							"id":         "evt_1",
							"title":      "Launch Webinar",
							"updated_at": "2025-01-01T12:00:00Z",
						},
					},
				},
			},
		},
		{
			Name: "Read people uses next_page token from response",
			Input: common.ReadParams{
				ObjectName: "people",
				NextPage:   testroutines.URLTestServer + "/v1/people?page[number]=2&page[size]=100",
				Fields:     connectors.Fields("email"),
			},
			Server: mockserver.NewServer(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				body := fmt.Sprintf(
					`{"data":[{"id":"person_1","type":"people","attributes":{"email":"a@example.com"}}],"meta":{"next_page":"http://%s/v1/people?page[number]=3&page[size]=100"}}`,
					r.Host,
				)

				_, _ = w.Write([]byte(body)) // nolint:errcheck
			}),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows:     1,
				Done:     false,
				NextPage: testroutines.URLTestServer + "/v1/people?page[number]=3&page[size]=100",
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"email": "a@example.com",
						},
						Raw: map[string]any{
							"id":    "person_1",
							"email": "a@example.com",
						},
					},
				},
			},
		},
		{
			Name: "Read session chat messages by session id",
			Input: common.ReadParams{
				ObjectName: objectSessionChatMessages,
				Filter:     "sess_1",
				Fields:     connectors.Fields("text"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/sessions/sess_1/chat_messages"),
					mockcond.QueryParam("page[number]", "0"),
					mockcond.QueryParam("page[size]", "100"),
				},
				Then: mockserver.Response(http.StatusOK, chatMessagesResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Done: true,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"text": "hello",
						},
						Raw: map[string]any{
							"id":   "chat_1",
							"text": "hello",
						},
					},
				},
			},
		},
		{
			Name: "Read job by id",
			Input: common.ReadParams{
				ObjectName: objectJobs,
				Filter:     "job_1",
				Fields:     connectors.Fields("status"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/jobs/job_1"),
				},
				Then: mockserver.Response(http.StatusOK, jobResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Done: true,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"status": "done",
						},
						Raw: map[string]any{
							"id":     "job_1",
							"status": "done",
						},
					},
				},
			},
		},
		{
			Name: "Read people attributes via generic object metadata path",
			Input: common.ReadParams{
				ObjectName: "people_attributes",
				Fields:     connectors.Fields("slug"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/people_attributes"),
					mockcond.QueryParam("page[number]", "0"),
					mockcond.QueryParam("page[size]", "100"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{
					"data": [{
						"id": "attr_1",
						"type": "people_attributes",
						"attributes": {
							"slug": "company"
						}
					}],
					"meta": {
						"current_page": 0,
						"page_count": 1
					}
				}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Done: true,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"slug": "company",
						},
						Raw: map[string]any{
							"id":   "attr_1",
							"slug": "company",
						},
					},
				},
			},
		},
		{
			Name: "Read events from NextPage token uses provided URL as-is",
			Input: common.ReadParams{
				ObjectName: objectEvents,
				NextPage:   testroutines.URLTestServer + "/v1/events?page[number]=5&page[size]=100",
				Fields:     connectors.Fields("title"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/events"),
					mockcond.QueryParam("page[number]", "5"),
					mockcond.QueryParam("page[size]", "100"),
				},
				Then: mockserver.Response(http.StatusOK, eventsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				Done:     false,
				NextPage: testroutines.URLTestServer + "/v1/events?page[number]=1&page[size]=100",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
