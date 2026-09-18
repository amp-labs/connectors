package hubspot

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestGetPostAuthInfo(t *testing.T) {
	t.Parallel()

	accountInfo := testutils.DataFromFile(t, "post-auth/account-info.json")
	accountInfoEU := testutils.DataFromFile(t, "post-auth/account-info-eu.json")

	tests := []testconn.TestCaseGetPostAuthInfo{
		{
			Name: "Get post auth info",
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/account-info/2026-03/details"),
				Then:  mockserver.Response(http.StatusOK, accountInfo),
			}.Server(),
			Expected: &common.PostAuthInfo{
				CatalogVars: &map[string]string{
					providers.HubspotApiDomainVar: "api.hubapi.com",
				},
				RawResponse: &common.JSONHTTPResponse{
					Code: http.StatusOK,
				},
				ProviderWorkspaceRef: "12345678",
			},
		},
		{
			// A portal hosted in eu1 must address its own hublet; the unsuffixed
			// host answers 488 for it.
			Name: "EU portal resolves to its own hublet host",
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/account-info/2026-03/details"),
				Then:  mockserver.Response(http.StatusOK, accountInfoEU),
			}.Server(),
			Expected: &common.PostAuthInfo{
				CatalogVars: &map[string]string{
					providers.HubspotApiDomainVar: "api-eu1.hubapi.com",
				},
				RawResponse: &common.JSONHTTPResponse{
					Code: http.StatusOK,
				},
				ProviderWorkspaceRef: "5865761",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestablePostAuthMetadata, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
