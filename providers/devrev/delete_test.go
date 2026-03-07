package devrev

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	// DevRev accounts-delete-response from schema: empty object {}
	responseAccountsDelete := testutils.DataFromFile(t, "write-accounts-delete-response.json")

	tests := []testroutines.Delete{
		{
			Name: "Delete account successfully",
			Input: common.DeleteParams{
				ObjectName: "accounts",
				RecordId:   "don:identity:devrev:ACCT-1",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/accounts.delete"),
					mockcond.Body(`{"id":"don:identity:devrev:ACCT-1"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountsDelete),
			}.Server(),
			Expected: &common.DeleteResult{
				Success: true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
