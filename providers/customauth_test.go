package providers

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestAuthContextFlattenPrecedence(t *testing.T) {
	t.Parallel()

	ac := AuthContext{
		System:         map[string]string{"k": "system", "s": "sys"},
		Metadata:       map[string]string{"k": "metadata", "m": "meta"},
		ProviderInputs: map[string]string{"k": "provider", "p": "prov"},
		ConsumerInputs: map[string]string{"k": "consumer", "c": "cons"},
		Secrets:        map[string]string{"k": "secret", "x": "sec"},
	}

	flat := ac.Flatten()

	// Secrets win the shared key; every unique key survives.
	for key, want := range map[string]string{
		"k": "secret", "s": "sys", "m": "meta", "p": "prov", "c": "cons", "x": "sec",
	} {
		if flat[key] != want {
			t.Errorf("Flatten()[%q] = %q, want %q", key, flat[key], want)
		}
	}
}

func TestJSONSecretParser(t *testing.T) {
	t.Parallel()

	handler := JSONSecretParser(map[string]string{
		"access_token": "accessToken",
		"expires_in":   "expiresIn",
	})

	state, err := handler(context.Background(), NewAuthContext(),
		jsonResponse(`{"access_token":"abc123","expires_in":3599,"ignored":"x"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if state.Secrets["accessToken"] != "abc123" {
		t.Errorf("accessToken = %q, want abc123", state.Secrets["accessToken"])
	}

	// Numeric fields (expires_in) are coerced to their string form.
	if state.Secrets["expiresIn"] != "3599" {
		t.Errorf("expiresIn = %q, want 3599", state.Secrets["expiresIn"])
	}
}

func TestQueryParamSecretParser(t *testing.T) {
	t.Parallel()

	handler := QueryParamSecretParser(map[string]string{"code": "code"})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet,
		"https://cb.example.com/callback?code=xyz&state=s", nil)

	state, err := handler(context.Background(), NewAuthContext(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if state.Secrets["code"] != "xyz" {
		t.Errorf("code = %q, want xyz", state.Secrets["code"])
	}
}

func TestMicrosoftBuildConsentURL(t *testing.T) {
	t.Parallel()

	state := NewAuthContext()
	state.ProviderInputs = map[string]string{"clientId": "app-123"}
	state.System = map[string]string{"callbackURL": "https://api.example.com/callbacks/v1/custom-auth"}

	_, consentURL, err := msBuildConsentURL(context.Background(), state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, _ := url.Parse(consentURL)
	if got := parsed.Query().Get("client_id"); got != "app-123" {
		t.Errorf("client_id = %q, want app-123", got)
	}

	if got := parsed.Query().Get("redirect_uri"); got != "https://api.example.com/callbacks/v1/custom-auth" {
		t.Errorf("redirect_uri = %q", got)
	}
}

func TestMicrosoftBuildConsentURLMissingClientId(t *testing.T) {
	t.Parallel()

	if _, _, err := msBuildConsentURL(context.Background(), NewAuthContext()); err == nil {
		t.Fatal("expected error when clientId is missing")
	}
}

func TestMicrosoftParseConsentCallback(t *testing.T) {
	t.Parallel()

	t.Run("captures tenant", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet,
			"https://cb?tenant=tid-1&admin_consent=True", nil)

		state, err := msParseConsentCallback(context.Background(), NewAuthContext(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if state.Metadata["workspace"] != "tid-1" {
			t.Errorf("workspace = %q, want tid-1", state.Metadata["workspace"])
		}
	})

	t.Run("error param surfaces", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet,
			"https://cb?error=access_denied&error_description=nope", nil)
		if _, err := msParseConsentCallback(context.Background(), NewAuthContext(), req); err == nil {
			t.Fatal("expected error when callback carries an error param")
		}
	})

	t.Run("missing tenant is an error", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet,
			"https://cb?admin_consent=True", nil)
		if _, err := msParseConsentCallback(context.Background(), NewAuthContext(), req); err == nil {
			t.Fatal("expected error when tenant is missing")
		}
	})

	t.Run("consent not granted is an error", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet,
			"https://cb?tenant=tid-1", nil) // no admin_consent=True
		if _, err := msParseConsentCallback(context.Background(), NewAuthContext(), req); err == nil {
			t.Fatal("expected error when admin_consent is not True")
		}
	})
}

func TestMicrosoftBuildTokenRequest(t *testing.T) {
	t.Parallel()

	state := NewAuthContext()
	state.ProviderInputs = map[string]string{"clientId": "app-123", "clientSecret": "shh"}
	state.Metadata = map[string]string{"workspace": "tid-1"}

	_, req, err := msBuildTokenRequest(context.Background(), state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Method != http.MethodPost {
		t.Errorf("method = %s, want POST", req.Method)
	}

	if req.URL.String() != "https://login.microsoftonline.com/tid-1/oauth2/v2.0/token" {
		t.Errorf("token URL = %q", req.URL.String())
	}

	body, _ := io.ReadAll(req.Body)
	form, _ := url.ParseQuery(string(body))
	for key, want := range map[string]string{
		"grant_type": "client_credentials", "client_id": "app-123",
		"client_secret": "shh", "scope": msDefaultScope,
	} {
		if form.Get(key) != want {
			t.Errorf("form[%q] = %q, want %q", key, form.Get(key), want)
		}
	}
}

func TestMicrosoftBuildTokenRequestMissingTenant(t *testing.T) {
	t.Parallel()

	if _, _, err := msBuildTokenRequest(context.Background(), NewAuthContext()); err == nil {
		t.Fatal("expected error when tenant is missing")
	}
}

func TestMicrosoftFlowRegistered(t *testing.T) {
	t.Parallel()

	flow, ok := CustomAuthFlowFor(MicrosoftAdminConsent)
	if !ok {
		t.Fatal("microsoftAdminConsent custom flow not registered")
	}

	if len(flow.ConnectSteps) != 2 {
		t.Fatalf("want 2 connect steps, got %d", len(flow.ConnectSteps))
	}

	if flow.ConnectSteps[0].Redirect == nil {
		t.Error("first connect step should be a redirect")
	}

	if flow.ConnectSteps[1].HTTP == nil {
		t.Error("second connect step should be an HTTP call")
	}

	if !flow.HasRedirect() {
		t.Error("flow should report a redirect step")
	}

	if len(flow.RefreshSteps) != 1 {
		t.Errorf("want 1 refresh step, got %d", len(flow.RefreshSteps))
	}
}

func TestAuthContextEnsureMaps(t *testing.T) {
	t.Parallel()

	got := AuthContext{}.EnsureMaps() // all sub-maps nil, as after a Redis round-trip

	if got.ConsumerInputs == nil || got.ProviderInputs == nil ||
		got.Secrets == nil || got.Metadata == nil || got.System == nil {
		t.Fatalf("EnsureMaps left a nil sub-map: %+v", got)
	}

	got.Metadata["workspace"] = "tenant" // must not panic
	got.Secrets["accessToken"] = "tok"

	// Existing entries are preserved, not clobbered.
	preserved := AuthContext{Secrets: map[string]string{"k": "v"}}.EnsureMaps()
	if preserved.Secrets["k"] != "v" {
		t.Errorf("EnsureMaps clobbered a populated map: %+v", preserved.Secrets)
	}
}
