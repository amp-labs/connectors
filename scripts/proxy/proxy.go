package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/amp-labs/connectors/basic"
	"github.com/amp-labs/connectors/providers"
	"golang.org/x/oauth2"
)

// ================================
// Example usage
// ================================

//	go run proxy.go -provider salesforce \
//    --client-id=************** \
//    --client-secret=**************
//    --substitutions=workspace=demo

// go run proxy.go -provider hubspot \
//    --client-id=************** \
//    --client-secret=**************
//    --substitutions=workspace=demo
//    --scopes=crm.objects.contacts.read,crm.objects.contacts.write

// go run proxy.go -provider linkedIn \
//    --client-id=************** \
//    --client-secret=**************
//    --scopes=openid,profile,email

// ==============================
// Configuration
// ==============================

var (
	defaultTokenPath = ".amp-provider-token.json"
	defaultPort      = 4444
)

// ==============================
// Main (no changes needed)
// ==============================

func main() {
	provider, clientId, clientSecret, scopes, substitutions, tokenPath, port := parseFlags()
	validateRequiredFlags(provider, clientId, clientSecret)
	startProxy(provider, scopes, clientId, clientSecret, substitutions, tokenPath, port)
}

func parseFlags() (provider, clientId, clientSecret, scopes, substitutions, tokenPath string, port int) {
	flag.StringVar(&provider, "provider", "", "[required] the name of the provider")
	flag.StringVar(&clientId, "client-id", "", "[required] provider app client id")
	flag.StringVar(&clientSecret, "client-secret", "", "[required] provider app client secret")
	flag.StringVar(&scopes, "scopes", "", "[optional] the scopes to request (comma separated)")
	flag.StringVar(&tokenPath, "token-path", defaultTokenPath, "[optional] path to the token file")
	flag.IntVar(&port, "port", defaultPort, "[optional] the port to start the proxy on")
	flag.StringVar(&substitutions, "substitutions", "", "the substitutions to use (comma separated as key=value)")
	flag.Parse()
	return
}

func validateRequiredFlags(provider, clientId, clientSecret string) {
	if provider == "" || clientId == "" || clientSecret == "" {
		fmt.Fprintln(os.Stderr, "Missing required flags: -provider, -client-id, -client-secret")
		flag.Usage()
		os.Exit(1)
	}
}

func startProxy(provider, scopes, clientId, clientSecret, substitutions, tokenPath string, port int) {
	proxy := buildProxy(provider, scopes, clientId, clientSecret, substitutions, tokenPath)
	http.Handle("/", proxy)

	fmt.Printf("\nProxy server listening on :%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}

func buildProxy(provider, scopes, clientId, clientSecret, substitutions, tokenPath string) *Proxy {
	providerInfo := getProviderConfig(provider, substitutions)
	token := readToken(tokenPath)
	cfg := configureOAuth(clientId, clientSecret, scopes, providerInfo)
	httpClient := setupHttpClient(cfg, token, provider, providerInfo)

	target, err := url.Parse(providerInfo.BaseURL)
	if err != nil {
		panic(err)
	}

	return newProxy(target, httpClient)
}

func getProviderConfig(provider, substitutions string) *providers.ProviderInfo {
	var sMap map[string]string

	if substitutions != "" {
		sMap = convertSubstitutions(substitutions)
	}

	config, err := providers.ReadConfig(providers.Provider(provider), &sMap)
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

func configureOAuth(clientId, clientSecret, scopes string, providerInfo *providers.ProviderInfo) *oauth2.Config {
	scopesSlice := strings.Split(scopes, ",")
	if scopes == "" {
		scopesSlice = nil
	}

	return &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Scopes:       scopesSlice,
		Endpoint: oauth2.Endpoint{
			AuthURL:   providerInfo.OauthOpts.AuthURL,
			TokenURL:  providerInfo.OauthOpts.TokenURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

// This helps with refreshing tokens automatically.
func setupHttpClient(cfg *oauth2.Config, token *oauth2.Token, provider string, providerInfo *providers.ProviderInfo) *http.Client {
	ctx := context.Background()
	conn, err := basic.NewConnector(
		providers.Provider(provider),
		basic.WithClient(ctx, http.DefaultClient, cfg, token),
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
