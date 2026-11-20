package hubspot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"sync"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

// Ids the batch response leaves out are reported as not-found failures, in requested order.
func TestGetRecordsByIds_ReportsMissingIds(t *testing.T) {
	t.Parallel()

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.Method(http.MethodPost),
			mockcond.Path("/crm/v3/objects/contacts/batch/read"),
		},
		Then: mockserver.ResponseString(http.StatusOK, `{
			"status": "COMPLETE",
			"results": [
				{"id": "101", "properties": {"email": "anna@example.com"}},
				{"id": "103", "properties": {"email": "cleo@example.com"}}
			]
		}`),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contacts",
		[]string{"101", "102", "103", "104"}, []string{"email"}, nil)
	if err != nil {
		t.Fatalf("a partial batch response must not fail the call: %v", err)
	}

	testconn.RequireRowIds(t, result, "101", "103")
	testconn.RequireMissingIds(t, result, "102", "104")
}

// Returned ids are not reported as failures.
func TestGetRecordsByIds_NoFailuresWhenEverythingReturned(t *testing.T) {
	t.Parallel()

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.Method(http.MethodPost),
			mockcond.Path("/crm/v3/objects/contacts/batch/read"),
		},
		Then: mockserver.ResponseString(http.StatusOK, `{
			"status": "COMPLETE",
			"results": [
				{"id": "101", "properties": {"email": "anna@example.com"}},
				{"id": "102", "properties": {"email": "bob@example.com"}}
			]
		}`),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contacts",
		[]string{"101", "102"}, []string{"email"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	testconn.RequireRowIds(t, result, "101", "102")
	testconn.RequireNoFailures(t, result)
}

// Ids the 207 errors array reports as OBJECT_NOT_FOUND are not-found failures and do not fail the call.
func TestGetRecordsByIds_ObjectNotFoundErrorIsNotFound(t *testing.T) {
	t.Parallel()

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.Method(http.MethodPost),
			mockcond.Path("/crm/v3/objects/contacts/batch/read"),
		},
		Then: mockserver.ResponseString(http.StatusMultiStatus, `{
			"status": "COMPLETE",
			"results": [
				{"id": "101", "properties": {"email": "anna@example.com"}}
			],
			"errors": [{
				"status": "error",
				"category": "OBJECT_NOT_FOUND",
				"message": "Could not get some CONTACT objects, they may be deleted or not exist.",
				"context": {"ids": ["102"]}
			}],
			"numErrors": 1
		}`),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contacts",
		[]string{"101", "102"}, []string{"email"}, nil)
	if err != nil {
		t.Fatalf("a not-found id must not fail the call: %v", err)
	}

	testconn.RequireRowIds(t, result, "101")
	testconn.RequireMissingIds(t, result, "102")
}

// Ids the 207 errors array reports with another category are failures with that reason, and the
// error is returned alongside the rows.
func TestGetRecordsByIds_ValidationErrorIsReturned(t *testing.T) {
	t.Parallel()

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.Method(http.MethodPost),
			mockcond.Path("/crm/v3/objects/contacts/batch/read"),
		},
		Then: mockserver.ResponseString(http.StatusMultiStatus, `{
			"status": "COMPLETE",
			"results": [
				{"id": "101", "properties": {"email": "anna@example.com"}}
			],
			"errors": [
				{
					"status": "error",
					"category": "VALIDATION_ERROR",
					"message": "Invalid input id",
					"context": {"ids": ["abc"]}
				},
				{
					"status": "error",
					"category": "OBJECT_NOT_FOUND",
					"message": "Could not get some CONTACT objects, they may be deleted or not exist.",
					"context": {"ids": ["103"]}
				}
			],
			"numErrors": 2
		}`),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contacts",
		[]string{"101", "abc", "103"}, []string{"email"}, nil)
	if !errors.Is(err, common.ErrBadRequest) {
		t.Fatalf("expected the validation error to be returned, got %v", err)
	}

	testconn.RequireRowIds(t, result, "101")
	requireFailureReasons(t, result, map[string]common.FailureReason{
		"abc": common.FailureReasonPermanent,
		"103": common.FailureReasonNotFound,
	})
}

