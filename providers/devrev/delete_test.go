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
		{
			Name: "Delete rev-user successfully (hyphenated object)",
			Input: common.DeleteParams{
				ObjectName: "rev-users",
				RecordId:   "1b5d9e8e-6e12-4a0a-bf67-2a8e34c8e2aa",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/rev-users.delete"),
					mockcond.Body(`{"id":"1b5d9e8e-6e12-4a0a-bf67-2a8e34c8e2aa"}`),
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
