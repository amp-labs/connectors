package proxyserv

import (
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
	"github.com/amp-labs/connectors/providers"
	"golang.org/x/oauth2"
)

// Factory holds arguments used to create Proxy of any type.
type Factory struct {
	Provider         string
	CatalogVariables []catalogreplacer.CatalogVariable
	Debug            bool
	Registry         scanning.Registry
	CredsFilePath    string
}

type ClientAuthParams struct {
	ID     string
	Secret string
	Scopes []string
}

func createClientAuthParams(provider string, registry scanning.Registry) *ClientAuthParams {
	clientId := registry.MustString(credscanning.Fields.ClientId.Name)
	clientSecret := registry.MustString(credscanning.Fields.ClientSecret.Name)

	scopes, err := registry.GetString(credscanning.Fields.Scopes.Name)
	if err != nil {
		slog.Warn("no scopes attached, ensure that the provider doesn't require scopes")
	}

	validateRequiredOAuth2Flags(provider, clientId, clientSecret)

	return &ClientAuthParams{
		ID:     clientId,
		Secret: clientSecret,
		Scopes: strings.Split(scopes, ","),
	}
}

func validateRequiredOAuth2Flags(provider, clientId, clientSecret string) {
	if provider == "" || clientId == "" || clientSecret == "" {
		_, _ = fmt.Fprintln(os.Stderr, "Missing required flags: -provider, -client-id, -client-secret")

		flag.Usage()
		os.Exit(1)
	}
}

func getTokensFromRegistry(credsFile string) *oauth2.Token {
	reader, err := credscanning.NewJSONProviderCredentials(credsFile, true, false)
	if err != nil {
		panic(err)
	}

	return reader.GetOauthToken()
}

func getProviderConfig(provider string, catalogVariables []catalogreplacer.CatalogVariable) *providers.ProviderInfo {
	config, err := providers.ReadInfo(provider, catalogVariables...)
	if err != nil {
		panic(fmt.Errorf("%w: %s", err, provider))
	}

	return config
}

func getBaseURL(providerInfo *providers.ProviderInfo) *url.URL {
	target, err := url.Parse(providerInfo.BaseURL)
	if err != nil {
		panic(err)
	}

	return target
}
