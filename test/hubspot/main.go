package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/hubspot"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/spyzhov/ajson"
	"golang.org/x/oauth2"
)

// Contact is a basic Hubspot contact.
type Contact struct {
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Company   string `json:"company"`
	Website   string `json:"website"`
	Lastname  string `json:"lastname"`
	Firstname string `json:"firstname"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	// Get the Hubspot connector.
	hsConn := getConnector(ctx)
	defer utils.Close(hsConn)

	// Write an artificial contact to Hubspot.
	result, err := hsConn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		ObjectId:   "",
		ObjectData: map[string]interface{}{
			"properties": &Contact{
				Email:     gofakeit.Email(),
				Phone:     gofakeit.Phone(),
				Company:   gofakeit.Company(),
				Website:   gofakeit.URL(),
				Lastname:  gofakeit.LastName(),
				Firstname: gofakeit.FirstName(),
			},
		},
	})
	if err != nil {
		utils.Fail("error writing to hubspot", "error", err)
	}

	// Dump the result.
	utils.DumpJSON(result, os.Stdout)
}

// getConnector returns a Hubspot connector.
func getConnector(ctx context.Context) *hubspot.Connector {
	creds, err := utils.Credentials()
	if err != nil {
		utils.Fail("error getting credentials", "error", err)
	}

	cfg, tok := getOAuthInfo(creds)

	conn, err := connectors.Hubspot(
		hubspot.WithClient(ctx, http.DefaultClient, cfg, tok),
		hubspot.WithModule(hubspot.ModuleCRM))
	if err != nil {
		utils.Fail("error creating hubspot connector", "error", err)
	}

	return conn
}

// getOAuthInfo returns the OAuth2 config and token from the creds.json file.
func getOAuthInfo(creds *ajson.Node) (*oauth2.Config, *oauth2.Token) {
	config := creds.MustObject()

	clientId, found := config["CLIENT_ID"]
	if !found {
		utils.Fail("CLIENT_ID not found in creds")
	}

	clientSecret, found := config["CLIENT_SECRET"]

	if !found {
		utils.Fail("CLIENT_SECRET not found in creds")
	}

	accessToken, found := config["ACCESS_TOKEN"]

	if !found {
		utils.Fail("ACCESS_TOKEN not found in creds")
	}

	refreshToken, found := config["REFRESH_TOKEN"]

	if !found {
		utils.Fail("REFRESH_TOKEN not found in creds")
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
