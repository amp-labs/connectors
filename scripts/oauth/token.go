// nolint
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
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
//		"accessToken": "",
//		"refreshToken": ""
//	}

// Remember to run the script in the same directory as the script.
// go run token.go

const (
	HttpProtocol = "http"

	DefaultCredsFile    = "creds.json"
	DefaultServerPort   = 8080
	DefaultCallbackPath = "/callbacks/v1/oauth"
	DefaultSSLCert      = ".ssl/server.crt"
	DefaultSSLKey       = ".ssl/server.key"

	WaitBeforeExitSeconds    = 1
	ReadHeaderTimeoutSeconds = 3

	SubstitutionsFieldName = "Substitutions"
)

// ================================
// No changes required below
// ================================

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
	credscanning.Fields.State.GetJSONReader(DefaultCredsFile),
	credscanning.Fields.Username.GetJSONReader(DefaultCredsFile),
	credscanning.Fields.Password.GetJSONReader(DefaultCredsFile),
}

// OAuthApp is a simple OAuth app that can be used to get an OAuth token.
type OAuthApp struct {
	GrantType         providers.Oauth2OptsGrantType
	Callback          string
	Port              int
	Config            *oauth2.Config
	ClientCredsConfig *clientcredentials.Config
	Options           []oauth2.AuthCodeOption
	State             string
	Proto             string
	SSLCert           string
	SSLKey            string
	PasswordParams    *providers.BasicParams
}

// ServeHTTP implements the http.Handler interface.
func (a *OAuthApp) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	switch {
	case request.URL.Path == a.Callback && request.Method == "GET":
		// Process the callback.
		a.processCallback(writer, request)

	case request.URL.Path == "/" && request.Method == "GET":
		// Redirect to the OAuth provider.
		encState := base64.URLEncoding.EncodeToString([]byte(a.State))
		writer.Header().Set("Location", a.Config.AuthCodeURL(encState, a.Options...))
		writer.WriteHeader(http.StatusTemporaryRedirect)

	default:
		writer.WriteHeader(http.StatusNotFound)
	}
}

