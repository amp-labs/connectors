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
	"github.com/amp-labs/connectors/test/fourfour"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := fourfour.GetFourFourConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "Chats",
		Fields:     connectors.Fields("id", "created_at"),
		Since:      time.Date(2026, 03, 05, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		utils.Fail("error reading Chats from FourFour", "error", err)
	}

	slog.Info("Reading Chats..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "ChatMessages",
		Fields:     connectors.Fields("id", "created_at"),
		Since:      time.Date(2026, 03, 05, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		utils.Fail("error reading ChatMessages from FourFour", "error", err)
	}

	slog.Info("Reading ChatMessages..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "Labels",
		Fields:     connectors.Fields("id", "updated"),
		Since:      time.Date(2026, 03, 05, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		utils.Fail("error reading Labels from FourFour", "error", err)
	}

	slog.Info("Reading Labels..")
	utils.DumpJSON(res, os.Stdout)

}
