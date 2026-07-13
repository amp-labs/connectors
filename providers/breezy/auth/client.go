// Package auth implements Breezy HR access-token authentication with reactive
// refresh on HTTP 401.
//
// Breezy has no OAuth refresh token. POST /v3/signin returns an access_token
// that is sent as Authorization: <token> (no Bearer prefix). While valid,
// repeat sign-in returns the same token; a new token is issued only after expiry.
//
// This client caches the token in memory and re-signs in only when an API call
// receives 401 Unauthorized, then retries the original request once.
package auth

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/amp-labs/connectors/common"
)

// Option configures NewAuthenticatedClient.
type Option func(*clientConfig)

// WithHTTPClient sets the HTTP client used for sign-in and API calls.
func WithHTTPClient(client *http.Client) Option {
	return func(cfg *clientConfig) {
		cfg.httpClient = client
	}
}

// WithSignInURL overrides the sign-in endpoint (for tests).
func WithSignInURL(signInURL string) Option {
	return func(cfg *clientConfig) {
		cfg.signInURL = signInURL
	}
}

type clientConfig struct {
	httpClient *http.Client
	signInURL  string
}

// NewAuthenticatedClient builds an HTTP client for Breezy API calls.
// The server is expected to call this with connection email/password credentials.
func NewAuthenticatedClient( //nolint:ireturn
	ctx context.Context,
	email, password string,
	opts ...Option,
) (common.AuthenticatedHTTPClient, error) {
	cfg := clientConfig{
		signInURL: DefaultSignInURL,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	httpClient := cfg.httpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	holder := &tokenHolder{
		email:     email,
		password:  password,
		signInURL: cfg.signInURL,
		client:    httpClient,
	}

	return common.NewCustomAuthHTTPClient(ctx,
		common.WithCustomClient(httpClient),
		common.WithCustomDynamicHeaders(holder.authorizationHeader),
		common.WithCustomUnauthorizedHandler(holder.unauthorizedHandler(httpClient)),
	)
}

type tokenHolder struct {
	mu sync.Mutex

	email, password string
	signInURL       string
	client          *http.Client
	token           string
}

func (h *tokenHolder) authorizationHeader(req *http.Request) ([]common.Header, error) {
	token, err := h.getToken(req.Context())
	if err != nil {
		return nil, fmt.Errorf("breezy authorization header: %w", err)
	}

	return []common.Header{
		{Key: "Authorization", Value: token},
	}, nil
}

func (h *tokenHolder) getToken(ctx context.Context) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.token != "" {
		return h.token, nil
	}

	token, err := signIn(ctx, h.client, h.signInURL, h.email, h.password)
	if err != nil {
		return "", err
	}

	h.token = token

	return token, nil
}

func (h *tokenHolder) invalidateAndRefresh(ctx context.Context) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.token = ""

	token, err := signIn(ctx, h.client, h.signInURL, h.email, h.password)
	if err != nil {
		return "", err
	}

	h.token = token

	return token, nil
}

func (h *tokenHolder) unauthorizedHandler(httpClient *http.Client) func(
	[]common.Header,
	[]common.QueryParam,
	*http.Request,
	*http.Response,
) (*http.Response, error) {
	return func(
		_ []common.Header, _ []common.QueryParam,
		req *http.Request, rsp *http.Response,
	) (*http.Response, error) {
		if rsp != nil && rsp.Body != nil {
			rsp.Body.Close() //nolint:errcheck
		}

		token, err := h.invalidateAndRefresh(req.Context())
		if err != nil {
			return nil, fmt.Errorf("refreshing breezy token after 401: %w", err)
		}

		retryReq := req.Clone(req.Context())
		retryReq.Header.Set("Authorization", token)

		return httpClient.Do(retryReq)
	}
}
