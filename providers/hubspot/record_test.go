package hubspot

import (
	"net/http"
	"testing"

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
