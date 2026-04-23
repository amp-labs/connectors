package hubspot

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

func TestGetPostAuthInfo(t *testing.T) {
	t.Parallel()

	accountInfo := testutils.DataFromFile(t, "post-auth/account-info.json")

	tests := []testroutines.PostAuthInfo{
		{
			Name: "Get post auth info",
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/account-info/2026-03/details"),
				Then:  mockserver.Response(http.StatusOK, accountInfo),
			}.Server(),
			Expected: &common.PostAuthInfo{
				CatalogVars: nil,
				RawResponse: &common.JSONHTTPResponse{
					Code: http.StatusOK,
				},
				ProviderWorkspaceRef: "12345678",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.AuthMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
