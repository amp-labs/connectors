package proxyserv

import (
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
	"github.com/amp-labs/connectors/internal/components"
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
	Substitutions    map[string]string
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
	reader, err := credscanning.NewJSONProviderCredentials(credsFile, true)
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

// getBaseURL will use module URL if module is specified in the registry.
// Otherwise, it will fall back internally to ProviderInfo.BaseURL.
func (f Factory) getBaseURL() *url.URL {
	// Determine what module proxy should use.
	moduleID, err := f.Registry.GetString(credscanning.Fields.Module.Name)
	if err != nil {
		moduleID = string(common.ModuleRoot)
	}

	// Convert catalog variables into the registry of metadata.
	metadata := make(map[string]string)

	for _, variable := range f.CatalogVariables {
		plan := variable.GetSubstitutionPlan()
		metadata[plan.From] = plan.To
	}

	// Workspace is supplied separately from the metadata registry.
	workspace := metadata["workspace"]

	providerContext, err := components.NewProviderContext(f.Provider, common.ModuleID(moduleID), workspace, metadata)
	if err != nil {
		panic(err)
	}

	baseURL := providerContext.ModuleInfo().BaseURL

	slog.Info("Directing calls to", "url", baseURL)

	target, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}

	return target
}
