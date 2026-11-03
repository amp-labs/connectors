package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// passwordGrantToken signs in with the OAuth 2.0 resource owner password grant.
//
// It is used instead of oauth2.Config.PasswordCredentialsToken because that cannot send
// the provider's extra token parameters (Oauth2Opts.TokenParams), such as SugarAI's
// platform. Those are added to the form unless the grant already sets the field, and a
// client_id or client_secret among them is only used when no client ID is configured.
// The client is identified in the form body.
func passwordGrantToken(
	ctx context.Context, cfg *oauth2.Config, tokenParams map[string]string, username, password string,
) (*oauth2.Token, error) {
	form := passwordGrantForm(cfg, tokenParams, username, password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.Endpoint.TokenURL,
		strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(rsp.Body, 1<<20)) //nolint:mnd
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode < http.StatusOK || rsp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("token request failed with HTTP %d: %s", rsp.StatusCode, body) //nolint:err113
	}

	return parseTokenResponse(body)
}

// passwordGrantForm builds the form of a password grant token request.
func passwordGrantForm(
	cfg *oauth2.Config, tokenParams map[string]string, username, password string,
) url.Values {
	form := url.Values{
		"grant_type": {"password"},
		"username":   {username},
		"password":   {password},
	}

	if len(cfg.Scopes) > 0 {
		form.Set("scope", strings.Join(cfg.Scopes, " "))
	}

	if cfg.ClientID != "" {
		form.Set("client_id", cfg.ClientID)

		if cfg.ClientSecret != "" {
			form.Set("client_secret", cfg.ClientSecret)
		}
	}

	for key, value := range tokenParams {
		if cfg.ClientID != "" && (key == "client_id" || key == "client_secret") {
			continue
		}

		if !form.Has(key) {
			form.Set(key, value)
		}
	}

	return form
}

// parseTokenResponse parses a JSON token response.
func parseTokenResponse(body []byte) (*oauth2.Token, error) {
	var tok struct {
		AccessToken  string          `json:"access_token"`  //nolint:tagliatelle
		TokenType    string          `json:"token_type"`    //nolint:tagliatelle
		RefreshToken string          `json:"refresh_token"` //nolint:tagliatelle
		ExpiresIn    json.RawMessage `json:"expires_in"`    //nolint:tagliatelle
	}

	err := json.Unmarshal(body, &tok)
	if err != nil {
		return nil, fmt.Errorf("cannot parse token response: %w", err)
	}

	token := &oauth2.Token{
		AccessToken:  tok.AccessToken,
		TokenType:    tok.TokenType,
		RefreshToken: tok.RefreshToken,
	}

	// Some providers (Sugar among them) send expires_in as a string.
	seconds, err := strconv.Atoi(strings.Trim(string(tok.ExpiresIn), `"`))
	if err == nil && seconds > 0 {
		token.Expiry = time.Now().Add(time.Duration(seconds) * time.Second)
	}

	return token, nil
}
