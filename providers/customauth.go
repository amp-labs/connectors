package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// AuthContext carries data through a multi-step custom auth flow. It is pure,
// JSON-serializable data: the server persists it between steps (e.g. across a
// browser redirect) and, when the flow completes, into the connection. Handlers
// read the inputs and write acquired credentials/metadata into Secrets/Metadata.
type AuthContext struct {
	// ConsumerInputs are what the consumer submitted (username, password, ...).
	// Set once at flow start; not modified by steps.
	ConsumerInputs map[string]string `json:"consumerInputs,omitempty"`

	// ProviderInputs are what the builder configured on their provider app
	// (clientId, clientSecret, subscription keys). Set once at flow start.
	ProviderInputs map[string]string `json:"providerInputs,omitempty"`

	// Secrets are credentials accumulated by steps (accessToken, sessionId, ...).
	// Persisted to the connection (encrypted) when the flow completes.
	Secrets map[string]string `json:"secrets,omitempty"`

	// Metadata is non-sensitive config accumulated by steps (instanceUrl,
	// workspace, ...). Persisted to the connection when the flow completes.
	Metadata map[string]string `json:"metadata,omitempty"`

	// System is server-injected, environment-specific, read-only config
	// (e.g. callbackURL). Repopulated by the server on every step; never persisted.
	System map[string]string `json:"-"`
}

// NewAuthContext returns an AuthContext with every sub-map initialized, so
// handlers can write without nil checks.
func NewAuthContext() AuthContext {
	return AuthContext{
		ConsumerInputs: map[string]string{},
		ProviderInputs: map[string]string{},
		Secrets:        map[string]string{},
		Metadata:       map[string]string{},
		System:         map[string]string{},
	}
}

// Flatten merges every sub-map into one for template resolution. Precedence,
// lowest to highest: System, Metadata, ProviderInputs, ConsumerInputs, Secrets.
func (c AuthContext) Flatten() map[string]string {
	out := make(map[string]string)
	for _, m := range []map[string]string{c.System, c.Metadata, c.ProviderInputs, c.ConsumerInputs, c.Secrets} {
		for k, v := range m {
			out[k] = v
		}
	}

	return out
}

// HTTPStep is a server-side HTTP call that acquires or refreshes credentials.
// BuildRequest constructs the full request to send; ParseResponse extracts
// results. State is passed by value and returned explicitly (no pointer mutation).
type HTTPStep struct {
	// BuildRequest builds the request to send. Required.
	BuildRequest func(ctx context.Context, state AuthContext) (AuthContext, *http.Request, error)

	// ParseResponse extracts values from the response into the returned state. Required.
	ParseResponse func(ctx context.Context, state AuthContext, resp *http.Response) (AuthContext, error)
}

// RedirectStep sends the user's browser to a URL and resumes when the provider
// redirects back to the Ampersand callback.
type RedirectStep struct {
	// TimeoutSeconds bounds how long to wait for the callback. 0 = server default.
	TimeoutSeconds int

	// BuildURL returns the URL to redirect the browser to. Required.
	BuildURL func(ctx context.Context, state AuthContext) (AuthContext, string, error)

	// ParseCallback extracts values from the provider's callback request. Required.
	ParseCallback func(ctx context.Context, state AuthContext, callback *http.Request) (AuthContext, error)
}

// AuthStep is one step of a flow. Exactly one of HTTP or Redirect is set.
type AuthStep struct {
	HTTP     *HTTPStep
	Redirect *RedirectStep
}

// CustomAuthFlow is the executable definition backing a provider's declarative
// CustomAuthOpts.MultiStep flag: the handlers the server runs to acquire and
// refresh credentials. RefreshSteps are HTTP-only, since refresh is
// non-interactive (no browser redirects).
type CustomAuthFlow struct {
	ConnectSteps []AuthStep
	RefreshSteps []HTTPStep
}

// HasRedirect reports whether any connect step is a browser redirect, so the
// server knows to hand a URL back to the client rather than finishing inline.
func (f CustomAuthFlow) HasRedirect() bool {
	for _, s := range f.ConnectSteps {
		if s.Redirect != nil {
			return true
		}
	}

	return false
}

// customAuthFlows holds the executable flows, keyed by provider. Populated by
// provider init() via RegisterCustomAuthFlow; never serialized. This is why the
// catalog only needs the MultiStep flag: the steps live here, compiled in.
var customAuthFlows = map[Provider]CustomAuthFlow{} //nolint:gochecknoglobals

// RegisterCustomAuthFlow records the executable step handlers for a provider
// whose ProviderInfo has CustomOpts.MultiStep set. Called from provider init().
func RegisterCustomAuthFlow(provider Provider, flow CustomAuthFlow) {
	customAuthFlows[provider] = flow
}

// CustomAuthFlowFor returns the registered flow for a provider, if any.
func CustomAuthFlowFor(provider Provider) (CustomAuthFlow, bool) {
	f, ok := customAuthFlows[provider]

	return f, ok
}

// HasSteps reports whether this provider uses a multi-step custom auth flow
// (driven via /custom-auth/connect) rather than static header/query injection.
func (o *CustomAuthOpts) HasSteps() bool {
	return o != nil && o.MultiStep
}

// ExtractJSONSecrets returns a ParseResponse handler that decodes the JSON body
// and copies mapped response fields (responseKey -> secretKey) into Secrets.
func ExtractJSONSecrets(
	mapping map[string]string,
) func(context.Context, AuthContext, *http.Response) (AuthContext, error) {
	return func(_ context.Context, state AuthContext, resp *http.Response) (AuthContext, error) {
		defer resp.Body.Close()

		var body map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return state, fmt.Errorf("decoding JSON response: %w", err)
		}

		for responseKey, secretKey := range mapping {
			if v, ok := body[responseKey].(string); ok {
				state.Secrets[secretKey] = v
			}
		}

		return state, nil
	}
}

// ExtractQueryParamsSecrets returns a ParseCallback handler that copies mapped
// callback query params (paramName -> secretKey) into Secrets.
func ExtractQueryParamsSecrets(
	mapping map[string]string,
) func(context.Context, AuthContext, *http.Request) (AuthContext, error) {
	return func(_ context.Context, state AuthContext, callback *http.Request) (AuthContext, error) {
		query := callback.URL.Query()
		for param, secretKey := range mapping {
			if v := query.Get(param); v != "" {
				state.Secrets[secretKey] = v
			}
		}

		return state, nil
	}
}
