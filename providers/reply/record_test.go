package reply

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestGetRecordsByIds(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseContactFirst := testutils.DataFromFile(t, "read/by-ids/contact-761575999.json")
	responseContactSecond := testutils.DataFromFile(t, "read/by-ids/contact-761575998.json")
	responseSequence := testutils.DataFromFile(t, "read/by-ids/sequence-1802314.json")
	errorContactNotFound := testutils.DataFromFile(t, "read/by-ids/err-contact-not-found.json")

	tests := []testconn.TestCaseGetRecordsByIds{
		{
			Name:         "Object without a get-by-id endpoint is rejected",
			Input:        testconn.ReadByIdsParams{ObjectName: "tasks", RecordIds: []string{"1"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrGetRecordNotSupportedForObject},
		},
		{
			Name:     "No ids returns no records without calling Reply",
			Input:    testconn.ReadByIdsParams{ObjectName: "contacts"},
			Server:   mockserver.Dummy(),
			Expected: []common.ReadResultRow{},
		},
		{
			Name: "Contacts are fetched one by one with the row id set",
			Input: testconn.ReadByIdsParams{
				ObjectName: "contacts",
				RecordIds:  []string{"761575999", "761575998"},
				Fields:     []string{"email", "firstName"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If:   mockcond.Path("/v3/contacts/761575999"),
					Then: mockserver.Response(http.StatusOK, responseContactFirst),
				}, {
					If:   mockcond.Path("/v3/contacts/761575998"),
					Then: mockserver.Response(http.StatusOK, responseContactSecond),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSortedSubsetReadByIds,
			Expected: []common.ReadResultRow{{
				Id: "761575998",
				Fields: map[string]any{
					"email":     "jordan.lee@example.com",
					"firstname": "Jordan",
				},
				Raw: map[string]any{"id": float64(761575998), "company": "Example Co"},
			}, {
				Id: "761575999",
				Fields: map[string]any{
					"email":     "casey.moe@example.com",
					"firstname": "Casey",
				},
				Raw: map[string]any{"id": float64(761575999), "stage": "Qualified"},
			}},
		},
		{
			Name: "Object name is matched ignoring case",
			Input: testconn.ReadByIdsParams{
				ObjectName: "Sequences",
				RecordIds:  []string{"1802314"},
				Fields:     []string{"name"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/sequences/1802314"),
				Then:  mockserver.Response(http.StatusOK, responseSequence),
			}.Server(),
			Comparator: testconn.ComparatorSortedSubsetReadByIds,
			Expected: []common.ReadResultRow{{
				Id:     "1802314",
				Fields: map[string]any{"name": "Outbound Q4"},
				Raw:    map[string]any{"id": float64(1802314), "status": "Active"},
			}},
		},
		{
			Name: "An id that no longer exists is reported as a failure and does not fail the batch",
			Input: testconn.ReadByIdsParams{
				ObjectName: "contacts",
				RecordIds:  []string{"999999999", "761575999"},
				Fields:     []string{"email"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If:   mockcond.Path("/v3/contacts/999999999"),
					Then: mockserver.Response(http.StatusNotFound, errorContactNotFound),
				}, {
					If:   mockcond.Path("/v3/contacts/761575999"),
					Then: mockserver.Response(http.StatusOK, responseContactFirst),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSortedSubsetReadByIds,
			Expected: []common.ReadResultRow{{
				Id:     "761575999",
				Fields: map[string]any{"email": "casey.moe@example.com"},
				Raw:    map[string]any{"id": float64(761575999)},
			}},
		},
		{
			Name: "Server error is returned as the error",
			Input: testconn.ReadByIdsParams{
				ObjectName: "contacts",
				RecordIds:  []string{"761575999"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/contacts/761575999"),
				Then:  mockserver.Response(http.StatusInternalServerError, nil),
			}.Server(),
			Expected:     []common.ReadResultRow{},
			ExpectedErrs: []error{common.ErrServer},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableBatchReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

// Rows come back in requested order even when an earlier id responds slower.
func TestGetRecordsByIds_PreservesRequestedOrder(t *testing.T) {
	t.Parallel()

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: mockserver.Cases{{
			If: mockcond.Path("/v3/contacts/1"),
			Then: func(w http.ResponseWriter, _ *http.Request) {
				time.Sleep(50 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"id":1,"email":"first@example.com"}`))
			},
		}, {
			If:   mockcond.Path("/v3/contacts/2"),
			Then: mockserver.ResponseString(http.StatusOK, `{"id":2,"email":"second@example.com"}`),
		}},
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contacts", []string{"1", "2"}, []string{"email"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	testconn.RequireRowIds(t, result, "1", "2")
	testconn.RequireNoFailures(t, result)
}

// A deleted id is reported as a not-found failure without an error, and the other rows keep their order.
func TestGetRecordsByIds_ReportsNotFound(t *testing.T) {
	t.Parallel()

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: mockserver.Cases{{
			If:   mockcond.Path("/v3/contacts/1"),
			Then: mockserver.ResponseString(http.StatusOK, `{"id":1,"email":"first@example.com"}`),
		}, {
			If:   mockcond.Path("/v3/contacts/404"),
			Then: mockserver.ResponseString(http.StatusNotFound, `{"errors":[{"message":"Contact not found"}]}`),
		}, {
			If:   mockcond.Path("/v3/contacts/3"),
			Then: mockserver.ResponseString(http.StatusOK, `{"id":3,"email":"third@example.com"}`),
		}},
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contacts", []string{"1", "404", "3"}, []string{"email"}, nil)
	if err != nil {
		t.Fatalf("a not-found id must not produce an error, got %v", err)
	}

	testconn.RequireRowIds(t, result, "1", "3")

	// Reply's 404 does not wrap common.ErrNotFound, so the reason is asserted rather than the sentinel.
	if len(result.Failures) != 1 || result.Failures[0].RecordId != "404" ||
		result.Failures[0].Reason != common.FailureReasonNotFound {
		t.Errorf("failures = %+v, want one not_found failure for id 404", result.Failures)
	}
}

// A non-404 failure is reported and returned as the error alongside the rows that did resolve.
func TestGetRecordsByIds_ReturnsRowsWithServerError(t *testing.T) {
	t.Parallel()

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: mockserver.Cases{{
			If:   mockcond.Path("/v3/contacts/1"),
			Then: mockserver.ResponseString(http.StatusOK, `{"id":1,"email":"first@example.com"}`),
		}, {
			If:   mockcond.Path("/v3/contacts/500"),
			Then: mockserver.Response(http.StatusInternalServerError, nil),
		}},
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contacts", []string{"500", "1"}, []string{"email"}, nil)
	if !errors.Is(err, common.ErrServer) {
		t.Fatalf("error = %v, want one wrapping common.ErrServer", err)
	}

	testconn.RequireRowIds(t, result, "1")

	if len(result.Failures) != 1 || result.Failures[0].RecordId != "500" ||
		result.Failures[0].Reason != common.FailureReasonTransient {
		t.Errorf("failures = %+v, want one transient failure for id 500", result.Failures)
	}
}

// A failing id does not stop the remaining ids from being fetched.
func TestGetRecordsByIds_AttemptsEveryIdDespiteFailure(t *testing.T) {
	t.Parallel()

	var (
		mutex   sync.Mutex
		visited = map[string]bool{}
	)

	// More ids than maxConcurrentRecordFetch, so some are queued behind the failure.
	ids := []string{"500", "1", "2", "3", "4", "5", "6", "7"}

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: mockserver.Cases{{
			If: mockcond.Path("/v3/contacts/500"),
			Then: func(w http.ResponseWriter, r *http.Request) {
				mutex.Lock()
				visited[r.URL.Path] = true
				mutex.Unlock()

				w.WriteHeader(http.StatusInternalServerError)
			},
		}},
		Default: func(w http.ResponseWriter, r *http.Request) {
			mutex.Lock()
			visited[r.URL.Path] = true
			mutex.Unlock()

			time.Sleep(20 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":1,"email":"x@example.com"}`))
		},
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := conn.GetRecordsByIds(t.Context(), "contacts", ids, []string{"email"}, nil); err == nil {
		t.Fatal("expected the server error to be returned")
	}

	mutex.Lock()
	defer mutex.Unlock()

	if len(visited) != len(ids) {
		t.Errorf("fetched %d of %d ids; a failure must not stop the rest", len(visited), len(ids))
	}
}

// A cancelled context must not make ids disappear: every requested id is reported as a failure.
func TestGetRecordsByIds_ReportsIdsWhenContextCancelled(t *testing.T) {
	t.Parallel()

	conn, err := constructTestConnector(mockserver.Dummy().URL)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	result, err := conn.GetRecordsByIds(ctx, "contacts", []string{"1", "2"}, []string{"email"}, nil)
	if err == nil {
		t.Fatal("expected an error for a cancelled context")
	}

	if len(result.Rows) != 0 || len(result.Failures) != 2 {
		t.Errorf("rows=%d failures=%+v, want no rows and a failure for each id", len(result.Rows), result.Failures)
	}
}
