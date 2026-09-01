package salesforce

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
)

const userinfoResponse = `{
	"sub": "https://login.salesforce.com/id/00Dxx0000001gPFEAY/005xx000001SvokAAC",
	"user_id": "005xx000001SvokAAC",
	"organization_id": "00Dxx0000001gPFEAY",
	"preferred_username": "integration@example.com",
	"name": "Integration User"
}`

func TestGetCurrentUsername(t *testing.T) {
	t.Parallel()

	t.Run("Returns preferred_username from userinfo", func(t *testing.T) {
		t.Parallel()

		server := mockserver.Conditional{
			Setup: mockserver.ContentJSON(),
			If:    mockcond.Path("/services/oauth2/userinfo"),
			Then:  mockserver.Response(http.StatusOK, []byte(userinfoResponse)),
		}.Server()
		defer server.Close()

		connector, err := constructTestConnector(server.URL)
		if err != nil {
			t.Fatalf("failed to construct connector: %v", err)
		}

		username, err := connector.GetCurrentUsername(t.Context())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if username != "integration@example.com" {
			t.Errorf("username = %q, want integration@example.com", username)
		}
	})

	t.Run("Errors when preferred_username is absent", func(t *testing.T) {
		t.Parallel()

		server := mockserver.Fixed{
			Setup:  mockserver.ContentJSON(),
			Always: mockserver.Response(http.StatusOK, []byte(`{"sub": "only"}`)),
		}.Server()
		defer server.Close()

		connector, err := constructTestConnector(server.URL)
		if err != nil {
			t.Fatalf("failed to construct connector: %v", err)
		}

		if _, err := connector.GetCurrentUsername(t.Context()); err == nil {
			t.Error("expected error for missing preferred_username")
		}
	})
}

func TestGetPostAuthInfo(t *testing.T) {
	t.Parallel()

	t.Run("Reports the username as a catalog var", func(t *testing.T) {
		t.Parallel()

		server := mockserver.Conditional{
			Setup: mockserver.ContentJSON(),
			If:    mockcond.Path("/services/oauth2/userinfo"),
			Then:  mockserver.Response(http.StatusOK, []byte(userinfoResponse)),
		}.Server()
		defer server.Close()

		connector, err := constructTestConnector(server.URL)
		if err != nil {
			t.Fatalf("failed to construct connector: %v", err)
		}

		info, err := connector.GetPostAuthInfo(t.Context())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if info.CatalogVars == nil {
			t.Fatal("expected catalog vars, got nil")
		}

		if got := (*info.CatalogVars)[PostAuthCatalogVarUsername]; got != "integration@example.com" {
			t.Errorf("username catalog var = %q, want integration@example.com", got)
		}
	})

	t.Run("Degrades to empty info when userinfo is scope-blocked", func(t *testing.T) {
		t.Parallel()

		// Tokens minted without an identity scope get a 403 from userinfo.
		// GetPostAuthInfo must NOT propagate the error: the server treats a
		// post-auth error as fatal for connection creation.
		server := mockserver.Fixed{
			Setup:  mockserver.ContentJSON(),
			Always: mockserver.Response(http.StatusForbidden, []byte(`[{"errorCode":"FORBIDDEN"}]`)),
		}.Server()
		defer server.Close()

		connector, err := constructTestConnector(server.URL)
		if err != nil {
			t.Fatalf("failed to construct connector: %v", err)
		}

		info, err := connector.GetPostAuthInfo(t.Context())
		if err != nil {
			t.Fatalf("expected graceful degradation, got error: %v", err)
		}

		if info == nil {
			t.Fatal("expected empty PostAuthInfo, got nil")
		}

		if info.CatalogVars != nil {
			t.Errorf("expected no catalog vars on failure, got %v", *info.CatalogVars)
		}
	})
}