// processCallback processes the code obtained from the OAuth callback.
func (a *OAuthApp) processCallback(writer http.ResponseWriter, request *http.Request) {
	// Get the code from the query string.
	query := request.URL.Query()
	code := query.Get("code")

	// If given, get the state from the query string.
	var state string

	encState := query.Get("state")

	if encState != "" {
		stateBts, err := base64.URLEncoding.DecodeString(encState)
		if err != nil {
			slog.Error("Error base64-decoding state", "error", err)
			http.Error(writer, err.Error(), http.StatusBadRequest)

			return
		}

		state = string(stateBts)
	}

	if len(state) > 0 {
		slog.Info("got a state", "state", state)
	}

	// Exchange the code for a token.
	tok, err := a.Config.Exchange(request.Context(), code)
	if err != nil {
		slog.Error("Error exchanging code for token", "error", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	// And also in the browser
	jsonBody, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		slog.Error("Error marshalling token", "error", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	// Print the token which will also print raw metadata
	fmt.Printf("%+v", tok)

	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Content-Length", strconv.FormatInt(int64(len(jsonBody)), 10))
	writer.WriteHeader(http.StatusOK)

	// All done
	if _, err = writer.Write([]byte("Received a token, printed in the console")); err != nil { // nosemgrep
		slog.Error("Error writing token", "error", err)

		os.Exit(1)
	}

	go func() {
		time.Sleep(WaitBeforeExitSeconds * time.Second)

		os.Exit(0)
	}()
}

// Run executes the OAuth flow to get a token.
func (a *OAuthApp) Run() error {
	if a.GrantType == providers.ClientCredentials {
		src := a.ClientCredsConfig.TokenSource(context.Background())

		tok, err := src.Token()
		if err != nil {
			return err
		}

		header := tok.Type() + " " + tok.AccessToken
		fmt.Println("Expiry: " + tok.Expiry.String())
		fmt.Println("Authorization: " + header)

		return nil
	} else if a.GrantType == providers.Password {
		tok, err := a.Config.PasswordCredentialsToken(context.Background(), a.PasswordParams.User, a.PasswordParams.Pass)
		if err != nil {
			return err
		}

		header := tok.Type() + " " + tok.AccessToken
		fmt.Println("Expiry: " + tok.Expiry.String())
		fmt.Println("Authorization: " + header)
		fmt.Println("Refresh Token: " + tok.RefreshToken)

		return nil
	} else {
		slog.Info("starting OAuth app", "port", a.Port)

		http.Handle("/", a)

		go func() {
			time.Sleep(1 * time.Second)
			openBrowser(fmt.Sprintf("%s://localhost:%d", a.Proto, a.Port))
		}()

		server := &http.Server{
			Addr:              fmt.Sprintf("0.0.0.0:%d", a.Port),
			ReadHeaderTimeout: ReadHeaderTimeoutSeconds * time.Second,
		}

		if a.Proto == HttpProtocol {
			// nosemgrep: go.lang.security.audit.net.use-tls.use-tls
			return server.ListenAndServe()
		} else {
			return server.ListenAndServeTLS(a.SSLCert, a.SSLKey)
		}
	}
}

// openBrowser tries to open the URL in a browser. Should work on most standard platforms.
func openBrowser(url string) {
	slog.Info("opening browser", "url", url)

	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform: %s", runtime.GOOS) //nolint:goerr113
	}

	if err != nil {
		log.Fatal(err)
	}
}

// setup parses the CLI flags and initializes the OAuth app.
func setup() *OAuthApp {
	// Define the CLI flags.
	port := flag.Int("port", DefaultServerPort, "port to listen on")
	SSLCert := flag.String("sslcert", DefaultSSLCert, "ssl certificate")
	SSLKey := flag.String("sslkey", DefaultSSLKey, "ssl key")
	proto := flag.String("proto", HttpProtocol, "http or https protocol")

	callback := flag.String("callback", DefaultCallbackPath, "the full OAuth callback path (arbitrary)")
	flag.Parse()

	if err := registry.AddReaders(readers...); err != nil {
		return nil
	}

	substitutions, err := registry.GetMap(SubstitutionsFieldName)
	if err != nil {
		slog.Warn("no substitutions, ensure that the provider info doesn't have any variables", "error", err)
	}

	provider := registry.MustString(credscanning.Fields.Provider.Name)

	providerInfo, err := providers.ReadInfo(provider, paramsbuilder.NewCatalogVariables(substitutions)...)
	if err != nil {
		slog.Error("failed to read provider config", "error", err)

		os.Exit(1)
	}

	if providerInfo.AuthType != providers.Oauth2 {
		slog.Error("provider does not support OAuth2, not compatible with this script", "provider", provider)

		os.Exit(1)
	}

	if providerInfo.Oauth2Opts == nil {
		slog.Error("provider does not have OAuth2 options, not compatible with this script", "provider", provider)

		os.Exit(1)
	}

	// Get the OAuth scopes from the flag.
	clientId := registry.MustString(credscanning.Fields.ClientId.Name)
	clientSecret := registry.MustString(credscanning.Fields.ClientSecret.Name)

	scopes, err := registry.GetString(credscanning.Fields.Scopes.Name)
	if err != nil {
		slog.Warn("no scopes attached, ensure that the provider doesn't require scopes")
	}

	oauthScopes := strings.Split(scopes, ",")

	switch providerInfo.Oauth2Opts.GrantType {
	case providers.AuthorizationCode:
		if providerInfo.Oauth2Opts.AuthURL == "" {
			slog.Error("provider does not have an AuthURL, not compatible with this script", "provider", provider)

			os.Exit(1)
		}

		// Determine the OAuth redirect URL.
		redirect := fmt.Sprintf("%s://localhost:%d%s", *proto, *port, *callback)

		state, err := registry.GetString(credscanning.Fields.State.Name)
		if err != nil {
			slog.Warn("no state attached, ensure that the provider doesn't require state")
		}

		// Create the OAuth app.
		app := &OAuthApp{
			GrantType: providers.AuthorizationCode,
			Callback:  *callback,
			Port:      *port,
			Proto:     *proto,
			SSLCert:   *SSLCert,
			SSLKey:    *SSLKey,
			Config: &oauth2.Config{
				ClientID:     clientId,
				ClientSecret: clientSecret,
				RedirectURL:  redirect,
				Scopes:       oauthScopes,
			},
		}
		if state != "" {
			app.State = state
		}

		// Set up the OAuth config based on the provider.
		app.Config.Endpoint = oauth2.Endpoint{
			AuthURL:   providerInfo.Oauth2Opts.AuthURL,
			TokenURL:  providerInfo.Oauth2Opts.TokenURL,
			AuthStyle: oauth2.AuthStyleAutoDetect,
		}

		var authCodeOptions []oauth2.AuthCodeOption

		authURLParams := providerInfo.Oauth2Opts.AuthURLParams
		if authURLParams != nil {
			for k, v := range authURLParams {
				option := oauth2.SetAuthURLParam(k, v)
				app.Options = append(authCodeOptions, option)
			}
		}

		return app
	case providers.ClientCredentials:
		state, err := registry.GetString(credscanning.Fields.State.Name)
		if err != nil {
			slog.Warn("no state attached, ensure that the provider doesn't require state")
		}

		// Create the OAuth app.
		app := &OAuthApp{
			GrantType: providers.ClientCredentials,
			ClientCredsConfig: &clientcredentials.Config{
				ClientID:     clientId,
				ClientSecret: clientSecret,
				TokenURL:     providerInfo.Oauth2Opts.TokenURL,
				Scopes:       oauthScopes,
				AuthStyle:    oauth2.AuthStyleAutoDetect,
			},
		}
		if state != "" {
			app.State = state
		}

		if providerInfo.Oauth2Opts.Audience != nil {
			aud := providerInfo.Oauth2Opts.Audience
			app.ClientCredsConfig.EndpointParams = url.Values{"audience": aud}
		}

		return app
	case providers.Password:
		state, err := registry.GetString(credscanning.Fields.State.Name)
		if err != nil {
			slog.Warn("no state attached, ensure that the provider doesn't require state")
		}

		username := registry.MustString(credscanning.Fields.Username.Name)
		password := registry.MustString(credscanning.Fields.Password.Name)

		app := &OAuthApp{
			GrantType: providers.Password,
			Config: &oauth2.Config{
				ClientID:     clientId,
				ClientSecret: clientSecret,
				Scopes:       oauthScopes,
				Endpoint: oauth2.Endpoint{
					TokenURL: providerInfo.Oauth2Opts.TokenURL,
				},
			},
			PasswordParams: &providers.BasicParams{
				User: username,
				Pass: password,
			},
		}
		if state != "" {
			app.State = state
		}

		if providerInfo.Oauth2Opts.Audience != nil {
			aud := providerInfo.Oauth2Opts.Audience
			app.ClientCredsConfig.EndpointParams = url.Values{"audience": aud}
		}

		return app
	default:
		slog.Error("provider does not support authorization code or client credentials grant, not compatible with this script", "provider", provider)

		os.Exit(1)
	}

	return nil
}

func main() {
	// Parse flags and set up the OAuth app.
	app := setup()

	// Run the OAuth app.
	if err := app.Run(); err != nil {
		slog.Error("failed to run OAuth app", "error", err)

		time.Sleep(WaitBeforeExitSeconds * time.Second)

		os.Exit(1)
	}
}
