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
	"github.com/amp-labs/connectors/test/fathom"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := fathom.GetFathomConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "meetings",
		Fields:     connectors.Fields("title", "meeting_title", "url", "transcript", "crm_matches"),
		Since:      time.Date(2025, 0o7, 0o3, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		utils.Fail("error reading from Fathom", "error", err)
	}

	slog.Info("Reading meetings..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "teams",
		Fields:     connectors.Fields("name", "created_at"),
		Since:      time.Date(2025, 0o6, 20, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		utils.Fail("error reading from Fathom", "error", err)
	}

	slog.Info("Reading teams..")
	utils.DumpJSON(res, os.Stdout)
}
