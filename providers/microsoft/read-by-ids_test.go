package microsoft

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestGetRecordsByIds(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseMessages := testutils.DataFromFile(t, "read/messages/batch-by-ids.json")

	tests := []testconn.TestCaseGetRecordsByIds{
		{
			Name:         "Empty record identifiers",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Read messages by identifiers",
			Input: testconn.ReadByIdsParams{
				ObjectName: "me/messages",
				RecordIds:  []string{"msg1", "msg2", "msg3"},
				Fields:     []string{"subject"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1.0/$batch"),
					mockcond.Body(`{
					  "requests": [
						{
						  "id": "me/messages_msg1",
						  "method": "GET",
						  "url": "/me/messages/msg1",
						  "headers": {
							"Content-Type": "application/json"
						  }
						},
						{
						  "id": "me/messages_msg2",
						  "method": "GET",
						  "url": "/me/messages/msg2",
						  "headers": {
							"Content-Type": "application/json"
						  }
						},
						{
						  "id": "me/messages_msg3",
						  "method": "GET",
						  "url": "/me/messages/msg3",
						  "headers": {
							"Content-Type": "application/json"
						  }
						}
					  ]
					}`),
				},
				Then: mockserver.Response(http.StatusOK, responseMessages),
			}.Server(),
			Comparator: testconn.ComparatorSortedSubsetReadByIds,
			Expected: []common.ReadResultRow{{
				Id: "msg1",
				Fields: map[string]any{
					"subject": "Hello",
				},
				Raw: map[string]any{
					"id":          "msg1",
					"subject":     "Hello",
					"bodyPreview": "Hi there",
				},
			}, {
				Id: "msg2",
				Fields: map[string]any{
					"subject": "Meeting",
				},
				Raw: map[string]any{
					"id":          "msg2",
					"subject":     "Meeting",
					"bodyPreview": "See you soon",
				},
			}, {
				Id: "msg3",
				Fields: map[string]any{
					"subject": "Lunch",
				},
				Raw: map[string]any{
					"id":          "msg3",
					"subject":     "Lunch",
					"bodyPreview": "Hungry?",
				},
			}},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableBatchReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

// batchItem renders one entry of a Graph $batch response.
func batchItem(id string, status int, body string) string {
	return fmt.Sprintf(`{"id":"me/messages_%s","status":%d,"headers":{},"body":%s}`, id, status, body)
}

func batchResponse(items ...string) string {
	return `{"responses":[` + strings.Join(items, ",") + `]}`
}

const itemNotFoundBody = `{"error":{"code":"ErrorItemNotFound","message":"The specified object was not found in the store."}}`

func getMessagesByIds(t *testing.T, serverResponse http.HandlerFunc, ids ...string) (*common.BatchReadResult, error) {
	t.Helper()

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.And{mockcond.MethodPOST(), mockcond.Path("/v1.0/$batch")},
		Then:  serverResponse,
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	return conn.GetRecordsByIds(t.Context(), "me/messages", ids, []string{"subject"}, nil)
}

// A message that returns 404 inside the $batch is reported as not found, and the other messages are returned.
func TestGetRecordsByIds_ReportsNotFoundItem(t *testing.T) {
	t.Parallel()

	result, err := getMessagesByIds(t, mockserver.ResponseString(http.StatusOK, batchResponse(
		batchItem("msg1", http.StatusOK, `{"id":"msg1","subject":"Hello"}`),
		batchItem("msg2", http.StatusNotFound, itemNotFoundBody),
		batchItem("msg3", http.StatusOK, `{"id":"msg3","subject":"Lunch"}`),
	)), "msg1", "msg2", "msg3")
	if err != nil {
		t.Fatalf("a not-found message must not produce an error, got %v", err)
	}

	testconn.RequireRowIds(t, result, "msg1", "msg3")

	if len(result.Failures) != 1 || result.Failures[0].RecordId != "msg2" ||
		result.Failures[0].Reason != common.FailureReasonNotFound {
		t.Errorf("failures = %+v, want one not_found failure for msg2", result.Failures)
	}
}

// Failures other than not-found are reported with their reason and returned as the error, alongside the
// rows of the messages that resolved.
func TestGetRecordsByIds_ReturnsRowsWithItemErrors(t *testing.T) {
	t.Parallel()

	result, err := getMessagesByIds(t, mockserver.ResponseString(http.StatusOK, batchResponse(
		batchItem("msg1", http.StatusOK, `{"id":"msg1","subject":"Hello"}`),
		batchItem("msg2", http.StatusNotFound, itemNotFoundBody),
		batchItem("msg3", http.StatusServiceUnavailable, `{"error":{"code":"serviceNotAvailable"}}`),
		batchItem("msg4", http.StatusTooManyRequests, `{"error":{"code":"TooManyRequests"}}`),
	)), "msg1", "msg2", "msg3", "msg4")
	if err == nil {
		t.Fatal("expected the 503 and 429 to be returned as the error")
	}

	if strings.Contains(err.Error(), "msg2") {
		t.Errorf("error %q must not include the not-found message", err)
	}

	testconn.RequireRowIds(t, result, "msg1")

	want := []common.RecordFailure{
		{RecordId: "msg2", Reason: common.FailureReasonNotFound},
		{RecordId: "msg3", Reason: common.FailureReasonTransient},
		{RecordId: "msg4", Reason: common.FailureReasonRateLimited},
	}

	if len(result.Failures) != len(want) {
		t.Fatalf("failures = %+v, want %+v", result.Failures, want)
	}

	for index, failure := range result.Failures {
		if failure.RecordId != want[index].RecordId || failure.Reason != want[index].Reason {
			t.Errorf("failure %d = {%s %s}, want {%s %s}", index,
				failure.RecordId, failure.Reason, want[index].RecordId, want[index].Reason)
		}
	}
}

// When the $batch request itself fails, every id in it is reported with that error and a result is still
// returned.
func TestGetRecordsByIds_ReportsFailedBatchRequest(t *testing.T) {
	t.Parallel()

	result, err := getMessagesByIds(t, mockserver.Response(http.StatusInternalServerError, nil), "msg1", "msg2")
	if !errors.Is(err, common.ErrServer) {
		t.Fatalf("error = %v, want one wrapping common.ErrServer", err)
	}

	testconn.RequireRowIds(t, result)

	if len(result.Failures) != 2 {
		t.Fatalf("failures = %+v, want one per requested id", result.Failures)
	}

	for _, failure := range result.Failures {
		if failure.Reason != common.FailureReasonTransient {
			t.Errorf("failure for %s: reason %q, want %q", failure.RecordId, failure.Reason, common.FailureReasonTransient)
		}
	}
}

// An id Graph leaves out of its $batch response is reported as not found.
func TestGetRecordsByIds_ReportsIdMissingFromResponse(t *testing.T) {
	t.Parallel()

	result, err := getMessagesByIds(t, mockserver.ResponseString(http.StatusOK, batchResponse(
		batchItem("msg1", http.StatusOK, `{"id":"msg1","subject":"Hello"}`),
	)), "msg1", "msg2")
	if err != nil {
		t.Fatalf("a missing id must not produce an error, got %v", err)
	}

	testconn.RequireRowIds(t, result, "msg1")
	testconn.RequireMissingIds(t, result, "msg2")
}

// A repeated id is requested once, because Graph rejects a $batch whose requests share an id.
func TestGetRecordsByIds_RequestsRepeatedIdOnce(t *testing.T) {
	t.Parallel()

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodPOST(),
			mockcond.Path("/v1.0/$batch"),
			mockcond.Body(`{"requests":[
				{"id":"me/messages_msg1","method":"GET","url":"/me/messages/msg1",
				 "headers":{"Content-Type":"application/json"}},
				{"id":"me/messages_msg2","method":"GET","url":"/me/messages/msg2",
				 "headers":{"Content-Type":"application/json"}}
			]}`),
		},
		Then: mockserver.ResponseString(http.StatusOK, batchResponse(
			batchItem("msg1", http.StatusOK, `{"id":"msg1","subject":"Hello"}`),
			batchItem("msg2", http.StatusOK, `{"id":"msg2","subject":"Meeting"}`),
		)),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "me/messages", []string{"msg1", "msg2", "msg1"},
		[]string{"subject"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	testconn.RequireRowIds(t, result, "msg1", "msg2")
	testconn.RequireNoFailures(t, result)
}
