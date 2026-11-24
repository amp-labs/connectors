package freshchat

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	createdUser := []byte(`{"id": "u-1", "first_name": "Milton", "email": "milton@example.com"}`)

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Create posts to the collection and returns the new record",
			Input: common.WriteParams{ObjectName: "users", RecordData: map[string]any{"first_name": "Milton"}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/users"),
				},
				Then: mockserver.Response(http.StatusCreated, createdUser),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "u-1",
				Data: map[string]any{
					"id":         "u-1",
					"first_name": "Milton",
					"email":      "milton@example.com",
				},
			},
		},
		{
			// Freshchat answers an update with no body, so the identifier
			// comes from the record that was addressed.
			Name: "Update puts to the record and returns its identifier",
			Input: common.WriteParams{
				ObjectName: "users",
				RecordId:   "u-1",
				RecordData: map[string]any{"last_name": "Doe"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v2/users/u-1"),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "u-1",
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
