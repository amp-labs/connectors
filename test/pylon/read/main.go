package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/pylon"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := pylon.GetConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "accounts",
		Fields:     connectors.Fields("id", "channels", "name"),
	})
	if err != nil {
		utils.Fail("error reading from pylon", "error", err)
	}

	slog.Info("Reading accounts..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id", "email", "avatar_url"),
	})
	if err != nil {
		utils.Fail("error reading from pylon", "error", err)
	}

	slog.Info("Reading contacts..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "issues",
		Fields:     connectors.Fields("id", "created_at", "body_html"),
	})
	if err != nil {
		utils.Fail("error reading from pylon", "error", err)
	}

	slog.Info("Reading issues..")
	utils.DumpJSON(res, os.Stdout)

	slog.Info("Read operation completed successfully.")
}
