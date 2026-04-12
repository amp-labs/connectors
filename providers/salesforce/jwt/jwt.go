// Package jwt implements the Salesforce OAuth 2.0 JWT Bearer Flow
// (RFC 7523 §2.1) for server-to-server authentication.
//
// Unlike the NetSuite M2M flow (RFC 7523 §2.2, client_credentials +
// client_assertion), Salesforce uses the JWT itself as the grant:
//
//	grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer&assertion=<JWT>
//
// The JWT is signed with RS256 using the private key whose X.509 certificate
// is registered on the Salesforce Connected App. Salesforce returns a short
// Bearer access token and NEVER a refresh token — we re-sign a new assertion
// whenever the cached token expires.
//
// Reference: https://help.salesforce.com/s/articleView?id=sf.remoteaccess_oauth_jwt_flow.htm
package jwt

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

var (
	ErrNoPEMBlock     = errors.New("no PEM block found in private key")
	ErrNotRSAKey      = errors.New("PKCS8 key is not an RSA key")
	ErrTokenEndpoint  = errors.New("salesforce JWT token endpoint error")
	ErrInvalidPrivKey = errors.New("invalid Salesforce JWT private key")
)

// Audiences for the JWT 'aud' claim. Salesforce requires the literal login/test
// hostname regardless of whether the token endpoint is posted to a My Domain
// URL. See https://help.salesforce.com/s/articleView?id=sf.remoteaccess_oauth_jwt_flow.htm
const (
	ProductionAudience = "https://login.salesforce.com"
	SandboxAudience    = "https://test.salesforce.com"
)

// jwtExpirySeconds is the lifetime of the signed JWT assertion. Salesforce
// rejects assertions with an `exp` more than ~3 minutes in the future, so we
// stay safely under that cap.
const jwtExpirySeconds = 180

// accessTokenTTL is how long we treat an issued access token as valid before
// re-exchanging. Salesforce does not return `expires_in` in the JWT Bearer
// response, and actual session lifetime depends on the org's Session Settings
// (default 2 hours, configurable down to 15 minutes). We pick a conservative
// 30-minute window so that even orgs with short sessions stay authenticated,
// while avoiding a round-trip on every request. The 401 retry path in
// NewHeadersGenerator handles orgs whose sessions are even shorter.
const accessTokenTTL = 30 * time.Minute

// ParseRSAPrivateKey parses an RSA private key from either raw PEM or
// base64-encoded PEM. Supports both PKCS1 (BEGIN RSA PRIVATE KEY) and PKCS8
// (BEGIN PRIVATE KEY) formats.
func ParseRSAPrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	// If the input doesn't look like PEM, try base64-decoding it first.
	if !strings.Contains(string(pemBytes), "-----BEGIN") {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(pemBytes)))
		if err == nil {
			pemBytes = decoded
		}
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, ErrNoPEMBlock
	}

	// Try PKCS1 format (BEGIN RSA PRIVATE KEY)
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	// Try PKCS8 format (BEGIN PRIVATE KEY)
	pkcs8Key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: tried PKCS1 and PKCS8: %w", ErrInvalidPrivKey, err)
	}

	rsaKey, ok := pkcs8Key.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrNotRSAKey
	}

	return rsaKey, nil
}

// DetectAudience chooses the correct JWT 'aud' claim value by inspecting the
// workspace subdomain.
//
// Heuristic:
//
//   - Workspaces containing ".sandbox" indicate an Enhanced-Domains sandbox
//     (e.g. "acme--dev.sandbox"). These map to test.salesforce.com.
//   - Workspaces containing "--" indicate a legacy pre-Enhanced-Domains
//     sandbox subdomain (e.g. "acme--dev"). These also map to test.
//     Note: this can false-positive on a small number of production orgs
//     whose My Domain historically used "--" in the name itself. In practice
//     this is extremely rare in Enhanced-Domains orgs (Spring '23+) because
//     Salesforce restricts new My Domain names from containing "--".
//   - Anything else → production (login.salesforce.com).
func DetectAudience(workspace string) string {
	if strings.Contains(workspace, ".sandbox") || strings.Contains(workspace, "--") {
		return SandboxAudience
	}

	return ProductionAudience
}

// TokenURL returns the Salesforce token endpoint for the given workspace.
// We post to the My Domain endpoint (Salesforce accepts this regardless of
// which 'aud' we claim) so the HTTP request goes straight to the customer's
// org without a login.salesforce.com redirect.
func TokenURL(workspace string) string {
	return fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/token", workspace)
}

// tokenSource implements oauth2.TokenSource for Salesforce JWT Bearer flow.
// It signs a JWT with the RSA private key and exchanges it at the token
// endpoint for a Bearer access token.
type tokenSource struct {
	clientID   string
	username   string
	audience   string
	tokenURL   string
	privateKey *rsa.PrivateKey
	httpClient *http.Client
}

// Token creates a signed JWT assertion and exchanges it for an access token.
func (s *tokenSource) Token() (*oauth2.Token, error) {
	return s.tokenWithContext(context.Background())
}

func (s *tokenSource) tokenWithContext(ctx context.Context) (*oauth2.Token, error) {
	now := time.Now()

	signed, err := s.signAssertion(now)
	if err != nil {
		return nil, err
	}

	data := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {signed},
	}

	client := s.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, s.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error building salesforce JWT token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending salesforce JWT token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		json.NewDecoder(resp.Body).Decode(&errBody) //nolint:errcheck

		return nil, fmt.Errorf("%w: status %d: %v", ErrTokenEndpoint, resp.StatusCode, errBody)
	}

	return parseTokenResponse(resp, now)
}

