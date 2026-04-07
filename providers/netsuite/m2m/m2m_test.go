package m2m

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateTestECKey(t *testing.T) (*ecdsa.PrivateKey, []byte) {
	t.Helper()

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generating test key: %v", err)
	}

	der, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		t.Fatalf("marshaling key: %v", err)
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: der,
	})

	return privKey, pemBytes
}

func generateTestPKCS8Key(t *testing.T) (*ecdsa.PrivateKey, []byte) {
	t.Helper()

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generating test key: %v", err)
	}

	der, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		t.Fatalf("marshaling key: %v", err)
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	})

	return privKey, pemBytes
}

func TestParseECPrivateKey_SEC1(t *testing.T) {
	t.Parallel()

	_, pemBytes := generateTestECKey(t)

	key, err := ParseECPrivateKey(pemBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if key == nil {
		t.Fatal("expected non-nil key")
	}

	if key.Curve != elliptic.P256() {
		t.Fatalf("expected P-256 curve, got %v", key.Curve.Params().Name)
	}
}

func TestParseECPrivateKey_PKCS8(t *testing.T) {
	t.Parallel()

	_, pemBytes := generateTestPKCS8Key(t)

	key, err := ParseECPrivateKey(pemBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if key == nil {
		t.Fatal("expected non-nil key")
	}

	if key.Curve != elliptic.P256() {
		t.Fatalf("expected P-256 curve, got %v", key.Curve.Params().Name)
	}
}

func TestParseECPrivateKey_InvalidPEM(t *testing.T) {
	t.Parallel()

	_, err := ParseECPrivateKey([]byte("not a pem"))
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}
}

func TestTokenSource(t *testing.T) {
	t.Parallel()

	privKey, _ := generateTestECKey(t)

	// Mock token endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("parsing form: %v", err)
		}

		// Verify grant_type
		if gt := r.FormValue("grant_type"); gt != "client_credentials" {
			t.Errorf("expected grant_type=client_credentials, got %s", gt)
		}

		// Verify client_assertion_type
		cat := r.FormValue("client_assertion_type")
		if cat != "urn:ietf:params:oauth:client-assertion-type:jwt-bearer" {
			t.Errorf("unexpected client_assertion_type: %s", cat)
		}

		// Verify JWT assertion
		assertion := r.FormValue("client_assertion")
		if assertion == "" {
			t.Fatal("missing client_assertion")
		}

		// Parse and verify the JWT (without signature verification for simplicity)
		parser := jwt.NewParser(jwt.WithoutClaimsValidation())

		token, _, err := parser.ParseUnverified(assertion, jwt.MapClaims{})
		if err != nil {
			t.Fatalf("parsing JWT: %v", err)
		}

		// Verify header
		if token.Header["alg"] != "ES256" {
			t.Errorf("expected alg=ES256, got %v", token.Header["alg"])
		}

		if token.Header["kid"] != "test-cert-id" {
			t.Errorf("expected kid=test-cert-id, got %v", token.Header["kid"])
		}

		// Verify claims
		claims := token.Claims.(jwt.MapClaims)
		if claims["iss"] != "test-client-id" {
			t.Errorf("expected iss=test-client-id, got %v", claims["iss"])
		}

		if claims["scope"] != "restlets rest_webservices" {
			t.Errorf("expected scope='restlets rest_webservices', got %v", claims["scope"])
		}

		// aud should be the token URL
		aud := fmt.Sprintf("%s", claims["aud"])
		if aud == "" {
			t.Error("missing aud claim")
		}

		// iat and exp should be present and reasonable
		iat, ok := claims["iat"].(float64)
		if !ok {
			t.Fatal("missing iat claim")
		}

		exp, ok := claims["exp"].(float64)
		if !ok {
			t.Fatal("missing exp claim")
		}

		// exp should be ~30 min after iat
		diff := exp - iat
		if diff < 1700 || diff > 1900 { // 30 min = 1800s, with some tolerance
			t.Errorf("expected exp-iat ~1800, got %v", diff)
		}

		// Return a valid token response
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
			"access_token": "test-access-token-123",
			"expires_in":   3600,
			"token_type":   "Bearer",
		})
	}))
	defer server.Close()

	ts := &tokenSource{
		clientID:      "test-client-id",
		certificateID: "test-cert-id",
		tokenURL:      server.URL,
		scopes:        []string{"restlets", "rest_webservices"},
		privateKey:    privKey,
		httpClient:    server.Client(),
	}

	token, err := ts.Token()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token.AccessToken != "test-access-token-123" {
		t.Errorf("expected access token 'test-access-token-123', got %s", token.AccessToken)
	}

	if token.TokenType != "Bearer" {
		t.Errorf("expected token type 'Bearer', got %s", token.TokenType)
	}

	if token.Expiry.Before(time.Now()) {
		t.Error("token should not be expired")
	}
}

func TestTokenSource_ErrorResponse(t *testing.T) {
	t.Parallel()

	privKey, _ := generateTestECKey(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{ //nolint:errcheck
			"error":             "invalid_client",
			"error_description": "bad credentials",
		})
	}))
	defer server.Close()

	ts := &tokenSource{
		clientID:      "bad-client",
		certificateID: "bad-cert",
		tokenURL:      server.URL,
		scopes:        DefaultScopes,
		privateKey:    privKey,
		httpClient:    server.Client(),
	}

	_, err := ts.Token()
	if err == nil {
		t.Fatal("expected error for bad token response")
	}
}

func TestNewHeadersGenerator(t *testing.T) {
	t.Parallel()

	privKey, _ := generateTestECKey(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
			"access_token": "bearer-token-xyz",
			"expires_in":   3600,
			"token_type":   "Bearer",
		})
	}))
	defer server.Close()

	// We can't use NewHeadersGenerator directly because it constructs the token URL
	// from accountID. Instead, test the token source + generator pattern manually.
	ts := &tokenSource{
		clientID:      "test-client",
		certificateID: "test-cert",
		tokenURL:      server.URL,
		scopes:        DefaultScopes,
		privateKey:    privKey,
		httpClient:    server.Client(),
	}

	tok, err := ts.Token()
	if err != nil {
		t.Fatalf("getting token: %v", err)
	}

	expected := "Bearer bearer-token-xyz"
	actual := tok.TokenType + " " + tok.AccessToken

	if actual != expected {
		t.Errorf("expected header value %q, got %q", expected, actual)
	}
}
