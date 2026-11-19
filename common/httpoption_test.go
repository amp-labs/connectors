package common

import (
	"net/http"
	"testing"
)

type captureClient struct {
	req *http.Request
}

func (c *captureClient) Do(req *http.Request) (*http.Response, error) {
	c.req = req

	return &http.Response{StatusCode: http.StatusOK}, nil
}

func (c *captureClient) CloseIdleConnections() {}

func TestNewHTTPOptionsClientEmptyReturnsUnwrapped(t *testing.T) {
	t.Parallel()

	inner := &captureClient{}

	if got := NewHTTPOptionsClient(inner, nil); got != AuthenticatedHTTPClient(inner) {
		t.Errorf("expected inner client to be returned unwrapped, got %T", got)
	}
}

func TestHTTPOptionsClientAppliesHeaderAndQuery(t *testing.T) {
	t.Parallel()

	inner := &captureClient{}
	client := NewHTTPOptionsClient(inner, []HTTPOption{
		{In: HTTPOptionInHeader, Key: "ClientId", Value: "my-client-id"},
		{In: HTTPOptionInQuery, Key: "apiVersion", Value: "3.0"},
	})

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://example.com/path?existing=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.Do(req); err != nil {
		t.Fatal(err)
	}

	if got := inner.req.Header.Get("ClientId"); got != "my-client-id" {
		t.Errorf("expected ClientId header %q, got %q", "my-client-id", got)
	}

	query := inner.req.URL.Query()
	if got := query.Get("apiVersion"); got != "3.0" {
		t.Errorf("expected apiVersion query param %q, got %q", "3.0", got)
	}

	if got := query.Get("existing"); got != "1" {
		t.Errorf("expected existing query param to be preserved, got %q", got)
	}

	// The original request must not be mutated.
	if got := req.Header.Get("ClientId"); got != "" {
		t.Errorf("expected original request to be unmodified, got ClientId=%q", got)
	}

	if got := req.URL.Query().Get("apiVersion"); got != "" {
		t.Errorf("expected original request URL to be unmodified, got apiVersion=%q", got)
	}
}

func TestHTTPOptionsClientDoesNotOverwriteExistingValues(t *testing.T) {
	t.Parallel()

	inner := &captureClient{}
	client := NewHTTPOptionsClient(inner, []HTTPOption{
		{In: HTTPOptionInHeader, Key: "ClientId", Value: "from-options"},
		{In: HTTPOptionInQuery, Key: "apiVersion", Value: "from-options"},
	})

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://example.com/path?apiVersion=caller", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("ClientId", "caller")

	if _, err := client.Do(req); err != nil {
		t.Fatal(err)
	}

	if got := inner.req.Header.Get("ClientId"); got != "caller" {
		t.Errorf("expected caller-supplied header to win, got %q", got)
	}

	if got := inner.req.URL.Query()["apiVersion"]; len(got) != 1 || got[0] != "caller" {
		t.Errorf("expected caller-supplied query param to win, got %v", got)
	}
}
