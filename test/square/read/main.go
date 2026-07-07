package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/square"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSquareConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "locations",
		Fields:     connectors.Fields("id", "name", "status"),
		PageSize:   2,
	})
	if err != nil {
		utils.Fail("error reading from square", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "customers",
		Fields:     connectors.Fields("id", "given_name", "family_name", "email_address"),
		PageSize:   2,
	})
	if err != nil {
		utils.Fail("error reading from square", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "payments",
		Fields:     connectors.Fields("id", "amount_money", "status", "created_at"),
		PageSize:   2,
	})
	if err != nil {
		utils.Fail("error reading from square", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)
	slog.Info("Read operation completed successfully.")
}
