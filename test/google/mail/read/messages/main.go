package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGoogleMailConnector(ctx)

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "messages",
		Fields: connectors.Fields(
			"snippet",
			"$['payload']['body']",
			"$['payload']['mimeType']",
		),
		Since:    timestamp("2026-01-25T00:00:00"),
		Until:    timestamp("2026-01-28T00:00:00"),
		PageSize: 10,
	})
}

func timestamp(timeText string) time.Time {
	result, err := time.Parse("2006-01-02T15:04:05", timeText)
	if err != nil {
		utils.Fail("bad timestamp", "error", err)
	}

	return result
}
