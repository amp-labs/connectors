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

// TestSalesforceCustomDomainSeparatesHosts checks that the API host and the OAuth
// host resolve independently, which is the whole point of this provider: data
// traffic may be routed through a gateway while OAuth stays on Salesforce.
func TestSalesforceCustomDomainSeparatesHosts(t *testing.T) {
	t.Parallel()

	info, err := ReadInfo(SalesforceCustomDomain, customDomainVars(map[string]string{
		"apiDomain":  "gateway.example.com/salesforce",
		"authDomain": "login.salesforce.com",
	})...)
	if err != nil {
		t.Fatalf("ReadInfo failed: %v", err)
	}

	wantBase := "https://gateway.example.com/salesforce"
	if info.BaseURL != wantBase {
		t.Errorf("BaseURL: want %q, got %q", wantBase, info.BaseURL)
	}

	moduleInfo := info.ReadModuleInfo(ModuleSalesforceCRM)
	if moduleInfo.BaseURL != wantBase {
		t.Errorf("module BaseURL: want %q, got %q", wantBase, moduleInfo.BaseURL)
	}

	wantAuth := "https://login.salesforce.com/services/oauth2/authorize"
	if info.Oauth2Opts.AuthURL != wantAuth {
		t.Errorf("AuthURL: want %q, got %q", wantAuth, info.Oauth2Opts.AuthURL)
	}

	wantToken := "https://login.salesforce.com/services/oauth2/token"
	if info.Oauth2Opts.TokenURL != wantToken {
		t.Errorf("TokenURL: want %q, got %q", wantToken, info.Oauth2Opts.TokenURL)
	}
}

// TestSalesforceCustomDomainAuthDomainDefaults checks that authDomain falls back
// to login.salesforce.com. Without a default, an unsupplied value would fail the
// catalog's missingkey=error substitution rather than resolving sensibly.
func TestSalesforceCustomDomainAuthDomainDefaults(t *testing.T) {
	t.Parallel()

	info, err := ReadInfo(SalesforceCustomDomain, customDomainVars(map[string]string{
		"apiDomain": "acme.my.salesforce.com",
	})...)
	if err != nil {
		t.Fatalf("ReadInfo failed: %v", err)
	}

	wantAuth := "https://login.salesforce.com/services/oauth2/authorize"
	if info.Oauth2Opts.AuthURL != wantAuth {
		t.Errorf("AuthURL: want %q, got %q", wantAuth, info.Oauth2Opts.AuthURL)
	}
}

// TestSalesforceCustomDomainRequiresAPIDomain checks that apiDomain has no
// default, so a connection that omits it fails loudly at resolution rather than
// silently addressing the wrong host.
func TestSalesforceCustomDomainRequiresAPIDomain(t *testing.T) {
	t.Parallel()

	_, err := ReadInfo(SalesforceCustomDomain, customDomainVars(map[string]string{
		"authDomain": "login.salesforce.com",
	})...)
	if err == nil {
		t.Error("expected ReadInfo to fail when apiDomain is not supplied")
	}
}

// TestSalesforceCustomDomainMirrorsSalesforceSupport guards against the twin's
// capabilities drifting from the provider it mirrors.
func TestSalesforceCustomDomainMirrorsSalesforceSupport(t *testing.T) {
	t.Parallel()

	twin, err := ReadInfo(SalesforceCustomDomain, customDomainVars(map[string]string{
		"apiDomain":  "acme.my.salesforce.com",
		"authDomain": "login.salesforce.com",
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
