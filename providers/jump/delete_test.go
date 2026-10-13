package jump

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	deleteResponse := testutils.DataFromFile(t, "delete/contact.json")
	unsuccessfulResponse := testutils.DataFromFile(t, "delete/contact-unsuccessful.json")
	notFoundResponse := testutils.DataFromFile(t, "delete/error-not-found.json")

	// Expected request body, the mutation is sent with the record ID and userId as variables.
	deleteRequest := testutils.DataFromFile(t, "delete/request-contact.json")

	tests := []testconn.TestCaseDelete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Objects without mutations are not supported",
			Input:        common.DeleteParams{ObjectName: "meetings", RecordId: "mtg_123"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:     "Successfully delete a contact for the connection user",
			Input:    common.DeleteParams{ObjectName: "contacts", RecordId: "con_7c1e2f0a9b3d4e5f6a7"},
			Server:   graphqlServer(deleteResponse, mockcond.BodyBytes(deleteRequest)),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:         "Delete reported as unsuccessful",
			Input:        common.DeleteParams{ObjectName: "contacts", RecordId: "con_7c1e2f0a9b3d4e5f6a7"},
			Server:       graphqlServer(unsuccessfulResponse),
			ExpectedErrs: []error{common.ErrRequestFailed},
		},
		{
			Name:   "Unknown contact returned with 200 status",
			Input:  common.DeleteParams{ObjectName: "contacts", RecordId: "con_7c1e2f0a9b3d4e5f6a7"},
			Server: graphqlServer(notFoundResponse),
			ExpectedErrs: []error{
				common.ErrNotFound,
				testutils.StringError("Contact not found"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableDeleter, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
