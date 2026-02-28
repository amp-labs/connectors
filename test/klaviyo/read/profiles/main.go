package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/klaviyo"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetKlaviyoConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "profiles",
		Fields:     connectors.Fields("updated"),
		Since:      timestamp("2024-11-01T00:00:00.000Z"),
		Until:      timestamp("2024-12-01T00:00:00.000Z"),
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	fmt.Println("Reading...")
	utils.DumpJSON(res, os.Stdout)
}

func timestamp(timeText string) time.Time {
	result, err := time.Parse("2006-01-02T15:04:05.000Z", timeText)
	if err != nil {
		utils.Fail("bad timestamp", "error", err)
	}

	return result
}
