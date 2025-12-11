package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/outplay"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := outplay.GetOutplayConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "prospect",
		Fields:     connectors.Fields("city", "Phones", "firstname", "lastname"),
	})
	if err != nil {
		utils.Fail("error reading from outplay", "error", err)
	}

	slog.Info("Reading prospects..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "prospectaccount",
		Fields:     connectors.Fields("name", "accountid", "description"),
	})
	if err != nil {
		utils.Fail("error reading from outplay", "error", err)
	}

	slog.Info("Reading prospect accounts..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "sequence",
		Fields:     connectors.Fields("name", "description", "isactive"),
	})
	if err != nil {
		utils.Fail("error reading from outplay", "error", err)
	}

	slog.Info("Reading sequences..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "task",
		Fields:     connectors.Fields("subject", "dueat", "status"),
	})
	if err != nil {
		utils.Fail("error reading from outplay", "error", err)
	}

	slog.Info("Reading tasks..")
	utils.DumpJSON(res, os.Stdout)

	slog.Info("Read operation completed successfully.")
}
