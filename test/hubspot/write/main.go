package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/hubspot"
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

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	// Get the Hubspot connector.
	conn := connTest.GetHubspotConnector(ctx)
	defer utils.Close(conn)

	// Write an artificial contact to Hubspot.
	result, err := conn.Write(ctx, common.WriteParams{
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

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("email", "phone", "company", "website", "lastname", "firstname"),
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
