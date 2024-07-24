// nolint
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/connector"
	"github.com/amp-labs/connectors/providers"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
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

var registry = scanning.NewRegistry()

var readers = []scanning.Reader{
	&scanning.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['clientId']",
		KeyName:  "ClientId",
	},
	&scanning.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['clientSecret']",
		KeyName:  "ClientSecret",
	},
	&scanning.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['scopes']",
		KeyName:  "Scopes",
	},
	&scanning.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['provider']",
		KeyName:  "Provider",
	},
	&scanning.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['substitutions']",
		KeyName:  "Substitutions",
	},
	&scanning.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['accessToken']",
		KeyName:  "AccessToken",
	},
	&scanning.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['refreshToken']",
		KeyName:  "RefreshToken",
	},
	&scanning.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['expiry']",
		KeyName:  "Expiry",
	},
	&scanning.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['expiryFormat']",
		KeyName:  "ExpiryFormat",
	},
	&scanning.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['apiKey']",
		KeyName:  "ApiKey",
	},
	&scanning.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['userName']",
		KeyName:  "UserName",
	},
	&scanning.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['password']",
		KeyName:  "Password",
	},
}

var debug = flag.Bool("debug", false, "Enable debug logging")

func main() {
	flag.Parse()

	err := registry.AddReaders(readers...)
	if err != nil {
		panic(err)
	}

	provider := registry.MustString("Provider")

	substitutions, err := registry.GetMap("Substitutions")
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

	switch info.AuthType {
	case providers.Oauth2:
		if info.Oauth2Opts == nil {
			log.Fatalf("Missing OAuth options for provider %s", provider)
		}

		switch info.Oauth2Opts.GrantType {
		case providers.ClientCredentials:
			mainOAuth2ClientCreds(ctx, provider, catalogVariables)
		case providers.AuthorizationCode:
			mainOAuth2AuthCode(ctx, provider, catalogVariables)
		case providers.Password:
			// de facto, password grant acts as client credentials,
			// even so access and refresh tokens were acquired differently.
			mainOAuth2ClientCreds(ctx, provider, catalogVariables)
		default:
			log.Fatalf("Unsupported OAuth2 grant type: %s", info.Oauth2Opts.GrantType)
		}
	case providers.ApiKey:
		mainApiKey(ctx, provider, catalogVariables)
	case providers.Basic:
		mainBasic(ctx, provider, catalogVariables)
	default:
		log.Fatalf("Unsupported auth type: %s", info.AuthType)
	}
}

func mainOAuth2ClientCreds(ctx context.Context, provider string, catalogVariables []paramsbuilder.CatalogVariable) {
	params := createClientAuthParams(provider)
	tokens := getTokensFromRegistry()
	proxy := buildOAuth2AuthCodeProxy(ctx, provider, params.Scopes, params.ID, params.Secret, catalogVariables, tokens)
	startProxy(ctx, proxy, DefaultPort)
}

func mainOAuth2AuthCode(ctx context.Context, provider string, catalogVariables []paramsbuilder.CatalogVariable) {
	params := createClientAuthParams(provider)
	tokens := getTokensFromRegistry()
	proxy := buildOAuth2AuthCodeProxy(ctx, provider, params.Scopes, params.ID, params.Secret, catalogVariables, tokens)
	startProxy(ctx, proxy, DefaultPort)
}

func mainApiKey(ctx context.Context, provider string, catalogVariables []paramsbuilder.CatalogVariable) {
	apiKey := registry.MustString("ApiKey")
	if apiKey == "" {
		_, _ = fmt.Fprintln(os.Stderr, "api key from registry is empty")
		os.Exit(1)
	}

	proxy := buildApiKeyProxy(ctx, provider, catalogVariables, apiKey)
	startProxy(ctx, proxy, DefaultPort)
}

func mainBasic(ctx context.Context, provider string, catalogVariables []paramsbuilder.CatalogVariable) {
	params := createBasicParams()

	proxy := buildBasicAuthProxy(ctx, provider, catalogVariables, params.User, params.Pass)
	startProxy(ctx, proxy, DefaultPort)
}

