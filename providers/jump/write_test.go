package jump

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

const testUserID = "usr_42fe28110261430cadc"

func TestWrite(t *testing.T) {
	t.Parallel()

	createResponse := testutils.DataFromFile(t, "write/contact-create.json")
	updateResponse := testutils.DataFromFile(t, "write/contact-update.json")
	validationErrorResponse := testutils.DataFromFile(t, "write/error-validation.json")

	// Expected request bodies, the mutation is sent with the record fields and userId as variables.
	createRequest := testutils.DataFromFile(t, "write/request-contact-create.json")
	updateRequest := testutils.DataFromFile(t, "write/request-contact-update.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Objects without mutations are not supported",
			Input:        common.WriteParams{ObjectName: "meetings", RecordData: map[string]any{"title": "Review"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Successfully create a contact for the connection user",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordData: map[string]any{
					"firstName": "amp-test",
					"lastName":  "Write 20261027135509",
					"email":     "amp-test-write-20261027135509@example.com",
				},
			},
			Server:     graphqlServer(createResponse, mockcond.BodyBytes(createRequest)),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "con_2feda6a76d294253be5",
				Data: map[string]any{
					"id":    "con_2feda6a76d294253be5",
					"name":  "amp-test Write 20261027135509",
					"email": "amp-test-write-20261027135509@example.com",
				},
			},
		},
		{
			Name: "Update sends only the provided fields",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "con_2feda6a76d294253be5",
				RecordData: map[string]any{
					"lastName": "Write 20261027135509 Updated",
					"email":    "amp-test-write-20261027135509-updated@example.com",
				},
			},
			Server:     graphqlServer(updateResponse, mockcond.BodyBytes(updateRequest)),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "con_2feda6a76d294253be5",
				Data: map[string]any{
					"name":  "amp-test Write 20261027135509 Updated",
					"email": "amp-test-write-20261027135509-updated@example.com",
				},
			},
		},
		{
			Name: "Validation error returned with 200 status",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordData: map[string]any{
					"firstName": "amp-test",
					"lastName":  "Contact",
					"email":     "not-an-email",
				},
			},
			Server: graphqlServer(validationErrorResponse),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("Validation failed"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWriter, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
