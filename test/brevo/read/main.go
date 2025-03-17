package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/brevo"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetBrevoConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id", "smsBlacklisted", "email"),
	})
	if err != nil {
		utils.Fail("error reading from Brevo", "error", err)
	}

	slog.Info("Reading projects..")
	utils.DumpJSON(res, os.Stdout)

	slog.Info("Reading projects..")
	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "attributes/deals",
		Fields:     connectors.Fields("attributeOptions", "isRequired", "label"),
	})
	if err != nil {
		utils.Fail("error reading from Brevo", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)

}