func createBasicParams() *providers.BasicParams {
	user := registry.MustString("UserName")
	pass := registry.MustString("Password")

	if len(user)+len(pass) == 0 {
		log.Fatalf("Missing username or password")
	}

	if len(user) == 0 {
		slog.Warn("no username for basic authentication, ensure that it is not required")
	}

	if len(pass) == 0 {
		slog.Warn("no password for basic authentication, ensure that it is not required")
	}

	return &providers.BasicParams{
		User: user,
		Pass: pass,
	}
}

func createClientAuthParams(provider string) *ClientAuthParams {
	clientId := registry.MustString("ClientId")
	clientSecret := registry.MustString("ClientSecret")

	scopes, err := registry.GetString("Scopes")
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

func getTokensFromRegistry() *oauth2.Token {
	reader, err := credscanning.NewJSONProviderCredentials(DefaultCredsFile, true)
	if err != nil {
		panic(err)
	}

	return reader.GetOauthToken()
}

func validateRequiredOAuth2Flags(provider, clientId, clientSecret string) {
	if provider == "" || clientId == "" || clientSecret == "" {
		_, _ = fmt.Fprintln(os.Stderr, "Missing required flags: -provider, -client-id, -client-secret")

		flag.Usage()
		os.Exit(1)
	}
}

// listen will start a server on the given port and block until it is closed.
// This is used as opposed to http.ListenAndServe because it respects the context
// and has a cleaner shutdown sequence.
func listen(ctx context.Context, port int) error {
	var lc net.ListenConfig

	listener, err := lc.Listen(ctx, "tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	go func() {
		<-ctx.Done()

		_ = server.Shutdown(context.Background())
	}()

	if err := server.Serve(listener); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Println("HTTP server stopped")

			return nil
		}

		return err
	}

	return nil
}

func startProxy(ctx context.Context, proxy *Proxy, port int) {
	http.Handle("/", proxy)

	fmt.Printf("\nProxy server listening on :%d\n", port)

	if err := listen(ctx, port); err != nil {
		panic(err)
	}
}

func buildOAuth2ClientCredentialsProxy(ctx context.Context, provider string, scopes []string, clientId, clientSecret string, catalogVariables []paramsbuilder.CatalogVariable) *Proxy {
	providerInfo := getProviderConfig(provider, catalogVariables)
	cfg := configureOAuthClientCredentials(clientId, clientSecret, scopes, providerInfo)
	httpClient := setupOAuth2ClientCredentialsHttpClient(ctx, providerInfo, cfg)

	target, err := url.Parse(providerInfo.BaseURL)
	if err != nil {
		panic(err)
	}

	return newProxy(target, httpClient)
}

func buildApiKeyProxy(ctx context.Context, provider string, catalogVariables []paramsbuilder.CatalogVariable, apiKey string) *Proxy {
	providerInfo := getProviderConfig(provider, catalogVariables)
	httpClient := setupApiKeyHttpClient(ctx, providerInfo, apiKey)

	target, err := url.Parse(providerInfo.BaseURL)
	if err != nil {
		panic(err)
	}

	return newProxy(target, httpClient)
}

func buildBasicAuthProxy(ctx context.Context, provider string, catalogVariables []paramsbuilder.CatalogVariable, user, pass string) *Proxy {
	providerInfo := getProviderConfig(provider, catalogVariables)
	httpClient := setupBasicAuthHttpClient(ctx, providerInfo, user, pass)

	target, err := url.Parse(providerInfo.BaseURL)
	if err != nil {
		panic(err)
	}

	return newProxy(target, httpClient)
}

func buildOAuth2AuthCodeProxy(ctx context.Context, provider string, scopes []string, clientId, clientSecret string, catalogVariables []paramsbuilder.CatalogVariable, tokens *oauth2.Token) *Proxy {
	providerInfo := getProviderConfig(provider, catalogVariables)
	cfg := configureOAuthAuthCode(clientId, clientSecret, scopes, providerInfo)
	httpClient := setupOAuth2AuthCodeHttpClient(ctx, providerInfo, cfg, tokens)

	target, err := url.Parse(providerInfo.BaseURL)
	if err != nil {
		panic(err)
	}

	return newProxy(target, httpClient)
}

