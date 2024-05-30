// nolint
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/amp-labs/connectors/connector"
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
	&utils.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['expiry']",
		CredKey:  "Expiry",
	},
	&utils.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['expiryFormat']",
		CredKey:  "ExpiryFormat",
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
	tokens := getTokensFromRegistry()

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
	startProxy(provider, oauthScopes, clientId, clientSecret, substitutionsMap, DefaultPort, tokens)
}

// Some connectors may implement Refresh tokens, when it happens expiry must be provided alongside.
// Library shouldn't attempt to refresh tokens if API doesn't support `refresh_token` grant type.
func getTokensFromRegistry() *oauth2.Token {
	accessToken := registry.MustString("AccessToken")
	refreshToken, err := registry.GetString("RefreshToken")
	if err != nil {
		// we are working without refresh token
		return &oauth2.Token{
			AccessToken: accessToken,
		}
	}

	// refresh token should be specified with expiry
	atExpiry := registry.MustString("Expiry")
	atExpiryTimeFormat := registry.MustString("ExpiryFormat")
	expiry := parseAccessTokenExpiry(atExpiry, atExpiryTimeFormat)

	return &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry, // required: will trigger reuse of refresh token
	}
}

func parseAccessTokenExpiry(expiryStr, timeFormat string) time.Time {
	formatEnums := map[string]string{
		"Layout":      time.Layout,
		"ANSIC":       time.ANSIC,
		"UnixDate":    time.UnixDate,
		"RubyDate":    time.RubyDate,
		"RFC822":      time.RFC822,
		"RFC822Z":     time.RFC822Z,
		"RFC850":      time.RFC850,
		"RFC1123":     time.RFC1123,
		"RFC1123Z":    time.RFC1123Z,
		"RFC3339":     time.RFC3339,
		"RFC3339Nano": time.RFC3339Nano,
		"Kitchen":     time.Kitchen,
		"DateOnly":    time.DateOnly,
	}

	format, found := formatEnums[timeFormat]
	if !found {
		// specific format is specified instead of enum
		format = timeFormat
	}

	expiry, err := time.Parse(format, expiryStr)
	if err != nil {
		panic(err)
	}

	return expiry
}

func validateRequiredFlags(provider, clientId, clientSecret string) {
	if provider == "" || clientId == "" || clientSecret == "" {
		fmt.Fprintln(os.Stderr, "Missing required flags: -provider, -client-id, -client-secret")
		flag.Usage()
		os.Exit(1)
	}
}

func startProxy(provider string, scopes []string, clientId, clientSecret string, substitutions map[string]string, port int, tokens *oauth2.Token) {
	proxy := buildProxy(provider, scopes, clientId, clientSecret, substitutions, tokens)
	http.Handle("/", proxy)

	fmt.Printf("\nProxy server listening on :%d\n", port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil { // nosemgrep
		panic(err)
	}
}

func buildProxy(provider string, scopes []string, clientId, clientSecret string, substitutions map[string]string, tokens *oauth2.Token) *Proxy {
	providerInfo := getProviderConfig(provider, substitutions)
	cfg := configureOAuth(clientId, clientSecret, scopes, providerInfo)
	httpClient := setupHttpClient(cfg, tokens, provider, providerInfo)

	target, err := url.Parse(providerInfo.BaseURL)
	if err != nil {
		panic(err)
	}

	return newProxy(target, httpClient)
}

func getProviderConfig(provider string, substitutions map[string]string) *providers.ProviderInfo {
	config, err := providers.ReadInfo(provider, &substitutions)
	if err != nil {
		panic(fmt.Errorf("%w: %s", err, provider))
	}

	return config
}

func configureOAuth(clientId, clientSecret string, scopes []string, providerInfo *providers.ProviderInfo) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:   providerInfo.OauthOpts.AuthURL,
			TokenURL:  providerInfo.OauthOpts.TokenURL,
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}
}

// This helps with refreshing tokens automatically.
func setupHttpClient(cfg *oauth2.Config, tokens *oauth2.Token, provider string, providerInfo *providers.ProviderInfo) *http.Client {
	ctx := context.Background()

	conn, err := connector.NewConnector(
		provider,
		connector.WithClient(ctx, http.DefaultClient, cfg, tokens),
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
