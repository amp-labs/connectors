package outreach

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

type GetRecordsByIdsInput struct {
	ObjectName   string
	Ids          []string
	Fields       []string
	Associations []string
}

func TestGetRecordsByIds(t *testing.T) {
	t.Parallel()

	accountsByIdsResponse := testutils.DataFromFile(t, "accounts_by_ids.json")

	tests := []testconn.TestCase[GetRecordsByIdsInput, []common.ReadResultRow]{
		{
			Name:         "Missing object name returns error",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Successfully fetch accounts by IDs",
			Input: GetRecordsByIdsInput{
				ObjectName: "accounts",
				Fields:     []string{"name"},
				Ids:        []string{"1", "2", "5"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/accounts"),
					mockcond.QueryParam("filter[id]", "1,2,5"),
				},
				Then: mockserver.Response(http.StatusOK, accountsByIdsResponse),
			}.Server(),
			Expected: []common.ReadResultRow{
				{
					Id: "1",
					Fields: map[string]any{
						"id":   float64(1),
						"name": "Example Corp 1",
					},
					// Raw preserves the original JSON structure: "attributes" key is NOT removed.
					// Fields are extracted from the flattened record (attributes merged to top level).
					Raw: map[string]any{
						"id":   float64(1),
						"type": "account",
						"attributes": map[string]any{
							"buyerIntentScore":  nil,
							"companyType":       nil,
							"createdAt":         "2025-01-15T10:30:00.000Z",
							"custom1":           nil,
							"customId":          nil,
							"description":       "First test account",
							"domain":            "example1.com",
							"externalSource":    "outreach-api",
							"followers":         float64(0),
							"foundedAt":         nil,
							"industry":          "Technology",
							"linkedInEmployees": nil,
							"linkedInUrl":       nil,
							"locality":          "San Francisco",
							"name":              "Example Corp 1",
							"named":             true,
							"naturalName":       "Example Corp 1",
							"numberOfEmployees": float64(100),
							"tags":              []any{},
							"touchedAt":         "2025-01-15T10:30:00.000Z",
							"updatedAt":         "2025-01-15T10:30:00.000Z",
							"websiteUrl":        "https://example1.com",
						},
						"relationships": map[string]any{
							"owner": map[string]any{
								"data": map[string]any{
									"type": "user",
									"id":   float64(1),
								},
							},
						},
						"links": map[string]any{
							"self": "https://api.outreach.io/api/v2/accounts/1",
						},
					},
				},
				{
					Id: "2",
					Fields: map[string]any{
						"id":   float64(2),
						"name": "Example Corp 2",
					},
					Raw: map[string]any{
						"id":   float64(2),
						"type": "account",
						"attributes": map[string]any{
							"buyerIntentScore":  nil,
							"companyType":       nil,
							"createdAt":         "2025-01-16T11:45:00.000Z",
							"custom1":           nil,
							"customId":          nil,
							"description":       "Second test account",
							"domain":            "example2.com",
							"externalSource":    "outreach-api",
							"followers":         float64(0),
							"foundedAt":         nil,
							"industry":          "Finance",
							"linkedInEmployees": nil,
							"linkedInUrl":       nil,
							"locality":          "New York",
							"name":              "Example Corp 2",
							"named":             true,
							"naturalName":       "Example Corp 2",
							"numberOfEmployees": float64(250),
							"tags":              []any{},
							"touchedAt":         "2025-01-16T11:45:00.000Z",
							"updatedAt":         "2025-01-16T11:45:00.000Z",
							"websiteUrl":        "https://example2.com",
						},
						"relationships": map[string]any{
							"owner": map[string]any{
								"data": map[string]any{
									"type": "user",
									"id":   float64(2),
								},
							},
						},
						"links": map[string]any{
							"self": "https://api.outreach.io/api/v2/accounts/2",
						},
					},
				},
				{
					Id: "5",
					Fields: map[string]any{
						"id":   float64(5),
						"name": "Example Corp 5",
					},
					Raw: map[string]any{
						"id":   float64(5),
						"type": "account",
						"attributes": map[string]any{
							"buyerIntentScore":  nil,
							"companyType":       nil,
							"createdAt":         "2025-01-17T09:20:00.000Z",
							"custom1":           nil,
							"customId":          nil,
							"description":       "Third test account",
							"domain":            "example5.com",
							"externalSource":    "outreach-api",
							"followers":         float64(0),
							"foundedAt":         nil,
							"industry":          "Healthcare",
							"linkedInEmployees": nil,
							"linkedInUrl":       nil,
							"locality":          "Boston",
							"name":              "Example Corp 5",
							"named":             true,
							"naturalName":       "Example Corp 5",
							"numberOfEmployees": float64(500),
							"tags":              []any{},
							"touchedAt":         "2025-01-17T09:20:00.000Z",
							"updatedAt":         "2025-01-17T09:20:00.000Z",
							"websiteUrl":        "https://example5.com",
						},
						"relationships": map[string]any{
							"owner": map[string]any{
								"data": map[string]any{
									"type": "user",
									"id":   float64(1),
								},
							},
						},
						"links": map[string]any{
							"self": "https://api.outreach.io/api/v2/accounts/5",
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully fetch accounts by IDs with associations populates Id",
			Input: GetRecordsByIdsInput{
				ObjectName:   "accounts",
				Fields:       []string{"name"},
				Ids:          []string{"1", "2", "5"},
				Associations: []string{"owner"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/accounts"),
					mockcond.QueryParam("filter[id]", "1,2,5"),
					mockcond.QueryParam("include", "owner"),
				},
				Then: mockserver.Response(http.StatusOK, accountsByIdsResponse),
			}.Server(),
			Expected: []common.ReadResultRow{
				{
					Id: "1",
					Fields: map[string]any{
						"name": "Example Corp 1",
					},
					Raw: map[string]any{
						"id":   float64(1),
						"type": "account",
						"attributes": map[string]any{
							"buyerIntentScore":  nil,
							"companyType":       nil,
							"createdAt":         "2025-01-15T10:30:00.000Z",
							"custom1":           nil,
							"customId":          nil,
							"description":       "First test account",
							"domain":            "example1.com",
							"externalSource":    "outreach-api",
							"followers":         float64(0),
							"foundedAt":         nil,
							"industry":          "Technology",
							"linkedInEmployees": nil,
							"linkedInUrl":       nil,
							"locality":          "San Francisco",
							"name":              "Example Corp 1",
							"named":             true,
							"naturalName":       "Example Corp 1",
							"numberOfEmployees": float64(100),
							"tags":              []any{},
							"touchedAt":         "2025-01-15T10:30:00.000Z",
							"updatedAt":         "2025-01-15T10:30:00.000Z",
							"websiteUrl":        "https://example1.com",
						},
						"relationships": map[string]any{
							"owner": map[string]any{
								"data": map[string]any{
									"type": "user",
									"id":   float64(1),
								},
							},
						},
						"links": map[string]any{
							"self": "https://api.outreach.io/api/v2/accounts/1",
						},
					},
					Associations: map[string][]common.Association{},
				},
				{
					Id: "2",
					Fields: map[string]any{
						"name": "Example Corp 2",
					},
					Raw: map[string]any{
						"id":   float64(2),
						"type": "account",
						"attributes": map[string]any{
							"buyerIntentScore":  nil,
							"companyType":       nil,
							"createdAt":         "2025-01-16T11:45:00.000Z",
							"custom1":           nil,
							"customId":          nil,
							"description":       "Second test account",
							"domain":            "example2.com",
							"externalSource":    "outreach-api",
							"followers":         float64(0),
							"foundedAt":         nil,
							"industry":          "Finance",
							"linkedInEmployees": nil,
							"linkedInUrl":       nil,
							"locality":          "New York",
							"name":              "Example Corp 2",
							"named":             true,
							"naturalName":       "Example Corp 2",
							"numberOfEmployees": float64(250),
							"tags":              []any{},
							"touchedAt":         "2025-01-16T11:45:00.000Z",
							"updatedAt":         "2025-01-16T11:45:00.000Z",
							"websiteUrl":        "https://example2.com",
						},
						"relationships": map[string]any{
							"owner": map[string]any{
								"data": map[string]any{
									"type": "user",
									"id":   float64(2),
								},
							},
						},
						"links": map[string]any{
							"self": "https://api.outreach.io/api/v2/accounts/2",
						},
					},
					Associations: map[string][]common.Association{},
				},
				{
					Id: "5",
					Fields: map[string]any{
						"name": "Example Corp 5",
					},
					Raw: map[string]any{
						"id":   float64(5),
						"type": "account",
						"attributes": map[string]any{
							"buyerIntentScore":  nil,
							"companyType":       nil,
							"createdAt":         "2025-01-17T09:20:00.000Z",
							"custom1":           nil,
							"customId":          nil,
							"description":       "Third test account",
							"domain":            "example5.com",
							"externalSource":    "outreach-api",
							"followers":         float64(0),
							"foundedAt":         nil,
							"industry":          "Healthcare",
							"linkedInEmployees": nil,
							"linkedInUrl":       nil,
							"locality":          "Boston",
							"name":              "Example Corp 5",
							"named":             true,
							"naturalName":       "Example Corp 5",
							"numberOfEmployees": float64(500),
							"tags":              []any{},
							"touchedAt":         "2025-01-17T09:20:00.000Z",
							"updatedAt":         "2025-01-17T09:20:00.000Z",
							"websiteUrl":        "https://example5.com",
						},
						"relationships": map[string]any{
							"owner": map[string]any{
								"data": map[string]any{
									"type": "user",
									"id":   float64(1),
								},
							},
						},
						"links": map[string]any{
							"self": "https://api.outreach.io/api/v2/accounts/5",
						},
					},
					Associations: map[string][]common.Association{},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			t.Cleanup(func() {
				tt.Close()
			})

			conn, err := constructTestConnector(tt.Server.URL)
			if err != nil {
				t.Fatalf("failed to construct test connector: %v", err)
			}

			result, err := conn.GetRecordsByIds(t.Context(), tt.Input.ObjectName, tt.Input.Ids, tt.Input.Fields, tt.Input.Associations)

			var rows []common.ReadResultRow
			if result != nil {
				rows = result.Rows
			}

			tt.Validate(t, err, rows)

			// Every requested id is in the response, so none may be reported missing.
			if err == nil {
				testconn.RequireNoFailures(t, result)
			}
		})
	}
}

// Ids the batch response leaves out are reported as not-found failures, in requested order.
func TestGetRecordsByIds_ReportsMissingIds(t *testing.T) {
	t.Parallel()

	response := testutils.DataFromFile(t, "accounts_by_ids.json")

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.Path("/api/v2/accounts"),
		Then:  mockserver.Response(http.StatusOK, response),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	// The response holds "1", "2", "5"; the other ids are absent from it.
	result, err := conn.GetRecordsByIds(t.Context(), "accounts",
		[]string{"3", "1", "2", "4", "5"}, []string{"name"}, nil)
	if err != nil {
		t.Fatalf("a partial batch response must not fail the call: %v", err)
	}

	testconn.RequireRowIds(t, result, "1", "2", "5")
	testconn.RequireMissingIds(t, result, "3", "4")
}

// numberedIds returns the ids "1" to "n".
func numberedIds(n int) []string {
	ids := make([]string, n)
	for i := range ids {
		ids[i] = strconv.Itoa(i + 1)
	}

	return ids
}

// accountsPage renders an Outreach collection response with one account per id and an optional next link.
func accountsPage(ids []string, next string) string {
	records := make([]string, len(ids))
	for i, id := range ids {
		records[i] = fmt.Sprintf(`{"type":"account","id":%s,"attributes":{"name":"Account %s"}}`, id, id)
	}

	links := "{}"
	if next != "" {
		links = fmt.Sprintf(`{"next":%q}`, next)
	}

	return fmt.Sprintf(`{"data":[%s],"links":%s}`, strings.Join(records, ","), links)
}

// accountsServer answers filter[id] requests with one account per requested id, returning at most page[size]
// accounts, or 50 when page[size] is not set, as Outreach does. Each request's filter[id] values are recorded.
type accountsServer struct {
	mu       sync.Mutex
	requests [][]string
	// failWhen makes the server answer 500 to a request whose filter[id] includes this id.
	failWhen string
}

func (s *accountsServer) handle(w http.ResponseWriter, r *http.Request) {
	ids := strings.Split(r.URL.Query().Get("filter[id]"), ",")

	s.mu.Lock()
	s.requests = append(s.requests, ids)
	s.mu.Unlock()

	if s.failWhen != "" && slices.Contains(ids, s.failWhen) {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	pageSize := 50
	if size, err := strconv.Atoi(r.URL.Query().Get("page[size]")); err == nil {
		pageSize = size
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(accountsPage(ids[:min(pageSize, len(ids))], "")))
}

func getAccountsByIds(t *testing.T, handler http.HandlerFunc, ids []string) (*common.BatchReadResult, error) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	return conn.GetRecordsByIds(t.Context(), "accounts", ids, []string{"name"}, nil)
}

// Outreach returns 50 records per page by default, so more than 50 ids must still all come back.
func TestGetRecordsByIds_ReturnsMoreThanDefaultPageSize(t *testing.T) {
	t.Parallel()

	ids := numberedIds(120)
	server := &accountsServer{}

	result, err := getAccountsByIds(t, server.handle, ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testconn.RequireRowIds(t, result, ids...)
	testconn.RequireNoFailures(t, result)
}

// Ids beyond one chunk are sent in several requests, each with page[size] covering the whole chunk.
func TestGetRecordsByIds_SplitsIdsIntoChunks(t *testing.T) {
	t.Parallel()

	ids := numberedIds(2*readByIdsChunkSize + 50)
	server := &accountsServer{}

	result, err := getAccountsByIds(t, server.handle, ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testconn.RequireRowIds(t, result, ids...)
	testconn.RequireNoFailures(t, result)

	wantSizes := []int{readByIdsChunkSize, readByIdsChunkSize, 50}
	if len(server.requests) != len(wantSizes) {
		t.Fatalf("got %d requests, want %d", len(server.requests), len(wantSizes))
	}

	for i, requested := range server.requests {
		if len(requested) != wantSizes[i] {
			t.Errorf("request %d has %d ids, want %d", i, len(requested), wantSizes[i])
		}
	}
}

// When a response has links.next, the next page is read and its records are returned.
func TestGetRecordsByIds_FollowsNextLink(t *testing.T) {
	t.Parallel()

	var serverURL string

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Query().Get("page[after]") == "" {
			_, _ = w.Write([]byte(accountsPage([]string{"1"}, serverURL+"/api/v2/accounts?page[after]=cursor1")))

			return
		}

		_, _ = w.Write([]byte(accountsPage([]string{"2", "3"}, "")))
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(server.Close)

	serverURL = server.URL

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "accounts", []string{"1", "2", "3"}, []string{"name"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testconn.RequireRowIds(t, result, "1", "2", "3")
	testconn.RequireNoFailures(t, result)
}

// A failed chunk reports its ids with the chunk's error, keeps the rows of the other chunks,
// and returns the error alongside the result.
func TestGetRecordsByIds_ReportsFailedChunk(t *testing.T) {
	t.Parallel()

	ids := numberedIds(readByIdsChunkSize + 50)
	server := &accountsServer{failWhen: strconv.Itoa(readByIdsChunkSize + 1)}

	result, err := getAccountsByIds(t, server.handle, ids)
	if !errors.Is(err, common.ErrServer) {
		t.Fatalf("error = %v, want one wrapping common.ErrServer", err)
	}

	testconn.RequireRowIds(t, result, ids[:readByIdsChunkSize]...)

	failedIds := ids[readByIdsChunkSize:]
	if len(result.Failures) != len(failedIds) {
		t.Fatalf("got %d failures, want %d", len(result.Failures), len(failedIds))
	}

	for i, failure := range result.Failures {
		if failure.RecordId != failedIds[i] || failure.Reason != common.FailureReasonTransient {
			t.Errorf("failure %d = {%s %s}, want {%s %s}", i,
				failure.RecordId, failure.Reason, failedIds[i], common.FailureReasonTransient)
		}
	}
}

// No ids means no request: without a filter[id], Outreach would return unrelated records.
func TestGetRecordsByIds_NoIdsSendsNoRequest(t *testing.T) {
	t.Parallel()

	server := &accountsServer{}

	result, err := getAccountsByIds(t, server.handle, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testconn.RequireRowIds(t, result)
	testconn.RequireNoFailures(t, result)

	if len(server.requests) != 0 {
		t.Errorf("got %d requests, want none", len(server.requests))
	}
}
