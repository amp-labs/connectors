package m2m

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// DefaultScopes are the OAuth 2.0 scopes requested for NetSuite M2M connections.
var DefaultScopes = []string{"restlets", "rest_webservices"}

// ParseECPrivateKey parses a PEM-encoded EC private key.
// Supports both SEC 1 (BEGIN EC PRIVATE KEY) and PKCS#8 (BEGIN PRIVATE KEY) formats.
func ParseECPrivateKey(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in private key")
	}

	// Try SEC 1 format (BEGIN EC PRIVATE KEY)
	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	// Try PKCS#8 format (BEGIN PRIVATE KEY)
	pkcs8Key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EC private key (tried SEC1 and PKCS8): %w", err)
	}

	ecKey, ok := pkcs8Key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("PKCS8 key is not an ECDSA key")
	}

	return ecKey, nil
}

// tokenSource implements oauth2.TokenSource for NetSuite M2M OAuth 2.0.
// It signs a JWT with the EC private key and exchanges it at the token endpoint
// for a Bearer access token.
type tokenSource struct {
	clientID      string
	certificateID string
	tokenURL      string
	scopes        []string
	privateKey    *ecdsa.PrivateKey
	httpClient    *http.Client
}

// Token creates a signed JWT assertion and exchanges it for an access token.
func (s *tokenSource) Token() (*oauth2.Token, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"iss":   s.clientID,
		"scope": strings.Join(s.scopes, " "),
		"aud":   s.tokenURL,
		"iat":   now.Unix(),
		"exp":   now.Add(30 * time.Minute).Unix(), // Well under NetSuite's 60-min limit
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = s.certificateID

	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("signing M2M JWT: %w", err)
	}

	data := url.Values{
		"grant_type":            {"client_credentials"},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {signed},
	}

	client := s.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.PostForm(s.tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("M2M token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errBody) //nolint:errcheck

		return nil, fmt.Errorf("M2M token endpoint returned %d: %v", resp.StatusCode, errBody)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decoding M2M token response: %w", err)
	}

	return &oauth2.Token{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		Expiry:      now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}

// NewHeadersGenerator returns a DynamicHeadersGenerator that manages the full
// M2M token lifecycle: JWT signing → token endpoint → Bearer header.
//
// It wraps the token source in oauth2.ReuseTokenSource which handles:
//   - Thread-safe token caching
//   - Automatic refresh ~10s before expiry
//   - Concurrent request safety
func NewHeadersGenerator(
	accountID, clientID, certificateID string,
	privateKey *ecdsa.PrivateKey,
	scopes []string,
) common.DynamicHeadersGenerator {
	tokenURL := fmt.Sprintf(
		"https://%s.suitetalk.api.netsuite.com/services/rest/auth/oauth2/v1/token",
		accountID,
	)

	// ReuseTokenSource caches the token and only calls Token() when expired.
	ts := oauth2.ReuseTokenSource(nil, &tokenSource{
		clientID:      clientID,
		certificateID: certificateID,
		tokenURL:      tokenURL,
		scopes:        scopes,
		privateKey:    privateKey,
	})

	return func(req *http.Request) ([]common.Header, error) {
		tok, err := ts.Token()
		if err != nil {
			return nil, fmt.Errorf("M2M token: %w", err)
		}

		return []common.Header{
			{Key: "Authorization", Value: tok.TokenType + " " + tok.AccessToken},
		}, nil
	}
}
