package providers

import (
	"testing"
)

func TestSageIntacctClientCredentialsUsername(t *testing.T) {
	t.Parallel()

	info, err := ReadInfo(SageIntacctClientCredentials,
		createCatalogVars("workspace", "", "companyId", "acme-corp", "userId", "Ampersand_WS")...)
	if err != nil {
		t.Fatalf("ReadInfo: %v", err)
	}

	if got, want := info.Oauth2Opts.TokenParams["username"], "Ampersand_WS@acme-corp"; got != want {
		t.Errorf("username = %q, want %q", got, want)
	}
}
