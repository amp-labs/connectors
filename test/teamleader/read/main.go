package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/teamleader"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := teamleader.GetConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "departments",
		Fields:     connectors.Fields("id", "name", "currency"),
		Since:      time.Date(2025, 0o3, 0o1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		utils.Fail("error reading from Teamleader", "error", err)
	}

	slog.Info("Reading departments..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id", "first_name", "last_name", "emails"),
		Since:      time.Date(2025, 0o3, 0o1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		utils.Fail("error reading from Teamleader", "error", err)
	}

	slog.Info("Reading contacts..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "users",
		Fields:     connectors.Fields("id", "first_name", "last_name", "email"),
		Since:      time.Date(2025, 0o3, 0o1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		utils.Fail("error reading from Teamleader", "error", err)
	}

	slog.Info("Reading users..")
	utils.DumpJSON(res, os.Stdout)

	slog.Info("Read operation completed successfully.")
}
