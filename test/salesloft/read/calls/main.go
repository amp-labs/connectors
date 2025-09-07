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
	msTest "github.com/amp-labs/connectors/test/salesloft"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := msTest.GetSalesloftConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "activities/calls",
		Since:      timestamp("2025-03-17T20:00:22.806808-04:00"),
		Until:      timestamp("2025-03-17T21:14:43.917967-04:00"),
		Fields:     connectors.Fields("updated_at"),
	})
	if err != nil {
		utils.Fail("error reading from Salesloft", "error", err)
	}

	fmt.Println("Reading people..")
	utils.DumpJSON(res, os.Stdout)
}

func timestamp(timeText string) time.Time {
	result, err := time.Parse("2006-01-02T15:04:05.000000-07:00", timeText)
	if err != nil {
		utils.Fail("bad timestamp", "error", err)
	}

	return result
}
