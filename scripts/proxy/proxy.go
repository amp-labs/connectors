package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/amp-labs/connectors/basic"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/utils"
	"golang.org/x/oauth2"
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

var (
	DefaultCredsFile = "creds.json"
	DefaultPort      = 4444
)

// ==============================
// Main (no changes needed)
// ==============================

var registry = utils.NewCredentialsRegistry()

var readers = []utils.Reader{
	&utils.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['clientId']",
		CredKey:  "ClientId",
	},
	&utils.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['clientSecret']",
		CredKey:  "ClientSecret",
	},
	&utils.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['scopes']",
		CredKey:  "Scopes",
	},
	&utils.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['provider']",
		CredKey:  "Provider",
	},
	&utils.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['substitutions']",
		CredKey:  "Substitutions",
	},
	&utils.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['accessToken']",
		CredKey:  "AccessToken",
	},
	&utils.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['refreshToken']",
		CredKey:  "RefreshToken",
	},
}

func main() {
	err := registry.AddReaders(readers...)
	if err != nil {
		panic(err)
	}

	provider := registry.MustString("Provider")
	clientId := registry.MustString("ClientId")
	clientSecret := registry.MustString("ClientSecret")
	accessToken := registry.MustString("AccessToken")
	refreshToken := registry.MustString("RefreshToken")

	scopes, err := registry.GetString("Scopes")
	if err != nil {
		slog.Warn("no scopes attached, ensure that the provider doesn't require scopes")
	}

	oauthScopes := strings.Split(scopes, ",")

	substitutions, err := registry.GetMap("Substitutions")
	if err != nil {
		slog.Warn("no substitutions, ensure that the provider info doesn't have any {{variables}}")
	}

	// Cast the substitutions to a map[string]string
	substitutionsMap := make(map[string]string)
	for key, val := range substitutions {
		substitutionsMap[key] = val.MustString()
	}
	validateRequiredFlags(provider, clientId, clientSecret)
	startProxy(provider, oauthScopes, clientId, clientSecret, substitutionsMap, DefaultPort, accessToken, refreshToken)
}

func validateRequiredFlags(provider, clientId, clientSecret string) {
	if provider == "" || clientId == "" || clientSecret == "" {
		fmt.Fprintln(os.Stderr, "Missing required flags: -provider, -client-id, -client-secret")
		flag.Usage()
		os.Exit(1)
	}
}

func startProxy(provider string, scopes []string, clientId, clientSecret string, substitutions map[string]string, port int, accessToken, refreshToken string) {
	proxy := buildProxy(provider, scopes, clientId, clientSecret, substitutions, accessToken, refreshToken)
	http.Handle("/", proxy)

	fmt.Printf("\nProxy server listening on :%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil { // nosemgrep
		panic(err)
	}
}

func buildProxy(provider string, scopes []string, clientId, clientSecret string, substitutions map[string]string, accessToken, refreshToken string) *Proxy {
	providerInfo := getProviderConfig(provider, substitutions)
	cfg := configureOAuth(clientId, clientSecret, scopes, providerInfo)
	httpClient := setupHttpClient(cfg, accessToken, refreshToken, provider, providerInfo)

	target, err := url.Parse(providerInfo.BaseURL)
	if err != nil {
		panic(err)
	}

	return newProxy(target, httpClient)
}

func getProviderConfig(provider string, substitutions map[string]string) *providers.ProviderInfo {
	config, err := providers.ReadConfig(provider, &substitutions)
	if err != nil {
		panic(err)
	}

	return config
}

func readToken(tokenPath string) *oauth2.Token {
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		panic(err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		panic(err)
	}
	return &token
}

func configureOAuth(clientId, clientSecret string, scopes []string, providerInfo *providers.ProviderInfo) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:   providerInfo.OauthOpts.AuthURL,
			TokenURL:  providerInfo.OauthOpts.TokenURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

// This helps with refreshing tokens automatically.
func setupHttpClient(cfg *oauth2.Config, accessToken, refreshToken string, provider string, providerInfo *providers.ProviderInfo) *http.Client {
	ctx := context.Background()
	conn, err := basic.NewConnector(
		providers.Provider(provider),
		basic.WithClient(ctx, http.DefaultClient, cfg, &oauth2.Token{AccessToken: accessToken, RefreshToken: refreshToken}),
	)
	if err != nil {
		panic(err)
	}

	providerHTTPClient, ok := conn.HTTPClient().Client.(*http.Client)
	if !ok {
		panic("not an http client")
	}

	return providerHTTPClient
}

type Proxy struct {
	*httputil.ReverseProxy
	target *url.URL
}

func newProxy(target *url.URL, httpClient *http.Client) *Proxy {
	reverseProxy := httputil.NewSingleHostReverseProxy(target)
	reverseProxy.Transport = &customTransport{httpClient}

	return &Proxy{
		ReverseProxy: reverseProxy,
		target:       target,
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Host = p.target.Host
	r.Host = p.target.Host
	r.RequestURI = "" // Must be cleared

	fmt.Printf("Proxying request: %s %s%s\n", r.Method, r.URL.Host, r.URL.Path)
	p.ReverseProxy.ServeHTTP(w, r)
}

type customTransport struct {
	httpClient *http.Client
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.httpClient.Do(req)
}

func convertSubstitutions(sub string) map[string]string {
	if len(sub) == 0 {
		return nil
	}

	parts := strings.Split(sub, ",")

	substitutions := make(map[string]string)

	for _, part := range parts {
		parts := strings.Split(part, "=")
		if len(parts) != 2 {
			continue
		}

		substitutions[parts[0]] = parts[1]
	}

	return substitutions
}
