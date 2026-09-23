package hubspot

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
)

// HubSpot's batch/read omits ids it cannot find rather than erroring on them,
// so an id missing from the response is the only signal that the record is
// gone. Before this was reported, the caller saw a short list and could not
// tell a deleted record from one it never asked for.
func TestGetRecordsByIds_MissingIdsAreReported(t *testing.T) {
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

	if len(result.Rows) != 2 {
		t.Fatalf("expected the 2 returned records as rows, got %d", len(result.Rows))
	}

	if len(result.Failures) != 2 {
		t.Fatalf("expected the 2 absent ids reported, got %d (%+v)", len(result.Failures), result.Failures)
	}

	// Reported in the order they were requested, not in map order.
	if result.Failures[0].RecordId != "102" || result.Failures[1].RecordId != "104" {
		t.Errorf("failures %q, %q — want 102 then 104",
			result.Failures[0].RecordId, result.Failures[1].RecordId)
	}

	for _, failure := range result.Failures {
		if failure.Reason != common.FailureReasonNotFound {
			t.Errorf("failure for %q: reason %q, want %q",
				failure.RecordId, failure.Reason, common.FailureReasonNotFound)
		}
	}
}

// The ids that were returned must not be reported, which is what catches a
// connector whose rows are keyed differently from its request.
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

	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result.Rows))
	}

	if result.HasFailures() {
		t.Fatalf("no id was missing, but failures were reported: %+v", result.Failures)
	}
}
