package main

import (
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestPasswordGrantForm(t *testing.T) {
	t.Parallel()

	params := map[string]string{"client_id": "sugar", "client_secret": "", "platform": "custom_api"}

	form := passwordGrantForm(&oauth2.Config{}, params, "jane", "pw")
	if form.Get("client_id") != "sugar" || !form.Has("client_secret") || form.Get("platform") != "custom_api" {
		t.Fatalf("default client and platform not sent: %v", form)
	}

	form = passwordGrantForm(&oauth2.Config{ClientID: "mine", ClientSecret: "s"}, params, "jane", "pw")
	if form.Get("client_id") != "mine" || form.Get("client_secret") != "s" || form.Get("platform") != "custom_api" {
		t.Fatalf("own client should replace the default one: %v", form)
	}

	if form.Get("grant_type") != "password" || form.Get("username") != "jane" || form.Get("password") != "pw" {
		t.Fatalf("grant fields missing: %v", form)
	}
}

func TestParseTokenResponse(t *testing.T) {
	t.Parallel()

	tok, err := parseTokenResponse([]byte(`{"access_token":"AT","refresh_token":"RT","expires_in":"3600"}`))
	if err != nil {
		t.Fatal(err)
	}

	if tok.AccessToken != "AT" || tok.RefreshToken != "RT" || time.Until(tok.Expiry) < 59*time.Minute {
		t.Fatalf("unexpected token: %+v", tok)
	}
}
