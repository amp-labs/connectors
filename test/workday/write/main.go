package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/workday"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetWorkdayConnector(ctx)

	result, err := conn.Write(ctx, connectors.WriteParams{
		ObjectName: "workers",
		RecordData: map[string]any{
			"descriptor":       "Test Worker",
			"primaryWorkEmail": "test@workday.net",
		},
	})
	if err != nil {
		utils.Fail("error writing to workday", "error", err)
	}

	fmt.Println("workday write result...")
	utils.DumpJSON(result, os.Stdout)
}
