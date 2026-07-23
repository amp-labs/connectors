package providers

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

// roundTripFunc adapts a function to an http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// refreshTestInfo is a throwaway provider that declares a single auth header
// rendered from the accessToken value, exercising the refresh-on-401 path.
func refreshTestInfo() *ProviderInfo {
	return &ProviderInfo{
		DisplayName: "Refresh Test",
		AuthType:    Custom,
		BaseURL:     "https://example.com",
		CustomOpts: &CustomAuthOpts{
			Headers: []CustomAuthHeader{
				{Name: "Authorization", ValueTemplate: "Bearer {{ .accessToken }}"},
			},
		},
	}
}

func newResponse(status int, req *http.Request) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       http.NoBody,
		Request:    req,
	}
}

func TestCustomAuthRefreshOn401(t *testing.T) {
	t.Parallel()

	var (
		calls        int
		replayedAuth []string
	)

	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			// Stale token; server rejects it.
			return newResponse(http.StatusUnauthorized, req), nil
		}

		// Replay: capture the Authorization header(s) the client applied.
		replayedAuth = req.Header.Values("Authorization")

		return newResponse(http.StatusOK, req), nil
	})

	cfg := &CustomAuthParams{
		// Simulate the stale value the client was originally built with.
		Values: map[string]string{"accessToken": "STALE"},
		Refresh: func(context.Context) (map[string]string, error) {
			return map[string]string{"accessToken": "FRESH"}, nil
		},
	}

	client, err := createCustomHTTPClient(
		context.Background(),
		&http.Client{Transport: transport},
		false, nil, nil,
		refreshTestInfo(), cfg,
	)
	if err != nil {
		t.Fatalf("createCustomHTTPClient: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com/v1/thing", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	rsp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do: %v", err)
	}

	if rsp.StatusCode != http.StatusOK {
		t.Errorf("final status = %d, want 200", rsp.StatusCode)
	}

	if calls != 2 {
		t.Errorf("transport called %d times, want 2 (original + one replay)", calls)
	}

	if len(replayedAuth) != 1 {
		t.Fatalf("replay carried %d Authorization headers %v, want exactly 1", len(replayedAuth), replayedAuth)
	}

	if replayedAuth[0] != "Bearer FRESH" {
		t.Errorf("replay Authorization = %q, want %q", replayedAuth[0], "Bearer FRESH")
	}
}

// refreshTestInfoQueryParam declares auth via a query param, exercising the
// query-param overwrite path on refresh.
func refreshTestInfoQueryParam() *ProviderInfo {
	return &ProviderInfo{
		DisplayName: "Refresh Test QP",
		AuthType:    Custom,
		BaseURL:     "https://example.com",
		CustomOpts: &CustomAuthOpts{
			QueryParams: []CustomAuthQueryParam{
				{Name: "access_token", ValueTemplate: "{{ .accessToken }}"},
			},
		},
	}
}

func TestCustomAuthRefreshOn401_QueryParam(t *testing.T) {
	t.Parallel()

	var (
		calls        int
		replayedVals []string
	)

	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return newResponse(http.StatusUnauthorized, req), nil
		}

		replayedVals = req.URL.Query()["access_token"]

		return newResponse(http.StatusOK, req), nil
	})

	cfg := &CustomAuthParams{
		Values: map[string]string{"accessToken": "STALE"},
		Refresh: func(context.Context) (map[string]string, error) {
			return map[string]string{"accessToken": "FRESH"}, nil
		},
	}

	client, err := createCustomHTTPClient(
		context.Background(),
		&http.Client{Transport: transport},
		false, nil, nil,
		refreshTestInfoQueryParam(), cfg,
	)
	if err != nil {
		t.Fatalf("createCustomHTTPClient: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com/v1/thing", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	rsp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do: %v", err)
	}

	if rsp.StatusCode != http.StatusOK {
		t.Errorf("final status = %d, want 200", rsp.StatusCode)
	}

	// Overwrite (not append): exactly one access_token, and it's the refreshed value.
	if len(replayedVals) != 1 {
		t.Fatalf("replay carried %d access_token query params %v, want exactly 1", len(replayedVals), replayedVals)
	}

	if replayedVals[0] != "FRESH" {
		t.Errorf("replay access_token = %q, want FRESH", replayedVals[0])
	}
}

func TestCustomAuthRefreshErrorReturnsOriginal401(t *testing.T) {
	t.Parallel()

	var calls int

	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++

		return newResponse(http.StatusUnauthorized, req), nil
	})

	cfg := &CustomAuthParams{
		Values: map[string]string{"accessToken": "STALE"},
		Refresh: func(context.Context) (map[string]string, error) {
			return nil, errors.New("refresh boom")
		},
	}

	client, err := createCustomHTTPClient(
		context.Background(),
		&http.Client{Transport: transport},
		false, nil, nil,
		refreshTestInfo(), cfg,
	)
	if err != nil {
		t.Fatalf("createCustomHTTPClient: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com/v1/thing", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	rsp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do returned error, want original 401 with nil err: %v", err)
	}

	if rsp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 (original response, unmasked)", rsp.StatusCode)
	}

	if calls != 1 {
		t.Errorf("transport called %d times, want 1 (no replay on refresh error)", calls)
	}
}
