package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/providers"
	"golang.org/x/oauth2"
)

// ================================
// Example usage
// ================================

//	go run token.go -provider salesforce \
//	  --client-id=************** \
//	  --client-secret=************** \
//	  --substitutions=workspace=example

// go run token.go -provider hubspot \
//	  --client-id=************** \
//	  --client-secret=************** \
// 	  --scopes=crm.objects.contacts.read,crm.objects.companies.read

// go run token.go -provider linkedIn \
//      --client-id=************** \
//      --client-secret=**************
//      --scopes=openid,profile,email

// ===============================
// Variables required for testing
// ===============================

const (
	// TokenPath is the path to the token file, and it is relative to the current working directory.
	TokenPath = ".amp-provider-token.json"

	DefaultServerPort   = 8080
	DefaultCallbackPath = "/callbacks/v1/oauth"

	WaitBeforeExitSeconds    = 3
	ReadHeaderTimeoutSeconds = 3
)

// ===============================
// End of variables required for testing
// ===============================

// OAuthApp is a simple OAuth app that can be used to get an OAuth token.
type OAuthApp struct {
	Callback string
	Port     int
	Config   *oauth2.Config
	Options  []oauth2.AuthCodeOption
	State    string
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
		url := a.Config.AuthCodeURL(encState, a.Options...)
		writer.Header().Set("Location", url)
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

			removeTokenFile()

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

		removeTokenFile()

		return
	}

	// And also in the browser
	jsonBody, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		slog.Error("Error marshalling token", "error", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		removeTokenFile()

		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Content-Length", strconv.FormatInt(int64(len(jsonBody)), 10))
	writer.WriteHeader(http.StatusOK)

	err = createTokenFile(tok)
	if err != nil {
		slog.Error("Error writing token to file", "error", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		removeTokenFile()

		return
	}

	// All done
	if _, err = writer.Write([]byte("Successfully created token file (" + TokenPath + ")")); err != nil {
		slog.Error("Error writing token", "error", err)

		removeTokenFile()
	}

	slog.Info("token written to file", "path", TokenPath)

	go func() {
		time.Sleep(WaitBeforeExitSeconds * time.Second)

		os.Exit(0)
	}()
}

// Run executes the OAuth flow to get a token.
func (a *OAuthApp) Run() error {
	slog.Info("starting OAuth app", "port", a.Port)

	http.Handle("/", a)

	go func() {
		time.Sleep(1 * time.Second)
		openBrowser(fmt.Sprintf("http://localhost:%d", a.Port))
	}()

	server := &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", a.Port),
		ReadHeaderTimeout: ReadHeaderTimeoutSeconds * time.Second,
	}

	// nosemgrep: go.lang.security.audit.net.use-tls.use-tls
	return server.ListenAndServe()
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

// setup parses the CLI flags and initializes the OAuth app.
func setup() *OAuthApp {
	// Define the CLI flags.
	port := flag.Int("port", DefaultServerPort, "port to listen on")
	client := flag.String("client-id", "", "an OAuth client id")
	secret := flag.String("client-secret", "", "an OAuth client secret")
	provider := flag.String("provider", "", "the type of OAuth provider")
	scopes := flag.String("scopes", "", "the scopes to request (comma separated)")
	callback := flag.String("callback", DefaultCallbackPath, "the full OAuth callback path (arbitrary)")
	substitutions := flag.String("substitutions", "", "the substitutions to use (comma separated as key=value)")
	flag.Parse()

	// Make sure the required flags are set
	sanityCheckFlags(client, secret, provider)

	// Determine the OAuth redirect URL.
	redirect := fmt.Sprintf("http://localhost:%d%s", *port, *callback)

	// Optionally set the OAuth options based on the flags. Most users won't care about this.
	var opts []oauth2.AuthCodeOption

	// Get the OAuth scopes from the flag.
	oauthScopes := getScopes(*provider, scopes)

	// Create the OAuth app.
	app := &OAuthApp{
		Callback: *callback,
		Port:     *port,
		Options:  opts,
		Config: &oauth2.Config{
			ClientID:     *client,
			ClientSecret: *secret,
			RedirectURL:  redirect,
			Scopes:       oauthScopes,
		},
	}

	// Convert substitutions to a map
	substitutionsMap := convertSubstitutions(*substitutions)

	providerInfo, err := providers.ReadConfig(providers.Provider(*provider), &substitutionsMap)
	if err != nil {
		slog.Error("failed to read provider config", "error", err)

		os.Exit(1)
	}

	// Set up the OAuth config based on the provider.
	app.Config.Endpoint = oauth2.Endpoint{
		AuthURL:   providerInfo.OauthOpts.AuthURL,
		TokenURL:  providerInfo.OauthOpts.TokenURL,
		AuthStyle: oauth2.AuthStyleInParams,
	}

	return app
}

func sanityCheckFlags(client *string, secret *string, provider *string) {
	if *client == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Missing required flag: -client-id\n")

		flag.Usage()

		os.Exit(1)
	}

	if *secret == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Missing required flag: -client-secret\n")

		flag.Usage()

		os.Exit(1)
	}

	if *provider == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Missing required flag: -provider\n")

		flag.Usage()

		os.Exit(1)
	}
}

func getScopes(provider string, scopesFlag *string) []string {
	var scopes []string

	parts := strings.Split(*scopesFlag, ",")

	for _, part := range parts {
		v := strings.TrimSpace(part)
		if len(v) > 0 {
			scopes = append(scopes, v)
		}
	}

	if len(scopes) == 0 && provider == "hubspot" {
		slog.Error("no scopes provided for hubspot")
		os.Exit(1)
	}

	return scopes
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

func createTokenFile(token *oauth2.Token) error {
	file, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(TokenPath, file, 0600)
	if err != nil {
		return err
	}

	return nil
}

// removeTokenFile removes the token file if it exists.
func removeTokenFile() {
	if _, err := os.Stat(TokenPath); err == nil {
		// File exists, remove it
		if err := os.Remove(TokenPath); err != nil {
			slog.Error("Error removing token file", "error", err)
		}
	}
}
