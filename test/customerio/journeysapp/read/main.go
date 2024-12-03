package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/customerio/journeysapp"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetCustomerJourneysAppConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "newsletters",
		Fields:     connectors.Fields("name"),
		// Read newsletters using pagination, where it omits first record.
		NextPage: `https://api.customer.io/v1/newsletters?limit=50&start=MQ==`,
	})
	if err != nil {
		utils.Fail("error reading from Customer Journeys App", "error", err)
	}

	slog.Info("Reading newsletters..")
	utils.DumpJSON(res, os.Stdout)
}
