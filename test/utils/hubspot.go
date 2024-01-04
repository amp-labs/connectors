package utils

import (
	"context"
	"net/http"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/hubspot"
	"github.com/spyzhov/ajson"
	"golang.org/x/oauth2"
)

// GetHubspotConnector returns a Hubspot connector.
func GetHubspotConnector(ctx context.Context, defaultFileName string) *hubspot.Connector {
	creds, err := Credentials(defaultFileName)
	if err != nil {
		Fail("error getting credentials", "error", err)
	}

	cfg, tok := GetOAuthInfo(creds)

	conn, err := connectors.Hubspot(
		hubspot.WithClient(ctx, http.DefaultClient, cfg, tok),
		hubspot.WithModule(hubspot.ModuleCRM))
	if err != nil {
		Fail("error creating hubspot connector", "error", err)
	}

	return conn
}

// GetOAuthInfo returns the OAuth2 config and token from the creds.json file.
func GetOAuthInfo(creds *ajson.Node) (*oauth2.Config, *oauth2.Token) {
	config := creds.MustObject()

	clientId, found := config["CLIENT_ID"]
	if !found {
		Fail("CLIENT_ID not found in creds")
	}

	clientSecret, found := config["CLIENT_SECRET"]

	if !found {
		Fail("CLIENT_SECRET not found in creds")
	}

	// Treat this as optional, since we can get a new one if needed.
	accessToken, _ := config["ACCESS_TOKEN"]

	refreshToken, found := config["REFRESH_TOKEN"]

	if !found {
		Fail("REFRESH_TOKEN not found in creds")
	}

	cfg := &oauth2.Config{
		ClientID:     clientId.MustString(),
		ClientSecret: clientSecret.MustString(),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.hubspot.com/oauth/authorize",
			TokenURL:  "https://api.hubapi.com/oauth/v1/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"crm.objects.contacts.read",
			"crm.objects.contacts.write",
			"crm.objects.deals.read",
			"crm.objects.line_items.read",
			"oauth",
			"crm.objects.companies.read",
			"tickets",
		},
	}

	tok := &oauth2.Token{
		AccessToken:  accessToken.MustString(),
		RefreshToken: refreshToken.MustString(),
		TokenType:    "bearer",
		Expiry:       time.Now().Add(-1 * time.Hour), // just pretend it's expired already, whatever, it'll fetch a new one.
	}

	return cfg, tok
}
