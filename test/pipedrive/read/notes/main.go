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
	connTest "github.com/amp-labs/connectors/test/pipedrive"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetPipedriveConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "notes",
		Since:      timestamp("2025-06-10T22:57:53.000Z"),
		// Until:  timestamp("2020-01-23T00:00:00.000Z"),
		Fields: connectors.Fields("content", "id"),
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
