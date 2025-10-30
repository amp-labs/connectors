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
	connTest "github.com/amp-labs/connectors/test/sellsy"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSellsyConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		// Incremental reading is supported. Response has "updated" field.
		ObjectName: "contacts",
		Fields: connectors.Fields("first_name", "last_name", "updated",
			// Custom fields:
			"hobbies", "age", "fruits"),
		// Since:      time.Now().Add(-1 * time.Minute * 12),
		// Until:      time.Now().Add(-1 * time.Minute * 9),
		// NextPage: "https://api.sellsy.com/v2/contacts/search?limit=36&offset=WyIzOCJd",
		Since: timestamp("2025-10-21T23:01:30+02:00"),
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	slog.Info("Reading...")
	utils.DumpJSON(res, os.Stdout)
}

func timestamp(timeText string) time.Time {
	result, err := time.Parse(time.RFC3339, timeText)
	if err != nil {
		utils.Fail("bad timestamp", "error", err)
	}

	return result
}
