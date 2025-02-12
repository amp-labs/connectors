// nolint
package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"

	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/scripts/utils/proxyserv"
)

// ================================
// Example usage
// ================================

// Create a creds.json file with the following content:
//
//	{
//		"clientId": "**************",
//		"clientSecret": "**************",
//		"scopes": "crm.contacts.read,crm.contacts.write", (optional)
//		"provider": "salesforce",
//		"substitutions": { (optional)
//		    "workspace": "some-subdomain"
//		},
//		"accessToken": "**************",
//		"refreshToken": "**************"
//	}

// Remember to run the script in the same directory as the script.
// go run proxy.go

const (
	DefaultCredsFile = "creds.json"
	DefaultPort      = 4444

	SubstitutionsFieldName = "Substitutions"
)

// ==============================
// Main (no changes needed)
// ==============================

var registry = scanning.NewRegistry()

var readers = []scanning.Reader{
	&scanning.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['substitutions']",
		KeyName:  SubstitutionsFieldName,
	},
	credscanning.Fields.Provider.GetJSONReader(DefaultCredsFile),
	credscanning.Fields.ClientId.GetJSONReader(DefaultCredsFile),
	credscanning.Fields.ClientSecret.GetJSONReader(DefaultCredsFile),
	credscanning.Fields.Scopes.GetJSONReader(DefaultCredsFile),
	credscanning.Fields.AccessToken.GetJSONReader(DefaultCredsFile),
	credscanning.Fields.RefreshToken.GetJSONReader(DefaultCredsFile),
	credscanning.Fields.Expiry.GetJSONReader(DefaultCredsFile),
	credscanning.Fields.ExpiryFormat.GetJSONReader(DefaultCredsFile),
	credscanning.Fields.ApiKey.GetJSONReader(DefaultCredsFile),
	credscanning.Fields.Username.GetJSONReader(DefaultCredsFile),
	credscanning.Fields.Password.GetJSONReader(DefaultCredsFile),
}

var debug = flag.Bool("debug", false, "Enable debug logging")

func main() {
	flag.Parse()

	err := registry.AddReaders(readers...)
	if err != nil {
		panic(err)
	}

	provider := registry.MustString(credscanning.Fields.Provider.Name)

	substitutions, err := registry.GetMap(SubstitutionsFieldName)
	if err != nil {
		slog.Warn("no substitutions, ensure that the provider info doesn't have any {{variables}}")
	}

	catalogVariables := paramsbuilder.NewCatalogVariables(substitutions)

	info, err := providers.ReadInfo(provider, catalogVariables...)
	if err != nil {
		log.Fatalf("Error reading provider info: %v", err)
	}

	if info == nil {
		log.Fatalf("Provider %s not found", provider)
	}

	// Catch Ctrl+C and handle it gracefully by shutting down the context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	proxy := createProviderProxy(ctx, info, proxyserv.Factory{
		Provider:         provider,
		CatalogVariables: catalogVariables,
		Debug:            *debug,
		Registry:         registry,
		CredsFilePath:    DefaultCredsFile,
	})

	proxy.Start(ctx, DefaultPort)
}

func createProviderProxy(
	ctx context.Context, info *providers.ProviderInfo, factory proxyserv.Factory,
) *proxyserv.Proxy {
	switch info.AuthType {
	case providers.Oauth2:
		if info.Oauth2Opts == nil {
			log.Fatalf("Missing OAuth options for provider %s", factory.Provider)
		}

		switch info.Oauth2Opts.GrantType {
		case providers.ClientCredentials:
			return factory.CreateProxyOAuth2ClientCreds(ctx)
		case providers.AuthorizationCodePKCE:
			fallthrough
		case providers.AuthorizationCode:
			return factory.CreateProxyOAuth2AuthCode(ctx)
		case providers.Password:
			return factory.CreateProxyOAuth2Password(ctx)
		default:
			log.Fatalf("Unsupported OAuth2 grant type: %s", info.Oauth2Opts.GrantType)
		}
	case providers.ApiKey:
		return factory.CreateProxyAPIKey(ctx)
	case providers.Basic:
		return factory.CreateProxyBasic(ctx)
	default:
		log.Fatalf("Unsupported auth type: %s", info.AuthType)
	}

	// unreachable
	return nil
}
