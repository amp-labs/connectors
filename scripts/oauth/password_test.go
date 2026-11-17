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

	tests := []struct {
		name       string
		body       string
		wantExpiry bool
	}{
		{
			name:       "expires_in is a number",
			body:       `{"access_token":"AT","refresh_token":"RT","expires_in":3600}`,
			wantExpiry: true,
		},
		{
			name:       "expires_in is a string",
			body:       `{"access_token":"AT","refresh_token":"RT","expires_in":"3600"}`,
			wantExpiry: true,
		},
		{
			name:       "expires_in is null",
			body:       `{"access_token":"AT","refresh_token":"RT","expires_in":null}`,
			wantExpiry: false,
		},
		{
			name:       "expires_in is absent",
			body:       `{"access_token":"AT","refresh_token":"RT"}`,
			wantExpiry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tok, err := parseTokenResponse([]byte(tt.body))
			if err != nil {
				t.Fatal(err)
			}

			if tok.AccessToken != "AT" || tok.RefreshToken != "RT" {
				t.Fatalf("unexpected token: %+v", tok)
			}

			if !tt.wantExpiry {
				if !tok.Expiry.IsZero() {
					t.Fatalf("expiry should be unset, got %v", tok.Expiry)
				}

				return
			}

			// The expiry is set from time.Now(), so allow a minute for the test to run.
			if remaining := time.Until(tok.Expiry); remaining < 59*time.Minute || remaining > time.Hour {
				t.Fatalf("expiry should be an hour away, is %v", remaining)
			}
		})
	}
}
