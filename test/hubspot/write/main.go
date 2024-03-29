package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	hsTest "github.com/amp-labs/connectors/test/hubspot"
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

/*
	Running instructions

	Step 1: prepare "hubspot-creds.json" file

	Create a file called "hubspot-creds.json" in the root of the project with the following contents

		e.g. {
		"CLIENT_ID": "<client id goes here>",
		"CLIENT_SECRET": "<client secret goes here>",
		"ACCESS_TOKEN": "<access token goes here>",
		"REFRESH_TOKEN": "<refresh token goes here>"
		}

	or export to an environment variable HUBSPOT_CRED_FILE by following command

	$> export HUBSPOT_CRED_FILE=./hubspot-creds.json # or the path to your hubspot-creds.json file


	In 1password, you can find a Hubspot creds.json file in the "Shared" vault.
	Look for the title "Hubspot Sample OAuth Credentials".
	The 1password item has an attached file called "creds.json" that contains the JSON.

	Step 2: run the following command

		$> go run test/hubspot/write/main.go


*/

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	filePath := os.Getenv("HUBSPOT_CRED_FILE")
	if filePath == "" {
		filePath = "./hubspot-creds.json"
	}

	// Get the Hubspot connector.
	hsConn := hsTest.GetHubspotConnector(ctx, filePath)
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
	fmt.Println("Wrote contact")

	res, err := hsConn.Read(ctx, common.ReadParams{
		ObjectName: "contacts",
		Fields:     []string{"email", "phone", "company", "website", "lastname", "firstname"},
		NextPage:   "",
		Since:      time.Now().Add(-5 * time.Minute),
	})
	if err != nil {
		utils.Fail("error reading from hubspot", "error", err)
	}

	fmt.Println("Reading contacts..")
	// Dump the result.
	utils.DumpJSON(res, os.Stdout)
}
