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
	"github.com/amp-labs/connectors/test/acuityscheduling"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := acuityscheduling.GetAcuitySchedulingConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "appointments",
		Fields:     connectors.Fields("id", "firstName", "lastName"),
		Since:      time.Date(2025, 12, 0, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		utils.Fail("error reading from acuityscheduling", "error", err)
	}

	slog.Info("Reading appointments..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "clients",
		Fields:     connectors.Fields("id", "firstName", "lastName", "email"),
	})
	if err != nil {
		utils.Fail("error reading from acuityscheduling", "error", err)
	}

	slog.Info("Reading clients..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "calendars",
		Fields:     connectors.Fields("id", "name", "replyTo", "isValid"),
	})
	if err != nil {
		utils.Fail("error reading from acuityscheduling", "error", err)
	}

	slog.Info("Reading calendars..")
	utils.DumpJSON(res, os.Stdout)

	slog.Info("Read operation completed successfully.")
}
