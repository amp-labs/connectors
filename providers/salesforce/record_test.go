package salesforce

import (
	"net/http"
	"testing"

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
