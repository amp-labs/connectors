package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
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

// Running instructions
//
// Step 1: prepare creds.json file
//
//	e.g. {
//	  "CLIENT_ID": "<client id goes here>",
//	  "CLIENT_SECRET": "<client secret goes here>",
//	  "ACCESS_TOKEN": "<access token goes here>",
//	  "REFRESH_TOKEN": "<refresh token goes here>"
//	}
//
// In 1password, you can find a Hubspot creds.json file in the "Shared" vault.
// Look for the title "Hubspot Sample OAuth Credentials".
// The 1password item has an attached file called "creds.json" that contains the JSON.
//
// Step 2: run the following commands
//
//	$> export CREDENTIALS_FILE=creds.json
//	$> go run test/hubspot/main.go
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	// Get the Hubspot connector.
	hsConn := utils.GetHubspotConnector(ctx, "creds.json")
	defer utils.Close(hsConn)

	// Write an artificial contact to Hubspot.
	result, err := hsConn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordId:   "",
		RecordData: &Contact{
			Email:     gofakeit.Email(),
			Phone:     gofakeit.Phone(),
			Company:   gofakeit.Company(),
			Website:   gofakeit.URL(),
			Lastname:  gofakeit.LastName(),
			Firstname: gofakeit.FirstName(),
		},
	})
	if err != nil {
		utils.Fail("error writing to hubspot", "error", err)
	}

	// Dump the result.
	utils.DumpJSON(result, os.Stdout)
}