// An error that names no requested id is reported for every id the response did not return,
// instead of those ids being reported as not found.
func TestGetRecordsByIds_ErrorWithoutIdsFailsUnreturnedIds(t *testing.T) {
	t.Parallel()

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.Method(http.MethodPost),
			mockcond.Path("/crm/v3/objects/contacts/batch/read"),
		},
		Then: mockserver.ResponseString(http.StatusMultiStatus, `{
			"status": "COMPLETE",
			"results": [
				{"id": "101", "properties": {"email": "anna@example.com"}}
			],
			"errors": [{
				"status": "error",
				"category": "VALIDATION_ERROR",
				"message": "Property values were not valid",
				"context": {}
			}],
			"numErrors": 1
		}`),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contacts",
		[]string{"101", "102"}, []string{"email"}, nil)
	if !errors.Is(err, common.ErrBadRequest) {
		t.Fatalf("expected the validation error to be returned, got %v", err)
	}

	testconn.RequireRowIds(t, result, "101")
	requireFailureReasons(t, result, map[string]common.FailureReason{
		"102": common.FailureReasonPermanent,
	})
}

// More than 100 ids are sent in requests of at most 100, and every row comes back in requested order.
func TestGetRecordsByIds_SplitsIdsIntoRequestsOf100(t *testing.T) {
	t.Parallel()

	server, requestSizes := newBatchReadEchoServer(t, -1)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	ids := numberedIds(250)

	result, err := conn.GetRecordsByIds(t.Context(), "contacts", ids, []string{"email"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if got := requestSizes(); !slices.Equal(got, []int{100, 100, 50}) {
		t.Errorf("request sizes %v, want [100 100 50]", got)
	}

	testconn.RequireRowIds(t, result, ids...)
	testconn.RequireNoFailures(t, result)
}

// A request that fails as a whole fails only its own ids: rows of the other requests are returned,
// and the request's error is returned with the result.
func TestGetRecordsByIds_FailedRequestKeepsOtherRows(t *testing.T) {
	t.Parallel()

	server, _ := newBatchReadEchoServer(t, 1)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	ids := numberedIds(150)

	result, err := conn.GetRecordsByIds(t.Context(), "contacts", ids, []string{"email"}, nil)
	if !errors.Is(err, common.ErrForbidden) {
		t.Fatalf("expected the failed request's error, got %v", err)
	}

	testconn.RequireRowIds(t, result, ids[:100]...)

	want := make(map[string]common.FailureReason)
	for _, id := range ids[100:] {
		want[id] = common.FailureReasonPermanent
	}

	requireFailureReasons(t, result, want)
}

// newBatchReadEchoServer returns a server that answers each batch read with a record for every
// requested id, in reverse order. The request at failIndex (0-based) gets a 403 instead; -1 never fails.
// The returned function lists the number of ids in each request, in the order they were received.
func newBatchReadEchoServer(t *testing.T, failIndex int) (*httptest.Server, func() []int) {
	t.Helper()

	var (
		mutex sync.Mutex
		sizes []int
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Inputs []struct {
				ID string `json:"id"`
			} `json:"inputs"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		mutex.Lock()
		index := len(sizes)
		sizes = append(sizes, len(body.Inputs))
		mutex.Unlock()

		w.Header().Set("Content-Type", "application/json")

		if index == failIndex {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"status": "error", "message": "missing scopes", "category": "MISSING_SCOPES"}`))

			return
		}

		results := make([]map[string]any, 0, len(body.Inputs))
		for _, input := range slices.Backward(body.Inputs) {
			results = append(results, map[string]any{
				"id":         input.ID,
				"properties": map[string]any{"email": input.ID + "@example.com"},
			})
		}

		_ = json.NewEncoder(w).Encode(map[string]any{"status": "COMPLETE", "results": results})
	}))
	t.Cleanup(server.Close)

	return server, func() []int {
		mutex.Lock()
		defer mutex.Unlock()

		return slices.Clone(sizes)
	}
}

func numberedIds(count int) []string {
	ids := make([]string, count)
	for i := range ids {
		ids[i] = strconv.Itoa(1000 + i)
	}

	return ids
}

// requireFailureReasons checks that the result's failures are exactly the ids in want, each with its reason.
func requireFailureReasons(t *testing.T, result *common.BatchReadResult, want map[string]common.FailureReason) {
	t.Helper()

	got := make(map[string]common.FailureReason, len(result.Failures))
	for _, failure := range result.Failures {
		if _, ok := got[failure.RecordId]; ok {
			t.Errorf("id %q reported more than once", failure.RecordId)
		}

		got[failure.RecordId] = failure.Reason
	}

	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("failures %v, want %v", got, want)
	}
}

// A request that fails with 404 reports its ids as not found and does not produce an error.
func TestGetRecordsByIds_NotFoundRequestReturnsNoError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status": "error", "message": "Not Found"}`))
	}))
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contacts", []string{"1", "2"}, []string{"email"}, nil)
	if err != nil {
		t.Fatalf("a not-found request must not produce an error, got %v", err)
	}

	requireFailureReasons(t, result, map[string]common.FailureReason{
		"1": common.FailureReasonNotFound,
		"2": common.FailureReasonNotFound,
	})
}
