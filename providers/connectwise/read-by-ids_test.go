package connectwise

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

func TestGetRecordsByIds(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseContacts := testutils.DataFromFile(t, "read/contacts-batch-by-ids.json")

	tests := []testconn.TestCaseGetRecordsByIds{
		{
			Name:         "Empty record identifiers",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Read contacts by identifiers",
			Input: testconn.ReadByIdsParams{
				ObjectName: "contacts",
				RecordIds:  []string{"57920", "57921", "57922"},
				Fields:     []string{"firstName", "customField53", "AMPERSAND-defaultPhone", "AMPERSAND-defaultPhoneId"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v4_6_release/apis/3.0/company/contacts"),
					mockcond.Permute(
						queryParam("conditions", "id in (%v)"),
						"57920", "57921", "57922",
					),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testconn.ComparatorSortedSubsetReadByIds,
			Expected: []common.ReadResultRow{{
				Id: "57920",
				Fields: map[string]any{
					"firstname":                "Maxime Schaefer [1]",
					"customfield53":            "Software Developer",
					"ampersand-defaultphone":   "+123",
					"ampersand-defaultphoneid": "111",
				},
				Raw: map[string]any{"lastName": "Wayne Blanda"},
			}, {
				Id: "57921",
				Fields: map[string]any{
					"firstname":                "Roderick Rippin [2]",
					"customfield53":            "Sales Representative",
					"ampersand-defaultphone":   "+456",
					"ampersand-defaultphoneid": "222",
				},
				Raw: map[string]any{"lastName": "Lemuel Hackett"},
			}, {
				Id: "57922",
				Fields: map[string]any{
					"firstname":                "Randi Haag [3]",
					"customfield53":            "Manager",
					"ampersand-defaultphone":   "+789",
					"ampersand-defaultphoneid": "333",
				},
				Raw: map[string]any{"lastName": "Ferne Bradtke"},
			}},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableBatchReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}

func queryParam(key, valueTemplate string) func(fields []string) mockcond.Condition {
	return func(identifiers []string) mockcond.Condition {
		param := fmt.Sprintf(valueTemplate, strings.Join(identifiers, ","))

		return mockcond.QueryParam(key, param)
	}
}

// An id that the provider does not return is reported as not found, without an error.
func TestGetRecordsByIds_ReportsMissingId(t *testing.T) {
	t.Parallel()

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/v4_6_release/apis/3.0/company/contacts"),
		},
		Then: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "read/contacts-batch-by-ids.json")),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server)
	if err != nil {
		t.Fatal(err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contacts",
		[]string{"57920", "57921", "99999", "57922"}, []string{"firstName"}, nil)
	if err != nil {
		t.Fatalf("a missing id must not produce an error, got %v", err)
	}

	if len(result.Rows) != 3 {
		t.Errorf("got %d rows, want 3", len(result.Rows))
	}

	testconn.RequireMissingIds(t, result, "99999")
}

// When one request fails, the rows of the other requests are kept, every id of the failed request is
// reported with its error, and the error is returned.
func TestGetRecordsByIds_FailedRequestKeepsOtherRows(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	// Enough ids that the conditions query is split across several URLs.
	recordIds := make([]string, 0, 600)
	for id := 100000; id < 100600; id++ {
		recordIds = append(recordIds, strconv.Itoa(id))
	}

	const failingId = "100000"

	var (
		mutex        sync.Mutex
		idLists      = make(map[string]bool) // first id of each distinct request
		failedIdList []string
	)

	server := mockserver.NewServer(func(w http.ResponseWriter, r *http.Request) {
		ids := idsFromConditions(t, r.URL.Query().Get("conditions"))

		mutex.Lock()
		idLists[ids[0]] = true

		if slices.Contains(ids, failingId) {
			failedIdList = ids
		}
		mutex.Unlock()

		w.Header().Set("Content-Type", "application/json")

		if slices.Contains(ids, failingId) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code":"ConnectWiseApi","message":"Unexpected error"}`))

			return
		}

		records := make([]map[string]any, 0, len(ids))
		for _, id := range ids {
			number, _ := strconv.Atoi(id)
			records = append(records, map[string]any{"id": number, "firstName": "Name " + id})
		}

		body, _ := json.Marshal(records)
		_, _ = w.Write(body)
	})
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server)
	if err != nil {
		t.Fatal(err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contacts", recordIds, []string{"firstName"}, nil)
	if !errors.Is(err, common.ErrServer) {
		t.Fatalf("error = %v, want one wrapping common.ErrServer", err)
	}

	if len(idLists) < 2 {
		t.Fatalf("got %d distinct requests, the test needs the ids split across several", len(idLists))
	}

	if result == nil {
		t.Fatal("expected a result alongside the error")
	}

	failed := make(map[string]bool, len(result.Failures))

	for _, failure := range result.Failures {
		failed[failure.RecordId] = true

		if failure.Reason != common.FailureReasonTransient {
			t.Errorf("failure for %s: reason %q, want %q", failure.RecordId, failure.Reason, common.FailureReasonTransient)
		}
	}

	if len(failed) != len(failedIdList) {
		t.Errorf("got %d failed ids, want the %d ids of the failed request", len(failed), len(failedIdList))
	}

	for _, id := range failedIdList {
		if !failed[id] {
			t.Errorf("id %s of the failed request is not reported", id)
		}
	}

	if len(result.Rows) != len(recordIds)-len(failedIdList) {
		t.Errorf("got %d rows, want %d", len(result.Rows), len(recordIds)-len(failedIdList))
	}

	for _, row := range result.Rows {
		if failed[row.Id] {
			t.Errorf("id %s is both a row and a failure", row.Id)
		}
	}
}

// idsFromConditions reads the ids out of a "id in (1,2,3)" conditions query.
func idsFromConditions(t *testing.T, conditions string) []string {
	t.Helper()

	list, ok := strings.CutPrefix(conditions, "id in (")
	if !ok {
		t.Errorf("unexpected conditions %q", conditions)

		return nil
	}

	return strings.Split(strings.TrimSuffix(list, ")"), ",")
}

// A request that fails with 404 reports its ids as not found and does not produce an error.
func TestGetRecordsByIds_NotFoundRequestReturnsNoError(t *testing.T) {
	t.Parallel()

	recordIds := make([]string, 0, 600)
	for id := 100000; id < 100600; id++ {
		recordIds = append(recordIds, strconv.Itoa(id))
	}

	const failingId = "100000"

	server := mockserver.NewServer(func(w http.ResponseWriter, r *http.Request) {
		ids := idsFromConditions(t, r.URL.Query().Get("conditions"))

		w.Header().Set("Content-Type", "application/json")

		if slices.Contains(ids, failingId) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"code":"NotFound","message":"Not found"}`))

			return
		}

		records := make([]map[string]any, 0, len(ids))
		for _, id := range ids {
			number, _ := strconv.Atoi(id)
			records = append(records, map[string]any{"id": number})
		}

		body, _ := json.Marshal(records)
		_, _ = w.Write(body)
	})
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server)
	if err != nil {
		t.Fatal(err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contacts", recordIds, []string{"id"}, nil)
	if err != nil {
		t.Fatalf("a not-found request must not produce an error, got %v", err)
	}

	if !result.HasFailures() {
		t.Fatal("expected the ids of the 404 request to be reported")
	}

	for _, failure := range result.Failures {
		if failure.Reason != common.FailureReasonNotFound {
			t.Errorf("failure for %s: reason %q, want %q", failure.RecordId, failure.Reason, common.FailureReasonNotFound)
		}
	}

	if len(result.Rows)+len(result.Failures) != len(recordIds) {
		t.Errorf("got %d rows and %d failures, want %d ids in all", len(result.Rows), len(result.Failures), len(recordIds))
	}
}
