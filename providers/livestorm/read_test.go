package livestorm

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/go-test/deep"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	since := time.Unix(1735689600, 0).UTC()
	until := time.Unix(1735776000, 0).UTC()

	eventsBody := testutils.DataFromFile(t, "read-events.json")
	eventsMultipageBody := testutils.DataFromFile(t, "read-events-multipage.json")
	peopleFirstPageBody := testutils.DataFromFile(t, "read-people-first-page.json")
	jobBody := testutils.DataFromFile(t, "read-job.json")
	peopleAttributesBody := testutils.DataFromFile(t, "read-people-attributes.json")

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
				Then: mockserver.Response(http.StatusOK, eventsBody),
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
							"id":   "evt_1",
							"type": "events",
							"attributes": map[string]any{
								"title":      "Launch Webinar",
								"updated_at": "2025-01-01T12:00:00Z",
							},
						},
					},
				},
			},
		},
		{
			Name: "Read events advances page from meta when next_page absent",
			Input: common.ReadParams{
				ObjectName: objectEvents,
				Fields:     connectors.Fields("title"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/events"),
					mockcond.QueryParam("page[number]", "0"),
					mockcond.QueryParam("page[size]", "100"),
					mockcond.QueryParamsMissing("filter[updated_since]", "filter[updated_until]"),
				},
				Then: mockserver.Response(http.StatusOK, eventsMultipageBody),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows:     1,
				Done:     false,
				NextPage: testroutines.URLTestServer + "/v1/events?page[number]=1&page[size]=100",
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"title": "Paged Event",
						},
						Raw: map[string]any{
							"id":   "evt_1",
							"type": "events",
							"attributes": map[string]any{
								"title":      "Paged Event",
								"updated_at": "2025-01-02T12:00:00Z",
							},
						},
					},
				},
			},
		},
		{
			Name: "Read people first page",
			Input: common.ReadParams{
				ObjectName: "people",
				Fields:     connectors.Fields("email", "first_name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/people"),
					mockcond.QueryParam("page[number]", "0"),
					mockcond.QueryParam("page[size]", "100"),
				},
				Then: mockserver.Response(http.StatusOK, peopleFirstPageBody),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Done: true,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"email":      "alpha@example.com",
							"first_name": "Alpha",
						},
						Raw: map[string]any{
							"id":   "person_a",
							"type": "people",
							"attributes": map[string]any{
								"email":      "alpha@example.com",
								"first_name": "Alpha",
								"last_name":  "User",
							},
						},
					},
				},
			},
		},
		{
			Name: "Read people offset pagination via meta.next_page absolute URL",
			Input: common.ReadParams{
				ObjectName: "people",
				NextPage:   testroutines.URLTestServer + "/v1/people?page[number]=2&page[size]=100",
				Fields:     connectors.Fields("email"),
			},
			Server: mockserver.NewServer(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				// Connector path: readNextPageFromMeta reads meta.next_page and returns it as the next
				// page token (offset-style pagination with an absolute URL). The mock must use r.Host so
				// that URL matches the httptest origin expected after {{testServerURL}} resolution.
				body := fmt.Sprintf(
					`{"data":[{"id":"person_1","type":"people","attributes":{"email":"a@example.com"}}],"meta":{"next_page":"http://%s/v1/people?page[number]=3&page[size]=100"}}`,
					r.Host,
				)

				_, _ = w.Write([]byte(body)) // nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
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
							"id":   "person_1",
							"type": "people",
							"attributes": map[string]any{
								"email": "a@example.com",
							},
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
				Then: mockserver.Response(http.StatusOK, jobBody),
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
							"id":   "job_1",
							"type": "jobs",
							"attributes": map[string]any{
								"status": "done",
							},
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
				Then: mockserver.Response(http.StatusOK, peopleAttributesBody),
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
							"type": "people_attributes",
							"attributes": map[string]any{
								"slug": "company",
							},
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
				Then: mockserver.Response(http.StatusOK, eventsBody),
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

// TestRead_fieldsFlattenedRawPreservesJSONAPI asserts the read contract: Fields holds a flat map
// suitable for field projection (id + attributes merged for extraction only), while Raw is the
// full JSON:API resource object. Full equality (not subset matching) shows the two shapes differ
// and Raw is not flattened to match Fields.
func TestRead_fieldsFlattenedRawPreservesJSONAPI(t *testing.T) {
	t.Parallel()

	body := testutils.DataFromFile(t, "read-people-first-page.json")

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/v1/people"),
			mockcond.QueryParam("page[number]", "0"),
			mockcond.QueryParam("page[size]", "100"),
		},
		Then: mockserver.Response(http.StatusOK, body),
	}.Server()
	t.Cleanup(srv.Close)

	conn, err := constructTestConnector(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	got, err := conn.Read(t.Context(), common.ReadParams{
		ObjectName: "people",
		Fields:     connectors.Fields("email", "first_name"),
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	want := &common.ReadResult{
		Rows:     1,
		Done:     true,
		NextPage: "",
		Data: []common.ReadResultRow{
			{
				Fields: map[string]any{
					"email":      "alpha@example.com",
					"first_name": "Alpha",
				},
				Raw: map[string]any{
					"id":   "person_a",
					"type": "people",
					"attributes": map[string]any{
						"email":      "alpha@example.com",
						"first_name": "Alpha",
						"last_name":  "User",
					},
				},
				Id: "person_a",
			},
		},
	}

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("ReadResult not equal to expected (full match, not subset):\n%s",
			strings.Join(deep.Equal(want, got), "\n"))
	}

	if _, has := got.Data[0].Fields["attributes"]; has {
		t.Fatal("Fields must be flat for projection; must not include JSON:API \"attributes\" object")
	}
}
