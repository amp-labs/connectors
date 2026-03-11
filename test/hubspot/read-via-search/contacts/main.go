package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/providers/hubspot"
	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetHubspotConnector(ctx)

	res, err := conn.ReadUsingSearchAPI(ctx, hubspot.SearchParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("email", "phone", "company", "website", "lastname", "firstname"),
		AssociatedObjects: []string{
			"companies",
		},
	})
	if err != nil {
		utils.Fail("error reading from Hubspot", "error", err)
	}

	slog.Info("Reading contacts..")
	utils.DumpJSON(res, os.Stdout)
}
