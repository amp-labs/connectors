package attio

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// GetRecordsByIdsInput represents the input parameters for GetRecordsByIds method.
type GetRecordsByIdsInput struct {
	ObjectName   string
	Ids          []string
	Fields       []string
	Associations []string
}

func TestGetRecordByIds(t *testing.T) {
	t.Parallel()

	responseGetRecordsByIds := testutils.DataFromFile(t, "get_records_by_ids.json")
	responseCompaniesWithAssociations := testutils.DataFromFile(t, "companies_with_associations.json")

	tests := []testconn.TestCase[GetRecordsByIdsInput, []common.ReadResultRow]{
		{
			Name:         "Missing object name returns error",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Successfully fetch companies by IDs",
			Input: GetRecordsByIdsInput{
				ObjectName: "companies",
				Fields:     []string{"name"},
				Ids:        []string{"1bdb55e3-67f4-48d3-829b-45db3039a960", "3a95b53c-e7a1-4e53-a4e4-436f72283818"},
			},

			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/objects/companies/records/query"),
					// Attio expects a singular "filter" key with record_id $in, and a limit covering every id.
					mockcond.Body(`{
						"filter": {
							"record_id": {
								"$in": ["1bdb55e3-67f4-48d3-829b-45db3039a960", "3a95b53c-e7a1-4e53-a4e4-436f72283818"]
							}
						},
						"limit": 2
					}`),
				},
				Then: mockserver.Response(http.StatusOK, responseGetRecordsByIds),
			}.Server(),

			Expected: []common.ReadResultRow{
				{
					Id: "1bdb55e3-67f4-48d3-829b-45db3039a960",
					Fields: map[string]any{
						"name": []any{
							map[string]any{
								"active_from":  "2026-01-29T08:22:24.888000000Z",
								"active_until": nil,
								"created_by_actor": map[string]any{
									"type": "system",
									"id":   nil,
								},
								"value":          "Apple",
								"attribute_type": "text",
							},
						},
					},
					Raw: map[string]any{
						"id": map[string]any{
							"workspace_id": "e8d74639-96e5-41be-af46-ced812aef5c5",
							"object_id":    "8381cb5a-fc74-4421-aaad-3092c8bea210",
							"record_id":    "1bdb55e3-67f4-48d3-829b-45db3039a960",
						},
						"created_at": "2026-01-29T08:22:24.253000000Z",
						"web_url":    "https://app.attio.com/ampersand-test/company/1bdb55e3-67f4-48d3-829b-45db3039a960",
						"values": map[string]any{
							"record_id": []any{
								map[string]any{
									"active_from":  "2026-01-29T08:22:24.253000000Z",
									"active_until": nil,
									"created_by_actor": map[string]any{
										"type": "system",
										"id":   nil,
									},
									"value":          "1bdb55e3-67f4-48d3-829b-45db3039a960",
									"attribute_type": "text",
								},
							},
							"name": []any{
								map[string]any{
									"active_from":  "2026-01-29T08:22:24.888000000Z",
									"active_until": nil,
									"created_by_actor": map[string]any{
										"type": "system",
										"id":   nil,
									},
									"value":          "Apple",
									"attribute_type": "text",
								},
							},
						},
					},
				},
				{
					Id: "3a95b53c-e7a1-4e53-a4e4-436f72283818",
					Fields: map[string]any{
						"name": []any{
							map[string]any{
								"active_from":  "2026-01-29T08:22:23.198000000Z",
								"active_until": nil,
								"created_by_actor": map[string]any{
									"type": "system",
									"id":   nil,
								},
								"value":          "Attio",
								"attribute_type": "text",
							},
						},
					},
					Raw: map[string]any{
						"id": map[string]any{
							"workspace_id": "e8d74639-96e5-41be-af46-ced812aef5c5",
							"object_id":    "8381cb5a-fc74-4421-aaad-3092c8bea210",
							"record_id":    "3a95b53c-e7a1-4e53-a4e4-436f72283818",
						},
						"created_at": "2026-01-29T08:22:22.656000000Z",
						"web_url":    "https://app.attio.com/ampersand-test/company/3a95b53c-e7a1-4e53-a4e4-436f72283818",
						"values": map[string]any{
							"record_id": []any{
								map[string]any{
									"active_from":  "2026-01-29T08:22:22.656000000Z",
									"active_until": nil,
									"created_by_actor": map[string]any{
										"type": "system",
										"id":   nil,
									},
									"value":          "3a95b53c-e7a1-4e53-a4e4-436f72283818",
									"attribute_type": "text",
								},
							},
							"name": []any{
								map[string]any{
									"active_from":  "2026-01-29T08:22:23.198000000Z",
									"active_until": nil,
									"created_by_actor": map[string]any{
										"type": "system",
										"id":   nil,
									},
									"value":          "Attio",
									"attribute_type": "text",
								},
							},
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Fetch companies by IDs with record-reference associations",
			Input: GetRecordsByIdsInput{
				ObjectName:   "companies",
				Fields:       []string{"name"},
				Ids:          []string{"2db97cee-6c6b-4486-ae52-db8e4b6f44e9"},
				Associations: []string{"people"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/objects/companies/records/query"),
					mockcond.Body(`{
						"filter": {
							"record_id": {
								"$in": ["2db97cee-6c6b-4486-ae52-db8e4b6f44e9"]
							}
						},
						"limit": 1
					}`),
				},
				Then: mockserver.Response(http.StatusOK, responseCompaniesWithAssociations),
			}.Server(),
			Expected: []common.ReadResultRow{
				{
					Id: "2db97cee-6c6b-4486-ae52-db8e4b6f44e9",
					Fields: map[string]any{
						"name": []any{
							map[string]any{
								"active_from":  "2025-03-12T07:55:38.981000000Z",
								"active_until": nil,
								"created_by_actor": map[string]any{
									"type": "system",
									"id":   nil,
								},
								"value":          "Attio",
								"attribute_type": "text",
							},
						},
					},
					Associations: map[string][]common.Association{
						"people": {
							{ObjectId: "891dcbfc-9141-415d-9b2a-2238a6cc012d"},
							{ObjectId: "5e3fb280-007b-495a-a530-9354bde01de1"},
						},
					},
					Raw: map[string]any{
						"id": map[string]any{
							"workspace_id": "63d34516-b287-4c27-9d28-fe2adbebcd50",
							"object_id":    "1fa986a6-952e-4e92-ba01-acca61a7b616",
							"record_id":    "2db97cee-6c6b-4486-ae52-db8e4b6f44e9",
						},
						"created_at": "2025-03-12T07:55:38.327000000Z",
						"values": map[string]any{
							"name": []any{
								map[string]any{
									"active_from":  "2025-03-12T07:55:38.981000000Z",
									"active_until": nil,
									"created_by_actor": map[string]any{
										"type": "system",
										"id":   nil,
									},
									"value":          "Attio",
									"attribute_type": "text",
								},
							},
							"team": []any{
								map[string]any{
									"active_from":      "2025-03-12T07:55:39.000000000Z",
									"active_until":     nil,
									"created_by_actor": map[string]any{"type": "workspace-member", "id": "073f4c74-b60d-4de9-992a-0f799b5e442e"},
									"attribute_type":   "record-reference",
									"target_object":    "people",
									"target_record_id": "891dcbfc-9141-415d-9b2a-2238a6cc012d",
								},
								map[string]any{
									"active_from":      "2025-03-12T07:55:39.100000000Z",
									"active_until":     nil,
									"created_by_actor": map[string]any{"type": "workspace-member", "id": "073f4c74-b60d-4de9-992a-0f799b5e442e"},
									"attribute_type":   "record-reference",
									"target_object":    "people",
									"target_record_id": "5e3fb280-007b-495a-a530-9354bde01de1",
								},
							},
						},
					},
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

			result, err := conn.GetRecordsByIds(
				t.Context(), tt.Input.ObjectName, tt.Input.Ids, tt.Input.Fields, tt.Input.Associations,
			)

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

	response := testutils.DataFromFile(t, "get_records_by_ids.json")

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.Path("/v2/objects/companies/records/query"),
		Then:  mockserver.Response(http.StatusOK, response),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	// The response holds "1bdb55e3-67f4-48d3-829b-45db3039a960", "3a95b53c-e7a1-4e53-a4e4-436f72283818"; the other ids are absent from it.
	result, err := conn.GetRecordsByIds(t.Context(), "companies",
		[]string{"9c0e1f6a-0000-4000-8000-000000000001", "1bdb55e3-67f4-48d3-829b-45db3039a960", "9c0e1f6a-0000-4000-8000-000000000002", "3a95b53c-e7a1-4e53-a4e4-436f72283818"}, []string{"name"}, nil)
	if err != nil {
		t.Fatalf("a partial batch response must not fail the call: %v", err)
	}

	testconn.RequireRowIds(t, result, "1bdb55e3-67f4-48d3-829b-45db3039a960", "3a95b53c-e7a1-4e53-a4e4-436f72283818")
	testconn.RequireMissingIds(t, result, "9c0e1f6a-0000-4000-8000-000000000001", "9c0e1f6a-0000-4000-8000-000000000002")
}

// numberedRecordIds returns n distinct record ids.
func numberedRecordIds(n int) []string {
	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("00000000-0000-4000-8000-%012d", i+1)
	}

	return ids
}

// recordsQuery is the part of a list-records query body the tests inspect.
type recordsQuery struct {
	Filter struct {
		RecordID struct {
			In []string `json:"$in"`
		} `json:"record_id"`
	} `json:"filter"`
	Limit *int `json:"limit"`
}

// recordsServer answers list-records queries with one record per requested id, returning at most limit records,
// or 500 when limit is not set, as Attio does. Each query is recorded.
type recordsServer struct {
	mu      sync.Mutex
	queries []recordsQuery
	// failWhen makes the server answer 500 to a query whose ids include this id.
	failWhen string
}

func (s *recordsServer) handle(w http.ResponseWriter, r *http.Request) {
	var query recordsQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	s.mu.Lock()
	s.queries = append(s.queries, query)
	s.mu.Unlock()

	ids := query.Filter.RecordID.In
	if s.failWhen != "" && slices.Contains(ids, s.failWhen) {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	limit := 500
	if query.Limit != nil {
		limit = *query.Limit
	}

	records := make([]string, 0, len(ids))
	for _, id := range ids[:min(limit, len(ids))] {
		records = append(records, fmt.Sprintf(
			`{"id":{"workspace_id":"w","object_id":"o","record_id":%q},"created_at":"2026-01-01T00:00:00Z","values":{}}`,
			id))
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"data":[` + strings.Join(records, ",") + `]}`))
}

func getCompaniesByIds(t *testing.T, handler http.HandlerFunc, ids []string) (*common.BatchReadResult, error) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	return conn.GetRecordsByIds(t.Context(), "companies", ids, []string{"name"}, nil)
}

// Ids beyond one page are queried in chunks, each with a limit covering the whole chunk, and all rows come back.
func TestGetRecordsByIds_SplitsIdsIntoChunks(t *testing.T) {
	t.Parallel()

	ids := numberedRecordIds(2*readByIdsChunkSize + 100)
	server := &recordsServer{}

	result, err := getCompaniesByIds(t, server.handle, ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testconn.RequireRowIds(t, result, ids...)
	testconn.RequireNoFailures(t, result)

	wantSizes := []int{readByIdsChunkSize, readByIdsChunkSize, 100}
	if len(server.queries) != len(wantSizes) {
		t.Fatalf("got %d queries, want %d", len(server.queries), len(wantSizes))
	}

	for i, query := range server.queries {
		if len(query.Filter.RecordID.In) != wantSizes[i] {
			t.Errorf("query %d has %d ids, want %d", i, len(query.Filter.RecordID.In), wantSizes[i])
		}

		if query.Limit == nil || *query.Limit != wantSizes[i] {
			t.Errorf("query %d limit = %v, want %d", i, query.Limit, wantSizes[i])
		}
	}
}

// A failed chunk reports its ids with the chunk's error, keeps the rows of the other chunks,
// and returns the error alongside the result.
func TestGetRecordsByIds_ReportsFailedChunk(t *testing.T) {
	t.Parallel()

	ids := numberedRecordIds(readByIdsChunkSize + 100)
	server := &recordsServer{failWhen: ids[readByIdsChunkSize]}

	result, err := getCompaniesByIds(t, server.handle, ids)
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

// A 404 means the object in the path does not exist, so it is returned as a bad request without the 404 status,
// and is not reported as records that do not exist.
func TestGetRecordsByIds_ObjectNotFoundIsBadRequest(t *testing.T) {
	t.Parallel()

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.Path("/v2/objects/unknown/records/query"),
		Then: mockserver.ResponseString(http.StatusNotFound,
			`{"status_code":404,"type":"invalid_request_error","code":"not_found",`+
				`"message":"The referenced object was not found."}`),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "unknown",
		[]string{"1bdb55e3-67f4-48d3-829b-45db3039a960"}, []string{"name"}, nil)

	if result != nil {
		t.Errorf("expected no result, got %+v", result)
	}

	if !errors.Is(err, common.ErrBadRequest) {
		t.Errorf("error %v does not wrap common.ErrBadRequest", err)
	}

	if errors.Is(err, common.ErrNotFound) {
		t.Errorf("error %v must not wrap common.ErrNotFound", err)
	}

	if httpErr, ok := errors.AsType[*common.HTTPError](err); ok && httpErr.Status == http.StatusNotFound {
		t.Errorf("error %v must not keep the 404 status in its chain", err)
	}

	if err != nil && !strings.Contains(err.Error(), "The referenced object was not found.") {
		t.Errorf("error %q must keep the provider message", err)
	}
}
