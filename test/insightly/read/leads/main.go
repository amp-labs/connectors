package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/insightly"
	"github.com/amp-labs/connectors/test/utils"
)

const objectName = "Leads"

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetInsightlyConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("FIRST_NAME", "LEAD_ID"),
	})
	if err != nil {
		utils.Fail("error reading from provider", "error", err)
	}

	slog.Info("Reading objects..")
	utils.DumpJSON(res, os.Stdout)
}