func (s *tokenSource) signAssertion(now time.Time) (string, error) {
	// Salesforce JWT Bearer claims — no 'scope' (forbidden), no 'kid' header
	// (Salesforce matches against the cert registered on the Connected App).
	claims := jwt.MapClaims{
		"iss": s.clientID,
		"sub": s.username,
		"aud": s.audience,
		"exp": now.Add(jwtExpirySeconds * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("signing Salesforce JWT: %w", err)
	}

	return signed, nil
}

// parseTokenResponse decodes the Salesforce token response and synthesises an
// expiry. Salesforce omits `expires_in` from this flow, so we stamp a
// conservative fixed TTL (see accessTokenTTL).
func parseTokenResponse(resp *http.Response, now time.Time) (*oauth2.Token, error) {
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		InstanceURL string `json:"instance_url"`
		ID          string `json:"id"`
		Scope       string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decoding Salesforce JWT token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("%w: empty access_token in response", ErrTokenEndpoint)
	}

	tokenType := tokenResp.TokenType
	if tokenType == "" {
		tokenType = "Bearer"
	}

	return &oauth2.Token{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenType,
		Expiry:      now.Add(accessTokenTTL),
	}, nil
}

// retryingTokenSource wraps an oauth2.ReuseTokenSource so we can read the
// current cached token but also force a fresh exchange, bypassing the cache.
// We use the force path from the 401 handler, to cover orgs whose session
// timeout is shorter than our fixed accessTokenTTL.
type retryingTokenSource struct {
	base  *tokenSource
	inner oauth2.TokenSource
}

func (r *retryingTokenSource) Token() (*oauth2.Token, error) {
	return r.inner.Token()
}

// forceRefresh mints a brand-new token directly from the base source,
// bypassing the cache, and installs it as the new ReuseTokenSource base so
// subsequent calls see the refreshed value.
func (r *retryingTokenSource) forceRefresh(ctx context.Context) (*oauth2.Token, error) {
	fresh, err := r.base.tokenWithContext(ctx)
	if err != nil {
		return nil, err
	}

	r.inner = oauth2.ReuseTokenSource(fresh, r.base)

	return fresh, nil
}

// NewAuthenticatedClient builds a fully wired AuthenticatedHTTPClient for the
// Salesforce JWT Bearer flow. It handles:
//
//   - JWT signing + token exchange on first request
//   - Thread-safe token caching with refresh ahead of expiry
//     (via oauth2.ReuseTokenSource)
//   - Transparent re-exchange + single retry on a 401 from the downstream
//     API (covers orgs whose session timeout is shorter than accessTokenTTL)
func NewAuthenticatedClient( //nolint:ireturn
	ctx context.Context,
	clientID, username, audience, tokenURL string,
	privateKey *rsa.PrivateKey,
) (common.AuthenticatedHTTPClient, error) {
	base := &tokenSource{
		clientID:   clientID,
		username:   username,
		audience:   audience,
		tokenURL:   tokenURL,
		privateKey: privateKey,
	}

	rts := &retryingTokenSource{
		base:  base,
		inner: oauth2.ReuseTokenSource(nil, base),
	}

	return common.NewCustomAuthHTTPClient(ctx,
		common.WithCustomDynamicHeaders(newHeadersGenerator(rts)),
		common.WithCustomUnauthorizedHandler(newUnauthorizedHandler(rts)), //nolint:bodyclose
	)
}

// newHeadersGenerator returns a DynamicHeadersGenerator backed by the given
// token source.
func newHeadersGenerator(rts *retryingTokenSource) common.DynamicHeadersGenerator {
	return func(_ *http.Request) ([]common.Header, error) {
		tok, err := rts.Token()
		if err != nil {
			return nil, fmt.Errorf("error in headers generator: %w", err)
		}

		return []common.Header{
			{Key: "Authorization", Value: tok.TokenType + " " + tok.AccessToken},
		}, nil
	}
}

// newUnauthorizedHandler returns a handler that the custom auth HTTP client
// invokes on a 401 response. It forces a fresh token exchange, rewrites the
// Authorization header, and replays the original request exactly once. If
// the retry also fails auth, the second response is returned to the caller
// as-is — we do not loop.
func newUnauthorizedHandler(rts *retryingTokenSource) func(
	hdrs []common.Header,
	params []common.QueryParam,
	req *http.Request,
	rsp *http.Response,
) (*http.Response, error) {
	return func(
		_ []common.Header, _ []common.QueryParam,
		req *http.Request, rsp *http.Response,
	) (*http.Response, error) {
		// Drain and close the stale 401 response body so the connection can
		// be reused.
		if rsp != nil && rsp.Body != nil {
			rsp.Body.Close() //nolint:errcheck
		}

		tok, err := rts.forceRefresh(req.Context())
		if err != nil {
			return nil, fmt.Errorf("refreshing salesforce JWT token after 401: %w", err)
		}

		// Clone the original request so we don't mutate it; replace the
		// Authorization header with the freshly minted token.
		retryReq := req.Clone(req.Context())
		retryReq.Header.Set("Authorization", tok.TokenType+" "+tok.AccessToken)

		return http.DefaultClient.Do(retryReq)
	}
}
