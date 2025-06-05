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
	connTest "github.com/amp-labs/connectors/test/keap"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetKeapConnector(ctx)
	defer utils.Close(conn)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id", "experience", "last_updated"),
		Since:      timestamp("2025-05-28T18:30:05.000Z"),
		Until:      timestamp("2025-05-28T23:52:51.000Z"),
		// NextPage: "https://api.infusionsoft.com/crm/rest/v1/contacts/?limit=1&offset=50&since=2024-12-17T21:39:36.099Z&order=id",
	})
	if err != nil {
		utils.Fail("error reading from Keap", "error", err)
	}

	fmt.Println("Reading contacts..")
	utils.DumpJSON(res, os.Stdout)
}

func timestamp(timeText string) time.Time {
	result, err := time.Parse("2006-01-02T15:04:05.000Z", timeText)
	if err != nil {
		utils.Fail("bad timestamp", "error", err)
	}

	return result
}
