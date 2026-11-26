package salesforce

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"sync"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

// responseContactsByIds is a SOQL query response holding two of the requested contacts.
const responseContactsByIds = `{
	"totalSize": 2,
	"done": true,
	"records": [
		{
			"attributes": {"type": "Contact", "url": "/services/data/v60.0/sobjects/Contact/003ak000003dQCGAA2"},
			"Id": "003ak000003dQCGAA2",
			"Name": "Anna Lee"
		},
		{
			"attributes": {"type": "Contact", "url": "/services/data/v60.0/sobjects/Contact/003ak000003dQCDAA2"},
			"Id": "003ak000003dQCDAA2",
			"Name": "Cleo Park"
		}
	]
}`

func contactsByIdsServer(t *testing.T) string {
	t.Helper()

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.Path("/services/data/v60.0/query"),
		Then:  mockserver.ResponseString(http.StatusOK, responseContactsByIds),
	}.Server()
	t.Cleanup(server.Close)

	return server.URL
}

// Returned ids are not reported as failures.
func TestGetRecordsByIds_NoFailuresWhenEverythingReturned(t *testing.T) {
	t.Parallel()

	conn, err := constructTestConnector(contactsByIdsServer(t))
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contact",
		[]string{"003ak000003dQCGAA2", "003ak000003dQCDAA2"}, []string{"Name"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	testconn.RequireRowIds(t, result, "003ak000003dQCGAA2", "003ak000003dQCDAA2")
	testconn.RequireNoFailures(t, result)
}

// Ids the query response leaves out are reported as not-found failures, in requested order.
func TestGetRecordsByIds_ReportsMissingIds(t *testing.T) {
	t.Parallel()

	conn, err := constructTestConnector(contactsByIdsServer(t))
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contact",
		[]string{"003ak000003dQXXAA1", "003ak000003dQCGAA2", "003ak000003dQYYAA1", "003ak000003dQCDAA2"},
		[]string{"Name"}, nil)
	if err != nil {
		t.Fatalf("a partial query response must not fail the call: %v", err)
	}

	testconn.RequireRowIds(t, result, "003ak000003dQCGAA2", "003ak000003dQCDAA2")
	testconn.RequireMissingIds(t, result, "003ak000003dQXXAA1", "003ak000003dQYYAA1")
}

// A query that returns nextRecordsUrl is read to its last page.
func TestGetRecordsByIds_FollowsNextRecordsUrl(t *testing.T) {
	t.Parallel()

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{{
			If: mockcond.Path("/services/data/v60.0/query/01gak000003dQZZAA2-2000"),
			Then: mockserver.ResponseString(http.StatusOK, `{
				"totalSize": 3,
				"done": true,
				"records": [
					{"attributes": {"type": "Contact"}, "Id": "003ak000003dQCDAA2", "Name": "Cleo Park"}
				]
			}`),
		}, {
			If: mockcond.Path("/services/data/v60.0/query"),
			Then: mockserver.ResponseString(http.StatusOK, `{
				"totalSize": 3,
				"done": false,
				"nextRecordsUrl": "/services/data/v60.0/query/01gak000003dQZZAA2-2000",
				"records": [
					{"attributes": {"type": "Contact"}, "Id": "003ak000003dQCGAA2", "Name": "Anna Lee"},
					{"attributes": {"type": "Contact"}, "Id": "003ak000003dQCHAA2", "Name": "Bob Kim"}
				]
			}`),
		}},
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contact",
		[]string{"003ak000003dQCGAA2", "003ak000003dQCHAA2", "003ak000003dQCDAA2"}, []string{"Name"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	testconn.RequireRowIds(t, result, "003ak000003dQCGAA2", "003ak000003dQCHAA2", "003ak000003dQCDAA2")
	testconn.RequireNoFailures(t, result)
}

// When a later page of a query fails, the rows of its earlier pages are kept and only the ids the
// query did not return are reported with the error.
func TestGetRecordsByIds_FailedNextPageKeepsEarlierRows(t *testing.T) {
	t.Parallel()

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{{
			If: mockcond.Path("/services/data/v60.0/query/01gak000003dQZZAA2-2000"),
			Then: mockserver.ResponseString(http.StatusForbidden,
				`[{"message": "insufficient access rights", "errorCode": "INSUFFICIENT_ACCESS"}]`),
		}, {
			If: mockcond.Path("/services/data/v60.0/query"),
			Then: mockserver.ResponseString(http.StatusOK, `{
				"totalSize": 3,
				"done": false,
				"nextRecordsUrl": "/services/data/v60.0/query/01gak000003dQZZAA2-2000",
				"records": [
					{"attributes": {"type": "Contact"}, "Id": "003ak000003dQCGAA2", "Name": "Anna Lee"},
					{"attributes": {"type": "Contact"}, "Id": "003ak000003dQCHAA2", "Name": "Bob Kim"}
				]
			}`),
		}},
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contact",
		[]string{"003ak000003dQCGAA2", "003ak000003dQCHAA2", "003ak000003dQCDAA2"}, []string{"Name"}, nil)
	if !errors.Is(err, common.ErrForbidden) {
		t.Fatalf("expected the failed page's error, got %v", err)
	}

	testconn.RequireRowIds(t, result, "003ak000003dQCGAA2", "003ak000003dQCHAA2")

	if len(result.Failures) != 1 || result.Failures[0].RecordId != "003ak000003dQCDAA2" ||
		result.Failures[0].Reason != common.FailureReasonPermanent {
		t.Errorf("failures %+v, want only 003ak000003dQCDAA2 with reason %q",
			result.Failures, common.FailureReasonPermanent)
	}
}

