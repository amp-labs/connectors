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
	connTest "github.com/amp-labs/connectors/test/ashby"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetAshbyConnector(ctx)

	sinceTime := time.Date(2024, 12, 2, 6, 14, 0, 0, time.UTC)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "application",
		Fields:     connectors.Fields("id", "createdAt", "archivedAt", "source", "status"),
		Since:      sinceTime,
	})
	if err != nil {
		utils.Fail("error reading from Ashby", "error", err)
	}

	slog.Info("Reading application list..")
	utils.DumpJSON(res, os.Stdout)

	_, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "user",
		Fields:     connectors.Fields("id", "firstName", "lastName", "email"),
		Since:      sinceTime,
		NextPage:   res.NextPage,
	})
	if err != nil {
		utils.Fail("error reading from Ashby", "error", err)
	}

	slog.Info("Reading user list")
	// utils.DumpJSON(res, os.Stdout)
}
