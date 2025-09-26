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
	connTest "github.com/amp-labs/connectors/test/gong"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGongConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "calls",
		Fields:     connectors.Fields("id"),
		Until:      timestamp("2023-03-01T05:43:33.157Z"),
	})
	if err != nil {
		utils.Fail("error reading from Gong", "error", err)
	}

	slog.Info("Reading calls..")
	utils.DumpJSON(res, os.Stdout)
}

func timestamp(timeText string) time.Time {
	result, err := time.Parse("2006-01-02T15:04:05.000Z", timeText)
	if err != nil {
		utils.Fail("bad timestamp", "error", err)
	}

	return result
}