func getProviderConfig(provider string, catalogVariables []paramsbuilder.CatalogVariable) *providers.ProviderInfo {
	config, err := providers.ReadInfo(provider, catalogVariables...)
	if err != nil {
		panic(fmt.Errorf("%w: %s", err, provider))
	}

	return config
}

func configureOAuthClientCredentials(clientId, clientSecret string, scopes []string, providerInfo *providers.ProviderInfo) *clientcredentials.Config {
	cfg := &clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		TokenURL:     providerInfo.Oauth2Opts.TokenURL,
	}

	if providerInfo.Oauth2Opts.ExplicitScopesRequired {
		cfg.Scopes = scopes
	}

	if providerInfo.Oauth2Opts.Audience != nil {
		aud := providerInfo.Oauth2Opts.Audience
		cfg.EndpointParams = url.Values{"audience": aud}
	}

	return cfg
}

type ClientAuthParams struct {
	ID     string
	Secret string
	Scopes []string
}

func configureOAuthAuthCode(clientId, clientSecret string, scopes []string, providerInfo *providers.ProviderInfo) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:   providerInfo.Oauth2Opts.AuthURL,
			TokenURL:  providerInfo.Oauth2Opts.TokenURL,
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}
}

func setupOAuth2ClientCredentialsHttpClient(ctx context.Context, prov *providers.ProviderInfo, cfg *clientcredentials.Config) common.AuthenticatedHTTPClient {
	c, err := prov.NewClient(ctx, &providers.NewClientParams{
		Debug: *debug,
		OAuth2ClientCreds: &providers.OAuth2ClientCredentialsParams{
			Config: cfg,
		},
	})
	if err != nil {
		panic(err)
	}

	cc, err := connector.NewConnector(prov.Name, connector.WithAuthenticatedClient(c))
	if err != nil {
		panic(err)
	}

	return cc.HTTPClient().Client
}

// This helps with refreshing tokens automatically.
func setupOAuth2AuthCodeHttpClient(ctx context.Context, prov *providers.ProviderInfo, cfg *oauth2.Config, tokens *oauth2.Token) common.AuthenticatedHTTPClient {
	c, err := prov.NewClient(ctx, &providers.NewClientParams{
		Debug: *debug,
		OAuth2AuthCodeCreds: &providers.OAuth2AuthCodeParams{
			Config: cfg,
			Token:  tokens,
		},
	})
	if err != nil {
		panic(err)
	}

	cc, err := connector.NewConnector(prov.Name, connector.WithAuthenticatedClient(c))
	if err != nil {
		panic(err)
	}

	return cc.HTTPClient().Client
}

func setupBasicAuthHttpClient(ctx context.Context, prov *providers.ProviderInfo, user, pass string) common.AuthenticatedHTTPClient {
	c, err := prov.NewClient(ctx, &providers.NewClientParams{
		Debug: *debug,
		BasicCreds: &providers.BasicParams{
			User: user,
			Pass: pass,
		},
	})
	if err != nil {
		panic(err)
	}

	cc, err := connector.NewConnector(prov.Name, connector.WithAuthenticatedClient(c))
	if err != nil {
		panic(err)
	}

	return cc.HTTPClient().Client
}

func setupApiKeyHttpClient(ctx context.Context, prov *providers.ProviderInfo, apiKey string) common.AuthenticatedHTTPClient {
	c, err := prov.NewClient(ctx, &providers.NewClientParams{
		Debug:  *debug,
		ApiKey: apiKey,
	})
	if err != nil {
		panic(err)
	}

	cc, err := connector.NewConnector(prov.Name, connector.WithAuthenticatedClient(c))
	if err != nil {
		panic(err)
	}

	return cc.HTTPClient().Client
}

type Proxy struct {
	*httputil.ReverseProxy
	target *url.URL
}

func newProxy(target *url.URL, httpClient common.AuthenticatedHTTPClient) *Proxy {
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
	httpClient common.AuthenticatedHTTPClient
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.httpClient.Do(req)
}
