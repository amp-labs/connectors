package providers

import (
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
)

func customDomainVars(vars map[string]string) []catalogreplacer.CatalogVariable {
	result := make([]catalogreplacer.CatalogVariable, 0, len(vars))

	for key, value := range vars {
		result = append(result, catalogreplacer.CustomCatalogVariable{
			Plan: catalogreplacer.SubstitutionPlan{From: key, To: value},
		})
	}

	return result
}

// TestSalesforceCustomDomainRoutesEverythingThroughWorkspace checks that the API
// base and the token endpoint both derive from the workspace, including its path
// prefix. A gateway fronting Salesforce serves both under the same prefix.
func TestSalesforceCustomDomainRoutesEverythingThroughWorkspace(t *testing.T) {
	t.Parallel()

	const gateway = "gateway.example.com/process_salesforce_request"

	info, err := ReadInfo(SalesforceCustomDomain, customDomainVars(map[string]string{
		"workspace": gateway,
	})...)
	if err != nil {
		t.Fatalf("ReadInfo failed: %v", err)
	}

	wantBase := "https://" + gateway
	if info.BaseURL != wantBase {
		t.Errorf("BaseURL: want %q, got %q", wantBase, info.BaseURL)
	}

	moduleInfo := info.ReadModuleInfo(ModuleSalesforceCRM)
	if moduleInfo.BaseURL != wantBase {
		t.Errorf("module BaseURL: want %q, got %q", wantBase, moduleInfo.BaseURL)
	}

	wantToken := wantBase + "/services/oauth2/token"
	if info.Oauth2Opts.TokenURL != wantToken {
		t.Errorf("TokenURL: want %q, got %q", wantToken, info.Oauth2Opts.TokenURL)
	}
}

// TestSalesforceCustomDomainUsesClientCredentials guards the grant type. These
// connections are server-to-server, and the gateway does not serve the
// interactive authorize page, so an authorization code grant is not usable.
func TestSalesforceCustomDomainUsesClientCredentials(t *testing.T) {
	t.Parallel()

	info, err := ReadInfo(SalesforceCustomDomain, customDomainVars(map[string]string{
		"workspace": "gateway.example.com/process_salesforce_request",
	})...)
	if err != nil {
		t.Fatalf("ReadInfo failed: %v", err)
	}

	if info.Oauth2Opts.GrantType != ClientCredentials {
		t.Errorf("GrantType: want %q, got %q", ClientCredentials, info.Oauth2Opts.GrantType)
	}

	if info.Oauth2Opts.AuthURL != "" {
		t.Errorf("AuthURL should be unset for a client credentials grant, got %q", info.Oauth2Opts.AuthURL)
	}
}

// TestSalesforceCustomDomainDirectHost checks the non-gateway case: a workspace
// with no path prefix still yields well-formed URLs.
func TestSalesforceCustomDomainDirectHost(t *testing.T) {
	t.Parallel()

	info, err := ReadInfo(SalesforceCustomDomain, customDomainVars(map[string]string{
		"workspace": "acme.my.salesforce.com",
	})...)
	if err != nil {
		t.Fatalf("ReadInfo failed: %v", err)
	}

	wantToken := "https://acme.my.salesforce.com/services/oauth2/token"
	if info.Oauth2Opts.TokenURL != wantToken {
		t.Errorf("TokenURL: want %q, got %q", wantToken, info.Oauth2Opts.TokenURL)
	}
}

// TestSalesforceCustomDomainRequiresWorkspace checks that the workspace, which
// carries the API host, has no default — so a connection that omits it fails
// loudly at resolution rather than silently addressing the wrong host.
func TestSalesforceCustomDomainRequiresWorkspace(t *testing.T) {
	t.Parallel()

	_, err := ReadInfo(SalesforceCustomDomain, customDomainVars(map[string]string{
		"someOtherKey": "value",
	})...)
	if err == nil {
		t.Error("expected ReadInfo to fail when workspace is not supplied")
	}
}

// TestSalesforceCustomDomainResolvesWithEmptyWorkspace covers reading
// ProviderInfo without a connection, as the provider-metadata endpoint does.
// The substitution map always carries a workspace, so an empty one must resolve
// to an empty host rather than failing on a missing key.
func TestSalesforceCustomDomainResolvesWithEmptyWorkspace(t *testing.T) {
	t.Parallel()

	info, err := ReadInfo(SalesforceCustomDomain, customDomainVars(map[string]string{
		"workspace": "",
	})...)
	if err != nil {
		t.Fatalf("ReadInfo failed for an empty workspace: %v", err)
	}

	if info.BaseURL != "https://" {
		t.Errorf("BaseURL: want %q, got %q", "https://", info.BaseURL)
	}

	wantToken := "https:///services/oauth2/token"
	if info.Oauth2Opts.TokenURL != wantToken {
		t.Errorf("TokenURL: want %q, got %q", wantToken, info.Oauth2Opts.TokenURL)
	}
}

// TestSalesforceCustomDomainMirrorsSalesforceSupport guards against the twin's
// capabilities drifting from the provider it mirrors.
func TestSalesforceCustomDomainMirrorsSalesforceSupport(t *testing.T) {
	t.Parallel()

	twin, err := ReadInfo(SalesforceCustomDomain, customDomainVars(map[string]string{
		"workspace": "acme.my.salesforce.com",
	})...)
	if err != nil {
		t.Fatalf("ReadInfo for twin failed: %v", err)
	}

	base, err := ReadInfo(Salesforce, customDomainVars(map[string]string{
		"workspace": "acme",
	})...)
	if err != nil {
		t.Fatalf("ReadInfo for Salesforce failed: %v", err)
	}

	if !reflect.DeepEqual(twin.Support, base.Support) {
		t.Errorf("twin Support drifted from Salesforce\n  salesforce: %+v\n  twin:       %+v",
			base.Support, twin.Support)
	}

	if twin.DefaultModule != base.DefaultModule {
		t.Errorf("DefaultModule: want %q, got %q", base.DefaultModule, twin.DefaultModule)
	}
}
