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
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/connector"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/utils"
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
	&utils.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['apiKey']",
		CredKey:  "ApiKey",
	},
	&utils.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['userName']",
		CredKey:  "UserName",
	},
	&utils.JSONReader{
		FilePath: DefaultCredsFile,
		JSONPath: "$['password']",
		CredKey:  "Password",
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

	// Cast the substitutions to a map[string]string
	substitutionsMap := make(map[string]string)
	for key, val := range substitutions {
		substitutionsMap[key] = val.MustString()
	}

	info, err := providers.ReadInfo(provider, &substitutionsMap)
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
			mainOAuth2ClientCreds(ctx, provider, substitutionsMap)
		case providers.AuthorizationCode:
			mainOAuth2AuthCode(ctx, provider, substitutionsMap)
		case providers.Password:
			mainOAuth2PasswordCreds(ctx, provider, substitutionsMap)
		default:
			log.Fatalf("Unsupported OAuth2 grant type: %s", info.Oauth2Opts.GrantType)
		}
	case providers.ApiKey:
		mainApiKey(ctx, provider, substitutionsMap)
	case providers.Basic:
		mainBasic(ctx, provider, substitutionsMap)
	default:
		log.Fatalf("Unsupported auth type: %s", info.AuthType)
	}
}

func mainOAuth2ClientCreds(ctx context.Context, provider string, substitutions map[string]string) {
	params := createClientAuthParams(provider)
	proxy := buildOAuth2ClientCredentialsProxy(ctx, provider, params.Scopes, params.ID, params.Secret, substitutions)
	startProxy(ctx, proxy, DefaultPort)
}

func mainOAuth2AuthCode(ctx context.Context, provider string, substitutions map[string]string) {
	params := createClientAuthParams(provider)
	tokens := getTokensFromRegistry()
	proxy := buildOAuth2AuthCodeProxy(ctx, provider, params.Scopes, params.ID, params.Secret, substitutions, tokens)
	startProxy(ctx, proxy, DefaultPort)
}

func mainOAuth2PasswordCreds(ctx context.Context, provider string, substitutionsMap map[string]string) {
	authParams := createClientAuthParams(provider)
	basicParams := createBasicParams()
	proxy := buildOAuth2PasswordProxy(ctx, provider, authParams, substitutionsMap, basicParams)
	startProxy(ctx, proxy, DefaultPort)
}

func mainApiKey(ctx context.Context, provider string, substitutions map[string]string) {
	apiKey := registry.MustString("ApiKey")
	if apiKey == "" {
		_, _ = fmt.Fprintln(os.Stderr, "api key from registry is empty")
		os.Exit(1)
	}

	proxy := buildApiKeyProxy(ctx, provider, substitutions, apiKey)
	startProxy(ctx, proxy, DefaultPort)
}

func mainBasic(ctx context.Context, provider string, substitutions map[string]string) {
	params := createBasicParams()

	proxy := buildBasicAuthProxy(ctx, provider, substitutions, params.User, params.Pass)
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

func buildOAuth2ClientCredentialsProxy(ctx context.Context, provider string, scopes []string, clientId, clientSecret string, substitutions map[string]string) *Proxy {
	providerInfo := getProviderConfig(provider, substitutions)
	cfg := configureOAuthClientCredentials(clientId, clientSecret, scopes, providerInfo)
	httpClient := setupOAuth2ClientCredentialsHttpClient(ctx, providerInfo, cfg)

	target, err := url.Parse(providerInfo.BaseURL)
	if err != nil {
		panic(err)
	}

	return newProxy(target, httpClient)
}

func buildApiKeyProxy(ctx context.Context, provider string, substitutions map[string]string, apiKey string) *Proxy {
	providerInfo := getProviderConfig(provider, substitutions)
	httpClient := setupApiKeyHttpClient(ctx, providerInfo, apiKey)

	target, err := url.Parse(providerInfo.BaseURL)
	if err != nil {
		panic(err)
	}

	return newProxy(target, httpClient)
}

func buildBasicAuthProxy(ctx context.Context, provider string, substitutions map[string]string, user, pass string) *Proxy {
	providerInfo := getProviderConfig(provider, substitutions)
	httpClient := setupBasicAuthHttpClient(ctx, providerInfo, user, pass)

	target, err := url.Parse(providerInfo.BaseURL)
	if err != nil {
		panic(err)
	}

	return newProxy(target, httpClient)
}

func buildOAuth2AuthCodeProxy(ctx context.Context, provider string, scopes []string, clientId, clientSecret string, substitutions map[string]string, tokens *oauth2.Token) *Proxy {
	providerInfo := getProviderConfig(provider, substitutions)
	cfg := configureOAuthAuthCode(clientId, clientSecret, scopes, providerInfo)
	httpClient := setupOAuth2AuthCodeHttpClient(ctx, providerInfo, cfg, tokens)

	target, err := url.Parse(providerInfo.BaseURL)
	if err != nil {
		panic(err)
	}

	return newProxy(target, httpClient)
}

func buildOAuth2PasswordProxy(ctx context.Context, provider string, authParams *ClientAuthParams, substitutions map[string]string, params *providers.BasicParams) *Proxy {
	providerInfo := getProviderConfig(provider, substitutions)
	cfg := configureOAuthAuthCode(authParams.ID, authParams.Secret, authParams.Scopes, providerInfo)
	httpClient := setupOAuth2PasswordHttpClient(ctx, providerInfo, cfg, params)

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

func configureOAuthClientCredentials(clientId, clientSecret string, scopes []string, providerInfo *providers.ProviderInfo) *clientcredentials.Config {
	cfg := &clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		TokenURL:     providerInfo.Oauth2Opts.TokenURL,
	}

	if providerInfo.Oauth2Opts.ExplicitScopesRequired {
		cfg.Scopes = scopes
	}

	if providerInfo.Oauth2Opts.Audience != "" {
		aud := providerInfo.Oauth2Opts.Audience
		cfg.EndpointParams = url.Values{"audience": {aud}}
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
		Debug:             *debug,
		OAuth2ClientCreds: cfg,
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

func setupOAuth2PasswordHttpClient(ctx context.Context, prov *providers.ProviderInfo, cfg *oauth2.Config, params *providers.BasicParams) common.AuthenticatedHTTPClient {
	c, err := prov.NewClient(ctx, &providers.NewClientParams{
		Debug: *debug,
		Oauth2PasswordParams: &providers.Oauth2PasswordParams{
			Config:      cfg,
			BasicParams: params,
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