// More than 200 ids are split across queries of at most 200 ids, and every row is returned.
func TestGetRecordsByIds_SplitsIdsIntoQueriesOf200(t *testing.T) {
	t.Parallel()

	server, querySizes := newQueryByIdsEchoServer(t, -1)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	ids := contactIds(450)

	result, err := conn.GetRecordsByIds(t.Context(), "contact", ids, []string{"Name"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if got := querySizes(); !slices.Equal(got, []int{200, 200, 50}) {
		t.Errorf("query sizes %v, want [200 200 50]", got)
	}

	testconn.RequireRowIds(t, result, ids...)
	testconn.RequireNoFailures(t, result)
}

// A failed query fails only its own ids: rows of the other queries are returned, and the query's
// error is returned with the result.
func TestGetRecordsByIds_FailedQueryKeepsOtherRows(t *testing.T) {
	t.Parallel()

	server, _ := newQueryByIdsEchoServer(t, 1)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	ids := contactIds(250)

	result, err := conn.GetRecordsByIds(t.Context(), "contact", ids, []string{"Name"}, nil)
	if !errors.Is(err, common.ErrForbidden) {
		t.Fatalf("expected the failed query's error, got %v", err)
	}

	testconn.RequireRowIds(t, result, ids[:200]...)

	gotIds := make([]string, 0, len(result.Failures))

	for _, failure := range result.Failures {
		gotIds = append(gotIds, failure.RecordId)

		if failure.Reason != common.FailureReasonPermanent {
			t.Errorf("failure for %q: reason %q, want %q", failure.RecordId, failure.Reason, common.FailureReasonPermanent)
		}
	}

	if !slices.Equal(gotIds, ids[200:]) {
		t.Errorf("failed ids %v, want %v", gotIds, ids[200:])
	}
}

// soqlIdsPattern finds the quoted ids in a SOQL "Id IN (...)" condition.
var soqlIdsPattern = regexp.MustCompile(`'([^']+)'`)

// newQueryByIdsEchoServer returns a server that answers each SOQL query with a Contact for every id
// in its "Id IN" list. The query at failIndex (0-based) gets a 403 instead; -1 never fails. The
// returned function lists the number of ids in each query, in the order they were received.
func newQueryByIdsEchoServer(t *testing.T, failIndex int) (*httptest.Server, func() []int) {
	t.Helper()

	var (
		mutex sync.Mutex
		sizes []int
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		matches := soqlIdsPattern.FindAllStringSubmatch(r.URL.Query().Get("q"), -1)

		mutex.Lock()
		index := len(sizes)
		sizes = append(sizes, len(matches))
		mutex.Unlock()

		w.Header().Set("Content-Type", "application/json")

		if index == failIndex {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`[{"message": "insufficient access rights", "errorCode": "INSUFFICIENT_ACCESS"}]`))

			return
		}

		records := make([]map[string]any, 0, len(matches))
		for _, match := range matches {
			records = append(records, map[string]any{
				"attributes": map[string]any{"type": "Contact"},
				"Id":         match[1],
				"Name":       "Contact " + match[1],
			})
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"totalSize": len(records),
			"done":      true,
			"records":   records,
		})
	}))
	t.Cleanup(server.Close)

	return server, func() []int {
		mutex.Lock()
		defer mutex.Unlock()

		return slices.Clone(sizes)
	}
}

// contactIds returns count distinct 18 character Contact ids.
func contactIds(count int) []string {
	ids := make([]string, count)
	for i := range ids {
		ids[i] = fmt.Sprintf("003ak%010dAA2", i)
	}

	return ids
}

// A query that fails with 404 reports its ids as not found and does not produce an error.
func TestGetRecordsByIds_NotFoundQueryReturnsNoError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`[{"errorCode":"NOT_FOUND","message":"The requested resource does not exist"}]`))
	}))
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	ids := contactIds(2)

	result, err := conn.GetRecordsByIds(t.Context(), "contact", ids, []string{"Name"}, nil)
	if err != nil {
		t.Fatalf("a not-found query must not produce an error, got %v", err)
	}

	testconn.RequireRowIds(t, result)

	for _, failure := range result.Failures {
		if failure.Reason != common.FailureReasonNotFound {
			t.Errorf("failure for %q: reason %q, want %q", failure.RecordId, failure.Reason, common.FailureReasonNotFound)
		}
	}

	if len(result.Failures) != len(ids) {
		t.Errorf("got %d failures, want one per id", len(result.Failures))
	}
}
